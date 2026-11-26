package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/redis/go-redis/v9"

	"webgx/internal/db"
)

// usedGrace is how long a refresh token's jti stays redeemable after it has
// already been rotated once. Two tabs (or a poller and a page load) may race on
// the same cookie: the winner rotates it, the loser must not be signed out for
// that, so the loser's replay is accepted inside this window. The value is
// deliberately short and is cleared by every revocation path.
const usedGrace = 30 * time.Second

// tauriOrigins are the webview origins of the Tauri shells. They are cross-site
// with respect to the API host, so the refresh cookie must be SameSite=None
// when the request comes from one of them.
var tauriOrigins = []string{
	"tauri://localhost",
	"http://tauri.localhost",
	"https://tauri.localhost",
}

// CookieName is the refresh-token cookie name for this audience.
func (a *Auth) CookieName() string {
	return a.Prefix + "_refresh"
}

// CookiePath scopes the refresh cookie to the auth endpoints of this audience,
// so it is never sent to ordinary API calls.
func (a *Auth) CookiePath() string {
	return "/" + a.Prefix + "/api/v1/auth"
}

func (a *Auth) redisUsedKey(jti string) string {
	return a.Prefix + ":refresh:used:" + jti
}

// SetRefreshCookie stores the refresh token in an HttpOnly cookie.
func (a *Auth) SetRefreshCookie(c fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     a.CookieName(),
		Value:    token,
		Path:     a.CookiePath(),
		MaxAge:   int(a.RefreshTTL.Seconds()),
		Expires:  time.Now().Add(a.RefreshTTL),
		HTTPOnly: true,
		Secure:   cookieSecure(c),
		SameSite: cookieSameSite(c),
	})
}

// ClearRefreshCookie expires the refresh cookie. The attributes must match
// SetRefreshCookie or browsers keep the original cookie.
func (a *Auth) ClearRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     a.CookieName(),
		Value:    "",
		Path:     a.CookiePath(),
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HTTPOnly: true,
		Secure:   cookieSecure(c),
		SameSite: cookieSameSite(c),
	})
}

// RefreshFromCookie returns the refresh token sent by the client, or "".
func (a *Auth) RefreshFromCookie(c fiber.Ctx) string {
	return c.Cookies(a.CookieName())
}

// cookieSecure reports whether the client reached us over HTTPS. The nginx and
// caddy templates set X-Forwarded-Proto, and the app itself runs behind them.
func cookieSecure(c fiber.Ctx) bool {
	return c.Secure() || strings.EqualFold(c.Get("X-Forwarded-Proto"), "https")
}

// cookieSameSite returns None for the cross-site Tauri webview (which requires
// Secure as well; without it browsers drop the cookie, so Lax is the fallback),
// and Lax for the ordinary same-origin web deployment.
func cookieSameSite(c fiber.Ctx) string {
	origin := c.Get(fiber.HeaderOrigin)
	for _, o := range tauriOrigins {
		if strings.EqualFold(origin, o) && cookieSecure(c) {
			return fiber.CookieSameSiteNoneMode
		}
	}
	return fiber.CookieSameSiteLaxMode
}

// Rotate redeems a refresh token exactly once and returns its user ID.
//
// A replay of a jti that was rotated within usedGrace is still accepted: two
// tabs sharing one cookie would otherwise sign each other out. Anything older
// than the grace window, or never issued, is rejected, so a stolen token that
// the legitimate client already refreshed cannot be redeemed later.
func (a *Auth) Rotate(token string) (string, error) {
	claims, err := fext.ParseToken(token, a.SecretKey)
	if err != nil {
		return "", err
	}
	if claims.ID == "" || claims.Subject == "" {
		return "", errors.New("invalid refresh token")
	}
	if db.Rx == nil {
		return "", errors.New("redis unavailable")
	}
	val, getErr := db.Rx.Get(db.Rc, a.redisKey(claims.ID)).Result()
	switch {
	case getErr == nil:
		if val != claims.Subject {
			return "", errors.New("refresh token mismatch")
		}
		pipe := db.Rx.TxPipeline()
		pipe.Del(db.Rc, a.redisKey(claims.ID))
		pipe.Set(db.Rc, a.redisUsedKey(claims.ID), claims.Subject, usedGrace)
		pipe.SRem(db.Rc, a.redisUserKey(claims.Subject), claims.ID)
		if _, execErr := pipe.Exec(db.Rc); execErr != nil {
			return "", execErr
		}
		return claims.Subject, nil
	case errors.Is(getErr, redis.Nil):
		// Already rotated moments ago by a concurrent request.
		_, usedErr := db.Rx.Get(db.Rc, a.redisUsedKey(claims.ID)).Result()
		if usedErr == nil {
			return claims.Subject, nil
		}
		if !errors.Is(usedErr, redis.Nil) {
			return "", usedErr
		}
		return "", errors.New("refresh token revoked")
	default:
		return "", getErr
	}
}
