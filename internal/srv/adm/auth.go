package adm

import (
	"webtplmst/internal/db"
	"webtplmst/internal/srv/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/strs"
)

type Auth struct {
	ID    uint   `json:"id"`
	Token string `json:"token"`
} //	@name	RAuth

type AuthIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=4,max=20"`
} //	@name	RAuthIn

// SignIn godoc
//
//	@Summary		Admin sign in
//	@Description	Sign in with username and password
//	@Tags			Auth
//	@ID				admin__auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		AuthIn	true	"Credentials"
//	@Success		200		{object}	RAuth
//	@Failure		400		{object}	Fail
//	@Router			/adm/api/v1/auth/in [post]
func SignIn(c fiber.Ctx) error {
	d, err := fext.BodyVarser[AuthIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	v, err := db.AuthAdmin(db.Tx, d.Username, d.Password)
	if err != nil {
		return &fext.Fail{Code: internal.SignInFailed}
	}
	token, err := jwt.GenToken(strs.FormatUint(v.ID))
	if err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
	return c.JSON(Auth{v.ID, token})
}
