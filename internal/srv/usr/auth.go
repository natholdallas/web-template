package usr

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/natholdallas/natools4go/rands"
	"github.com/natholdallas/natools4go/strs"
	"gorm.io/gorm"

	"webgx/internal/auth"
	"webgx/internal/db"
	"webgx/internal/oauth"
	"webgx/internal/srv/internal"
)

type Auth struct {
	ID          uint   `json:"id"`
	AccessToken string `json:"accessToken"`
} //	@name	UsrAuth

type AuthIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=4,max=72"`
} //	@name	UsrAuthIn

// Login godoc
//
//	@Summary		User sign in
//	@Description	Sign in with username and password, returns access and refresh tokens
//	@Tags			usrAuth
//	@ID				usrSignIn
//	@Accept			json
//	@Produce		json
//	@Param			body	body		AuthIn	true	"Credentials"
//	@Success		200		{object}	Auth
//	@Failure		400		{object}	Fail
//	@Router			/usr/api/v1/auth/login [post]
func Login(c fiber.Ctx) error {
	d, err := fext.BodyVarser[AuthIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	v, err := db.AuthUser(db.Tx, d.Username, d.Password)
	if err != nil {
		// No Status on purpose: 401 would make the SDK treat this as an expired
		// session and silently try /auth/refresh instead of showing the error.
		return &fext.Fail{Code: internal.SignInFailed, Message: "invalid username or password"}
	}
	p, err := auth.Usr.GenPair(strs.FormatUint(v.ID))
	if err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
	auth.Usr.SetRefreshCookie(c, p.RefreshToken)
	return c.JSON(Auth{v.ID, p.AccessToken})
}

// Refresh godoc
//
//	@Summary		Rotate user tokens
//	@Description	Exchange the refresh-token cookie for a new access token and a
//	@Description	rotated refresh cookie. The request has no body.
//	@Tags			usrAuth
//	@ID				usrRefresh
//	@Produce		json
//	@Success		200	{object}	Auth
//	@Failure		401	{object}	Fail
//	@Router			/usr/api/v1/auth/refresh [post]
func Refresh(c fiber.Ctx) error {
	token := auth.Usr.RefreshFromCookie(c)
	if token == "" {
		return &fext.Fail{Code: internal.SignInFailed, Status: fiber.StatusUnauthorized, Message: "missing refresh token"}
	}
	userID, err := auth.Usr.Rotate(token)
	if err != nil {
		return &fext.Fail{Code: internal.SignInFailed, Status: fiber.StatusUnauthorized, Message: "invalid refresh token"}
	}
	p, err := auth.Usr.GenPair(userID)
	if err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
	auth.Usr.SetRefreshCookie(c, p.RefreshToken)
	return c.JSON(Auth{strs.ParseUint[uint](userID), p.AccessToken})
}

// Logout godoc
//
//	@Summary		Revoke user refresh token
//	@Description	Revoke the refresh-token cookie and clear it. Idempotent: a request
//	@Description	without a cookie still returns 200.
//	@Tags			usrAuth
//	@ID				usrLogout
//	@Produce		json
//	@Success		200
//	@Router			/usr/api/v1/auth/logout [post]
func Logout(c fiber.Ctx) error {
	if token := auth.Usr.RefreshFromCookie(c); token != "" {
		if err := auth.Usr.RevokeRefresh(token); err != nil {
			log.Warn("revoke refresh token failed: ", err)
		}
	}
	auth.Usr.ClearRefreshCookie(c)
	return nil
}

type GoogleAuthIn struct {
	Code string `json:"code" validate:"required"`
} //	@name	UsrGoogleAuthIn

// SignInGoogle godoc
//
//	@Summary		Sign in with Google
//	@Description	Exchange a Google OAuth code for an access/refresh token pair.
//	@Description	Creates an account automatically when no user is bound to the Google email.
//	@Tags			usrAuth
//	@ID				usrSignInGoogle
//	@Accept			json
//	@Produce		json
//	@Param			body	body		GoogleAuthIn	true	"Google OAuth code"
//	@Success		200		{object}	Auth
//	@Failure		400		{object}	Fail
//	@Router			/usr/api/v1/auth/google [post]
func SignInGoogle(c fiber.Ctx) error {
	d, err := fext.BodyVarser[GoogleAuthIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	user, err := oauth.GoogleExchange(d.Code)
	if err != nil {
		return &fext.Fail{Code: internal.OauthFailed, Status: fiber.StatusBadRequest, System: err}
	}
	if user.Email == "" {
		return &fext.Fail{Code: internal.OauthFailed, Message: "google email missing"}
	}
	if !user.VerifiedEmail {
		return &fext.Fail{Code: internal.OauthFailed, Message: "google email not verified"}
	}
	v, err := db.FindUserByGoogle(db.Tx, user.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		v, err = createGoogleUser(user.Email, user.Name, user.Picture)
		if err != nil {
			return &fext.Fail{Code: internal.CreateFailed, System: err}
		}
	} else if err != nil {
		return &fext.Fail{Code: internal.OperationFailed, System: err}
	}
	p, err := auth.Usr.GenPair(strs.FormatUint(v.ID))
	if err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
	auth.Usr.SetRefreshCookie(c, p.RefreshToken)
	return c.JSON(Auth{v.ID, p.AccessToken})
}

// SignInTelegram godoc
//
//	@Summary		Sign in with Telegram
//	@Description	Verify the Telegram Login Widget payload and return an access/refresh token pair.
//	@Description	Creates an account automatically when no user is bound to the Telegram ID.
//	@Tags			usrAuth
//	@ID				usrSignInTelegram
//	@Accept			json
//	@Produce		json
//	@Param			body	body		oauth.TelegramAuth	true	"Signed Telegram widget payload"
//	@Success		200		{object}	Auth
//	@Failure		400		{object}	Fail
//	@Router			/usr/api/v1/auth/telegram [post]
func SignInTelegram(c fiber.Ctx) error {
	d, err := fext.BodyVarser[oauth.TelegramAuth](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if err = oauth.CheckTelegramAuth(&d); err != nil {
		return &fext.Fail{Code: internal.OauthFailed, Status: fiber.StatusBadRequest, System: err}
	}
	v, err := db.FindUserByTelegram(db.Tx, d.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		v, err = createTelegramUser(&d)
		if err != nil {
			return &fext.Fail{Code: internal.CreateFailed, System: err}
		}
	} else if err != nil {
		return &fext.Fail{Code: internal.OperationFailed, System: err}
	}
	p, err := auth.Usr.GenPair(strs.FormatUint(v.ID))
	if err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
	auth.Usr.SetRefreshCookie(c, p.RefreshToken)
	return c.JSON(Auth{v.ID, p.AccessToken})
}

// oauthUsername derives a unique username (max 50 chars) from a base string.
func oauthUsername(base string) string {
	base = strings.TrimSpace(base)
	if i := strings.Index(base, "@"); i > 0 {
		base = base[:i]
	}
	var b strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	u := b.String()
	if u == "" {
		u = "user"
	}
	if len(u) > 42 {
		u = u[:42]
	}
	return u + "_" + rands.Char(4)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func randomPassword() string {
	return pwd.MustHash(rands.Char(20))
}

func createGoogleUser(email, name, picture string) (db.User, error) {
	u := db.User{
		Username: oauthUsername(email),
		Password: randomPassword(),
		GoogleInfo: db.GoogleInfo{
			Email:  strPtr(email),
			Name:   strPtr(name),
			Avatar: strPtr(picture),
		},
	}
	err := orms.Create(db.Tx, &u)
	return u, err
}

func createTelegramUser(d *oauth.TelegramAuth) (db.User, error) {
	name := strings.TrimSpace(d.FirstName + " " + d.LastName)
	base := d.Username
	if base == "" {
		base = "telegram_" + d.ID
	}
	u := db.User{
		Username: oauthUsername(base),
		Password: randomPassword(),
		TelegramInfo: db.TelegramInfo{
			ID:     strPtr(d.ID),
			Name:   strPtr(name),
			Avatar: strPtr(d.PhotoURL),
		},
	}
	err := orms.Create(db.Tx, &u)
	return u, err
}
