package usr

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/natholdallas/natools4go/rands"
	"github.com/natholdallas/natools4go/strs"
	"gorm.io/gorm"

	"webtplmst/internal/auth"
	"webtplmst/internal/conf"
	"webtplmst/internal/db"
	"webtplmst/internal/oauth"
	"webtplmst/internal/srv/internal"
)

type Auth struct {
	ID           uint   `json:"id"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
} //	@name	UsrAuth

type AuthIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=4,max=20"`
} //	@name	UsrAuthIn

type RefreshIn struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
} //	@name	UsrRefreshIn

var authMgr = auth.New(conf.App.SecretUsr, "usr", time.Duration(conf.App.AccessMinutes)*time.Minute, time.Duration(conf.App.RefreshHours)*time.Hour)

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
		return &fext.Fail{Code: internal.SignInFailed}
	}
	p, err := authMgr.GenPair(strs.FormatUint(v.ID))
	if err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
	return c.JSON(Auth{v.ID, p.AccessToken, p.RefreshToken})
}

// Refresh godoc
//
//	@Summary		Rotate user tokens
//	@Description	Exchange a valid refresh token for a new access/refresh pair
//	@Tags			usrAuth
//	@ID				usrRefresh
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RefreshIn	true	"Refresh token"
//	@Success		200		{object}	Auth
//	@Failure		400		{object}	Fail
//	@Router			/usr/api/v1/auth/refresh [post]
func Refresh(c fiber.Ctx) error {
	d, err := fext.BodyVarser[RefreshIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	userID, err := authMgr.VerifyRefresh(d.RefreshToken)
	if err != nil {
		return &fext.Fail{Code: internal.SignInFailed, Message: "invalid refresh token"}
	}
	if err := authMgr.RevokeRefresh(d.RefreshToken); err != nil {
		log.Warn("revoke refresh token failed: ", err)
	}
	p, err := authMgr.GenPair(userID)
	if err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
	return c.JSON(Auth{strs.ParseUint[uint](userID), p.AccessToken, p.RefreshToken})
}

// Logout godoc
//
//	@Summary		Revoke user refresh token
//	@Description	Revoke the given refresh token so it can no longer be used
//	@Tags			usrAuth
//	@ID				usrLogout
//	@Accept			json
//	@Produce		json
//	@Param			body	body	RefreshIn	true	"Refresh token"
//	@Success		200
//	@Failure		400	{object}	Fail
//	@Router			/usr/api/v1/auth/logout [post]
func Logout(c fiber.Ctx) error {
	d, err := fext.BodyVarser[RefreshIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if err := authMgr.RevokeRefresh(d.RefreshToken); err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
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
		return &fext.Fail{Code: internal.OauthFailed, System: err}
	}
	if user.Email == "" {
		return &fext.Fail{Code: internal.OauthFailed, Message: "google email missing"}
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
	p, err := authMgr.GenPair(strs.FormatUint(v.ID))
	if err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
	return c.JSON(Auth{v.ID, p.AccessToken, p.RefreshToken})
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
	if err := oauth.CheckTelegramAuth(&d); err != nil {
		return &fext.Fail{Code: internal.OauthFailed, System: err}
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
	p, err := authMgr.GenPair(strs.FormatUint(v.ID))
	if err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
	return c.JSON(Auth{v.ID, p.AccessToken, p.RefreshToken})
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
	return pwd.TryHash(rands.Char(20))
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
