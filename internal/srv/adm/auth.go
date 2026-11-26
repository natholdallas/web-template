package adm

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/strs"

	"webgx/internal/auth"
	"webgx/internal/db"
	"webgx/internal/srv/internal"
)

type Auth struct {
	ID          uint   `json:"id"`
	AccessToken string `json:"accessToken"`
} //	@name	AdmAuth

type AuthIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=4,max=72"`
} //	@name	AdmAuthIn

// Login godoc
//
//	@Summary		Admin sign in
//	@Description	Sign in with username and password. The access token is returned in the
//	@Description	body, the refresh token is set as an HttpOnly cookie scoped to
//	@Description	/adm/api/v1/auth.
//	@Tags			admAuth
//	@ID				admSignIn
//	@Accept			json
//	@Produce		json
//	@Param			body	body		AuthIn	true	"Credentials"
//	@Success		200		{object}	Auth
//	@Failure		400		{object}	Fail
//	@Router			/adm/api/v1/auth/login [post]
func Login(c fiber.Ctx) error {
	d, err := fext.BodyVarser[AuthIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	v, err := db.AuthAdmin(db.Tx, d.Username, d.Password)
	if err != nil {
		// No Status on purpose: 401 would make the SDK treat this as an expired
		// session and silently try /auth/refresh instead of showing the error.
		return &fext.Fail{Code: internal.SignInFailed, Message: "invalid username or password"}
	}
	p, err := auth.Adm.GenPair(strs.FormatUint(v.ID))
	if err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
	auth.Adm.SetRefreshCookie(c, p.RefreshToken)
	return c.JSON(Auth{v.ID, p.AccessToken})
}

// Refresh godoc
//
//	@Summary		Rotate admin tokens
//	@Description	Exchange the refresh-token cookie for a new access token and a rotated
//	@Description	refresh cookie. The request has no body.
//	@Tags			admAuth
//	@ID				admRefresh
//	@Produce		json
//	@Success		200	{object}	Auth
//	@Failure		401	{object}	Fail
//	@Router			/adm/api/v1/auth/refresh [post]
func Refresh(c fiber.Ctx) error {
	token := auth.Adm.RefreshFromCookie(c)
	if token == "" {
		return &fext.Fail{Code: internal.SignInFailed, Status: fiber.StatusUnauthorized, Message: "missing refresh token"}
	}
	userID, err := auth.Adm.Rotate(token)
	if err != nil {
		return &fext.Fail{Code: internal.SignInFailed, Status: fiber.StatusUnauthorized, Message: "invalid refresh token"}
	}
	p, err := auth.Adm.GenPair(userID)
	if err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
	auth.Adm.SetRefreshCookie(c, p.RefreshToken)
	return c.JSON(Auth{strs.ParseUint[uint](userID), p.AccessToken})
}

// Logout godoc
//
//	@Summary		Revoke admin refresh token
//	@Description	Revoke the refresh-token cookie and clear it. Idempotent: a request
//	@Description	without a cookie still returns 200.
//	@Tags			admAuth
//	@ID				admLogout
//	@Produce		json
//	@Success		200
//	@Router			/adm/api/v1/auth/logout [post]
func Logout(c fiber.Ctx) error {
	if token := auth.Adm.RefreshFromCookie(c); token != "" {
		if err := auth.Adm.RevokeRefresh(token); err != nil {
			log.Warn("revoke refresh token failed: ", err)
		}
	}
	auth.Adm.ClearRefreshCookie(c)
	return nil
}
