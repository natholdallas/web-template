package usr

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/strs"

	"webtplmst/internal/auth"
	"webtplmst/internal/conf"
	"webtplmst/internal/db"
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
