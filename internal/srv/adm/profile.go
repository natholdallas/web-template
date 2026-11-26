package adm

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/natholdallas/natools4go/strs"

	"webgx/internal/auth"
	"webgx/internal/db"
	"webgx/internal/srv/internal"
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
	v, err := orms.First[db.Admin](db.Tx, claims.ID)
	if err != nil {
		return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
	}
	return c.JSON(v)
}

type ProfileIn struct {
	Username string `json:"username" validate:"required,max=50"`
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
	if orms.Exists[db.Admin](db.Tx.Where("username = ? AND id <> ?", d.Username, claims.ID)) {
		return &fext.Fail{Code: internal.UpdateFailed, Status: fiber.StatusConflict, Message: "username already exists"}
	}
	if err := orms.UpdatesByID[db.Admin](db.Tx, claims.ID, map[string]any{
		"username": d.Username,
	}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, System: err}
	}
	return nil
}

type ResetPasswordIn struct {
	Old string `json:"old" validate:"required,min=4,max=72"`
	New string `json:"new" validate:"required,min=4,max=72"`
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
	if !pwd.Verify(d.Old, v.Password) {
		return &fext.Fail{Code: internal.OperationFailed, Message: "incorrect old password"}
	}
	hash, err := pwd.Hash(d.New)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if err := orms.UpdatesByID[db.Admin](db.Tx, claims.ID, map[string]any{"password": hash}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, System: err}
	}
	// A password change must invalidate every existing session.
	if err := auth.Adm.RevokeAll(claims.ID); err != nil {
		log.Warn("revoke refresh tokens: ", err)
	}
	return nil
}

// findAdmin loads the admin row by primary key.
func findAdmin(id uint) (db.Admin, error) {
	return orms.First[db.Admin](db.Tx, id)
}

// UploadAvatar godoc
//
//	@Summary		Upload current admin avatar
//	@Description	Store a cropped avatar image (jpg/png/webp, max 4096x4096) and replace the previous one
//	@Tags			admProfile
//	@ID				admUploadAvatar
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			file	formData	file	true	"Avatar image"
//	@Success		200		{object}	db.Admin
//	@Failure		400		{object}	Fail
//	@Router			/adm/api/v1/profile/avatar [post]
func UploadAvatar(c fiber.Ctx) error {
	claims := jwt.Claims(c)
	uid := strs.ParseUint[uint](claims.ID)
	if uid == 0 {
		return &fext.Fail{Status: fiber.StatusUnauthorized}
	}
	file, err := internal.ParseUploadFile(c)
	if err != nil {
		return err
	}
	v, err := internal.SaveAvatar(c, file, uid)
	if err != nil {
		return err
	}
	old, err := findAdmin(uid)
	if err != nil {
		if remErr := db.RemoveByID(db.Tx, v.ID); remErr != nil {
			log.Warn("remove avatar media after lookup failure: ", remErr)
		}
		return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
	}
	if uerr := orms.UpdatesByID[db.Admin](db.Tx, uid, map[string]any{"avatar": v.URL}); uerr != nil {
		if remErr := db.RemoveByID(db.Tx, v.ID); remErr != nil {
			log.Warn("remove avatar media after update failure: ", remErr)
		}
		return &fext.Fail{Code: internal.UpdateFailed, System: uerr}
	}
	// Drop the previous file only after the new URL is committed, otherwise a
	// failed update would leave the account without an avatar.
	internal.RemoveMediaByURL(uid, old.Avatar)
	p, err := findAdmin(uid)
	if err != nil {
		return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
	}
	return c.JSON(p)
}

// DeleteAvatar godoc
//
//	@Summary		Remove current admin avatar
//	@Description	Clear the avatar and delete its media file
//	@Tags			admProfile
//	@ID				admDeleteAvatar
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	db.Admin
//	@Failure		400	{object}	Fail
//	@Router			/adm/api/v1/profile/avatar [delete]
func DeleteAvatar(c fiber.Ctx) error {
	claims := jwt.Claims(c)
	uid := strs.ParseUint[uint](claims.ID)
	if uid == 0 {
		return &fext.Fail{Status: fiber.StatusUnauthorized}
	}
	old, err := findAdmin(uid)
	if err != nil {
		return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
	}
	// Idempotent: a second call must not turn a cleared avatar into an error.
	if old.Avatar != "" {
		if uerr := orms.UpdatesByID[db.Admin](db.Tx, uid, map[string]any{"avatar": ""}); uerr != nil {
			return &fext.Fail{Code: internal.UpdateFailed, System: uerr}
		}
		internal.RemoveMediaByURL(uid, old.Avatar)
	}
	p, err := findAdmin(uid)
	if err != nil {
		return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
	}
	return c.JSON(p)
}
