package user

import (
	"webtplmst/internal/db"
	"webtplmst/internal/srv/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
)

type User struct {
	orms.Model[uint]
	Username string `json:"username"`
} //	@name	UUser

// FindUser godoc
//
//	@Summary	Find User by ID
//	@Tags		User
//	@ID			user__find_user
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	User
//	@Router		/user/api/v1/user [get]
func FindUser(c fiber.Ctx) error {
	claims := jwt.Claims(c)
	v := orms.QM[User, db.User](db.Tx).IFirst(claims.ID)
	return c.JSON(v)
}

type UserIn struct {
	Username string `json:"username" validate:"required,min=4"`
} //	@name	UserIn

func (s *UserIn) Get() *db.User {
	return &db.User{
		Username: s.Username,
	}
}

// UpdateUser godoc
//
//	@Summary	Update User
//	@Tags		User
//	@ID			user__update_user
//	@Accept		json
//	@Produce	json
//	@Param		body	body	UserIn	true	"User object"
//	@Success	200
//	@Router		/user/api/v1/user [put]
func UpdateUser(c fiber.Ctx) error {
	d, err := fext.BodyVarser[UserIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	claims := jwt.Claims(c)
	v := d.Get()
	orms.UpdatesByID[db.User](db.Tx, claims.ID, &v)
	return nil
}

type ResetPasswordIn struct {
	Old string `json:"old" validate:"required,min=4,max=50"`
	New string `json:"new" validate:"required,min=4,max=50"`
} //	@name	UResetPasswordIn

// ResetPassword godoc
//
//	@Summary		Reset user password
//	@Description	Reset current user's password
//	@Tags			User
//	@ID				user__reset_password
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		400	{object}	Fail
//	@Router			/user/api/v1/user/reset/password [post]
func ResetPassword(c fiber.Ctx) error {
	d, err := fext.BodyVarser[ResetPasswordIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	claims := jwt.Claims(c)
	orms.QE[db.User](db.Tx).
		Where("id = ? and password = ?", claims.ID, d.Old).
		Update("password", d.New)
	return nil
}
