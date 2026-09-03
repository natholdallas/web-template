package adm

import (
	"webtplmst/internal/db"
	"webtplmst/internal/srv/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
)

// FindProfile godoc
//
//	@Summary	Current admin profile
//	@Tags		admProfile
//	@ID			admFindProfile
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Success	200	{object}	db.Admin
//	@Router		/adm/api/v1/profile/me [get]
func FindProfile(c fiber.Ctx) error {
	claims := jwt.Claims(c)
	v := orms.IFirst[db.Admin](db.Tx, claims.ID)
	return c.JSON(v)
}

type ProfileIn struct {
	Username string `json:"username" validate:"required"`
} //	@name	AdmProfileIn

// UpdateProfile godoc
//
//	@Summary	Update current admin profile
//	@Tags		admProfile
//	@ID			admUpdateProfile
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		body	body	ProfileIn	true	"Profile object"
//	@Success	200
//	@Router		/adm/api/v1/profile/me [put]
func UpdateProfile(c fiber.Ctx) error {
	d, err := fext.BodyVarser[ProfileIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	claims := jwt.Claims(c)
	if err := orms.UpdatesByID[db.Admin](db.Tx, claims.ID, map[string]any{
		"username": d.Username,
	}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, Message: err.Error()}
	}
	return nil
}

type ResetPasswordIn struct {
	Old string `json:"old" validate:"required,min=4,max=50"`
	New string `json:"new" validate:"required,min=4,max=50"`
} //	@name	AdmResetPasswordIn

// ResetPassword godoc
//
//	@Summary	Reset current admin password
//	@Tags		admProfile
//	@ID			admResetPassword
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		body	body	ResetPasswordIn	true	"Password object"
//	@Success	200
//	@Failure	400	{object}	Fail
//	@Router		/adm/api/v1/profile/password [put]
func ResetPassword(c fiber.Ctx) error {
	d, err := fext.BodyVarser[ResetPasswordIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	claims := jwt.Claims(c)
	v, err := orms.First[db.Admin](db.Tx, "id = ?", claims.ID)
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
	if err := orms.UpdatesByID[db.Admin](db.Tx, claims.ID, map[string]any{"password": hash}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, Message: err.Error()}
	}
	return nil
}
