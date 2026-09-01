package usr

import (
	"webtplmst/internal/db"
	"webtplmst/internal/srv/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
)

type Profile struct {
	orms.Model[uint]
	Username string `json:"username"`
} //	@name	UsrProfile

// FindProfile godoc
//
//	@Summary	Current user profile
//	@Tags		usrProfile
//	@ID			usrFindProfile
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Success	200	{object}	Profile
//	@Router		/usr/api/v1/profile/me [get]
func FindProfile(c fiber.Ctx) error {
	claims := jwt.Claims(c)
	v := orms.QM[Profile, db.User](db.Tx).IFirst(claims.ID)
	return c.JSON(v)
}

type ProfileIn struct {
	Username string `json:"username" validate:"required,min=4"`
} //	@name	UsrProfileIn

func (s *ProfileIn) Get() *db.User {
	return &db.User{
		Username: s.Username,
	}
}

// UpdateProfile godoc
//
//	@Summary	Update current user profile
//	@Tags		usrProfile
//	@ID			usrUpdateProfile
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		body	body	ProfileIn	true	"Profile object"
//	@Success	200
//	@Failure	400	{object}	Fail
//	@Router		/usr/api/v1/profile/me [put]
func UpdateProfile(c fiber.Ctx) error {
	d, err := fext.BodyVarser[ProfileIn](c)
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
} //	@name	UsrResetPasswordIn

// ResetPassword godoc
//
//	@Summary		Reset current user password
//	@Description	Reset current user's password
//	@Tags			usrProfile
//	@ID				usrResetPassword
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			body	body	ResetPasswordIn	true	"Password object"
//	@Success		200
//	@Failure		400	{object}	Fail
//	@Router			/usr/api/v1/profile/password [put]
func ResetPassword(c fiber.Ctx) error {
	d, err := fext.BodyVarser[ResetPasswordIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	claims := jwt.Claims(c)
	v, err := orms.First[db.User](db.Tx, "id = ?", claims.ID)
	if err != nil {
		return &fext.Fail{Code: internal.OperationFailed, Message: "incorrect old password"}
	}
	if !pwd.Verify(d.Old, v.Password) && (pwd.IsHashed(v.Password) || v.Password != d.Old) {
		return &fext.Fail{Code: internal.OperationFailed, Message: "incorrect old password"}
	}
	hash, err := pwd.Hash(d.New)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	orms.UpdatesByID[db.User](db.Tx, claims.ID, map[string]any{"password": hash})
	return nil
}
