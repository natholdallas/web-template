// Package auth provides access/refresh JWT pairs backed by Redis revocation.
//
// Access tokens are deliberately not revoked: they are short lived
// (conf.App.AccessMinutes) and Redis only tracks refresh-token jtis, so a
// password change or sign-out kills future refreshes and the current access
// token expires on its own.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/rands"

	"webgx/internal/conf"
	"webgx/internal/db"
)

// Pair is an access token and its matching refresh token.
type Pair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Auth manages token issuance and refresh-token revocation for one audience.
type Auth struct {
	SecretKey  string
	Prefix     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// New builds an Auth manager. Prefix (e.g. "adm"/"usr") namespaces the Redis keys.
func New(secretKey, prefix string, accessTTL, refreshTTL time.Duration) Auth {
	return Auth{secretKey, prefix, accessTTL, refreshTTL}
}

// Usr and Adm are the process-wide token managers. They are built once from
// the loaded configuration by Setup; the configuration is never reloaded at
// runtime.
var (
	Usr Auth
	Adm Auth
)

// Setup builds the process-wide token managers after the configuration is loaded.
func Setup() {
	Usr = New(conf.App.SecretUsr, "usr",
		time.Duration(conf.App.AccessMinutes)*time.Minute,
		time.Duration(conf.App.RefreshHours)*time.Hour)
	Adm = New(conf.App.SecretAdm, "adm",
		time.Duration(conf.App.AccessMinutes)*time.Minute,
		time.Duration(conf.App.RefreshHours)*time.Hour)
}

func (a *Auth) redisKey(jti string) string {
	return fmt.Sprintf("%s:refresh:%s", a.Prefix, jti)
}

// redisUserKey is the set of live refresh token IDs issued to one user, so a
// password change can invalidate every session at once.
func (a *Auth) redisUserKey(userID string) string {
	return fmt.Sprintf("%s:refresh:user:%s", a.Prefix, userID)
}

// GenPair issues an access token and a refresh token for the given user ID.
// The refresh token is recorded in Redis so it can be revoked or rotated.
func (a *Auth) GenPair(userID string) (Pair, error) {
	access, err := fext.GenToken(userID, a.SecretKey, a.AccessTTL)
	if err != nil {
		return Pair{}, err
	}
	// The refresh token needs the jti as its revocation key while the subject
	// stays the user, so it cannot use GenToken (which sets ID == Subject).
	jti := rands.Char(32)
	refresh, err := fext.GenTokenWithID(userID, jti, a.SecretKey, a.RefreshTTL)
	if err != nil {
		return Pair{}, err
	}
	if db.Rx == nil {
		return Pair{}, errors.New("redis unavailable")
	}
	// A refresh token that is not recorded in Redis can never be redeemed:
	// Rotate looks the jti up and treats a miss as revoked. Failing here is
	// better than handing the client a token that is dead on arrival.
	userKey := a.redisUserKey(userID)
	pipe := db.Rx.TxPipeline()
	pipe.Set(db.Rc, a.redisKey(jti), userID, a.RefreshTTL)
	pipe.SAdd(db.Rc, userKey, jti)
	pipe.Expire(db.Rc, userKey, a.RefreshTTL)
	if _, err := pipe.Exec(db.Rc); err != nil {
		return Pair{}, fmt.Errorf("store refresh token: %w", err)
	}
	return Pair{access, refresh}, nil
}

// RevokeRefresh invalidates the given refresh token in Redis.
func (a *Auth) RevokeRefresh(token string) error {
	claims, err := fext.ParseToken(token, a.SecretKey)
	if err != nil {
		return err
	}
	if claims.ID == "" {
		return errors.New("invalid refresh token")
	}
	if db.Rx == nil {
		return errors.New("redis unavailable")
	}
	pipe := db.Rx.TxPipeline()
	pipe.Del(db.Rc, a.redisKey(claims.ID))
	pipe.Del(db.Rc, a.redisUsedKey(claims.ID))
	if claims.Subject != "" {
		pipe.SRem(db.Rc, a.redisUserKey(claims.Subject), claims.ID)
	}
	_, err = pipe.Exec(db.Rc)
	return err
}

// RevokeAll invalidates every refresh token issued to the given user.
func (a *Auth) RevokeAll(userID string) error {
	if userID == "" {
		return errors.New("invalid user id")
	}
	if db.Rx == nil {
		return errors.New("redis unavailable")
	}
	userKey := a.redisUserKey(userID)
	jtis, err := db.Rx.SMembers(db.Rc, userKey).Result()
	if err != nil {
		return err
	}
	pipe := db.Rx.TxPipeline()
	for _, jti := range jtis {
		pipe.Del(db.Rc, a.redisKey(jti))
		// Drop the rotation grace marker too: a logged-out or
		// password-reset session must not be resurrectable by the
		// concurrent-rotation window.
		pipe.Del(db.Rc, a.redisUsedKey(jti))
	}
	pipe.Del(db.Rc, userKey)
	_, err = pipe.Exec(db.Rc)
	return err
}
