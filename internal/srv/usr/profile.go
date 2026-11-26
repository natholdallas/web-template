package usr

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/natholdallas/natools4go/strs"
	"gorm.io/gorm"

	"webgx/internal/auth"
	"webgx/internal/db"
	"webgx/internal/oauth"
	"webgx/internal/srv/internal"
)

type Profile struct {
	orms.Model[uint]
	Username       string  `json:"username"`
	Avatar         string  `json:"avatar"`
	GoogleEmail    *string `json:"googleEmail" extensions:"x-nullable"`
	GoogleName     *string `json:"googleName" extensions:"x-nullable"`
	GoogleAvatar   *string `json:"googleAvatar" extensions:"x-nullable"`
	TelegramName   *string `json:"telegramName" extensions:"x-nullable"`
	TelegramAvatar *string `json:"telegramAvatar" extensions:"x-nullable"`
} //	@name	UsrProfile

// findProfile loads the profile projection for the given user ID.
func findProfile(id string) (Profile, error) {
	return orms.QM[Profile, db.User](db.Tx).First(id)
}

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
	v, err := findProfile(claims.ID)
	if err != nil {
		return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
	}
	return c.JSON(v)
}

// ProfileIn carries the fields a user may change on their own profile. The
// avatar is intentionally absent: it is written only by the avatar endpoints,
// which validate the upload and clean up the previous file.
type ProfileIn struct {
	Username string `json:"username" validate:"required,min=4,max=50"`
} //	@name	UsrProfileIn

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
	if orms.Exists[db.User](db.Tx.Where("username = ? AND id <> ?", d.Username, claims.ID)) {
		return &fext.Fail{Code: internal.UpdateFailed, Status: fiber.StatusConflict, Message: "username already exists"}
	}
	if err := orms.UpdatesByID[db.User](db.Tx, claims.ID, map[string]any{
		"username": d.Username,
	}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, System: err}
	}
	return nil
}

// UploadAvatar godoc
//
//	@Summary		Upload current user avatar
//	@Description	Store a cropped avatar image (jpg/png/webp, max 4096x4096) and replace the previous one
//	@Tags			usrProfile
//	@ID				usrUploadAvatar
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			file	formData	file	true	"Avatar image"
//	@Success		200		{object}	Profile
//	@Failure		400		{object}	Fail
//	@Router			/usr/api/v1/profile/avatar [post]
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
	old, err := orms.First[db.User](db.Tx, uid)
	if err != nil {
		if remErr := db.RemoveByID(db.Tx, v.ID); remErr != nil {
			log.Warn("remove avatar media after lookup failure: ", remErr)
		}
		return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
	}
	if uerr := orms.UpdatesByID[db.User](db.Tx, uid, map[string]any{"avatar": v.URL}); uerr != nil {
		if remErr := db.RemoveByID(db.Tx, v.ID); remErr != nil {
			log.Warn("remove avatar media after update failure: ", remErr)
		}
		return &fext.Fail{Code: internal.UpdateFailed, System: uerr}
	}
	// Drop the previous file only after the new URL is committed, otherwise a
	// failed update would leave the account without an avatar.
	internal.RemoveMediaByURL(uid, old.Avatar)
	p, err := findProfile(claims.ID)
	if err != nil {
		return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
	}
	return c.JSON(p)
}

// DeleteAvatar godoc
//
//	@Summary		Remove current user avatar
//	@Description	Clear the avatar and delete its media file
//	@Tags			usrProfile
//	@ID				usrDeleteAvatar
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	Profile
//	@Failure		400	{object}	Fail
//	@Router			/usr/api/v1/profile/avatar [delete]
func DeleteAvatar(c fiber.Ctx) error {
	claims := jwt.Claims(c)
	uid := strs.ParseUint[uint](claims.ID)
	if uid == 0 {
		return &fext.Fail{Status: fiber.StatusUnauthorized}
	}
	old, err := orms.First[db.User](db.Tx, uid)
	if err != nil {
		return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
	}
	// Idempotent: a second call must not turn a cleared avatar into an error.
	if old.Avatar != "" {
		if uerr := orms.UpdatesByID[db.User](db.Tx, uid, map[string]any{"avatar": ""}); uerr != nil {
			return &fext.Fail{Code: internal.UpdateFailed, System: uerr}
		}
		internal.RemoveMediaByURL(uid, old.Avatar)
	}
	p, err := findProfile(claims.ID)
	if err != nil {
		return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
	}
	return c.JSON(p)
}

type ResetPasswordIn struct {
	Old string `json:"old" validate:"required,min=4,max=72"`
	New string `json:"new" validate:"required,min=4,max=72"`
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
	if !pwd.Verify(d.Old, v.Password) {
		return &fext.Fail{Code: internal.OperationFailed, Message: "incorrect old password"}
	}
	hash, err := pwd.Hash(d.New)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if err := orms.UpdatesByID[db.User](db.Tx, claims.ID, map[string]any{"password": hash}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, System: err}
	}
	// A password change must invalidate every existing session.
	if err := auth.Usr.RevokeAll(claims.ID); err != nil {
		log.Warn("revoke refresh tokens: ", err)
	}
	return nil
}

// BindGoogle godoc
//
//	@Summary		Bind Google account
//	@Description	Bind the Google account from the given OAuth code to the current user
//	@Tags			usrProfile
//	@ID				usrBindGoogle
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			body	body	GoogleAuthIn	true	"Google OAuth code"
//	@Success		200
//	@Failure		400	{object}	Fail
//	@Router			/usr/api/v1/profile/bind/google [put]
func BindGoogle(c fiber.Ctx) error {
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
	claims := jwt.Claims(c)
	if v, ferr := db.FindUserByGoogle(db.Tx, user.Email); ferr == nil && v.ID != strs.ParseUint[uint](claims.ID) {
		return &fext.Fail{Code: internal.OperationFailed, Message: "google account already bound to another user"}
	}
	if err := orms.UpdatesByID[db.User](db.Tx, claims.ID, map[string]any{
		"google_email":  user.Email,
		"google_name":   user.Name,
		"google_avatar": user.Picture,
	}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, System: err}
	}
	return nil
}

// BindTelegram godoc
//
//	@Summary		Bind Telegram account
//	@Description	Verify the Telegram Login Widget payload and bind it to the current user
//	@Tags			usrProfile
//	@ID				usrBindTelegram
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			body	body	oauth.TelegramAuth	true	"Signed Telegram widget payload"
//	@Success		200
//	@Failure		400	{object}	Fail
//	@Router			/usr/api/v1/profile/bind/telegram [put]
func BindTelegram(c fiber.Ctx) error {
	d, err := fext.BodyVarser[oauth.TelegramAuth](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if err = oauth.CheckTelegramAuth(&d); err != nil {
		return &fext.Fail{Code: internal.OauthFailed, Status: fiber.StatusBadRequest, System: err}
	}
	claims := jwt.Claims(c)
	if v, ferr := db.FindUserByTelegram(db.Tx, d.ID); ferr == nil && v.ID != strs.ParseUint[uint](claims.ID) {
		return &fext.Fail{Code: internal.OperationFailed, Message: "telegram account already bound to another user"}
	}
	name := strings.TrimSpace(d.FirstName + " " + d.LastName)
	if err := orms.UpdatesByID[db.User](db.Tx, claims.ID, map[string]any{
		"telegram_id":     d.ID,
		"telegram_name":   name,
		"telegram_avatar": d.PhotoURL,
	}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, System: err}
	}
	return nil
}

// UnbindPlatform godoc
//
//	@Summary		Unbind a login platform
//	@Description	Remove the Google or Telegram binding from the current user
//	@Tags			usrProfile
//	@ID				usrUnbindPlatform
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			platform	path	string	true	"Platform: google or telegram"
//	@Success		200
//	@Failure		400	{object}	Fail
//	@Router			/usr/api/v1/profile/bind/{platform} [delete]
func UnbindPlatform(c fiber.Ctx) error {
	platform := strings.ToLower(c.Params("platform"))
	var cols []string
	switch platform {
	case "google":
		cols = []string{"google_email", "google_name", "google_avatar"}
	case "telegram":
		cols = []string{"telegram_id", "telegram_name", "telegram_avatar"}
	default:
		return &fext.Fail{Code: internal.InvalidData, Message: "unknown platform"}
	}
	values := map[string]any{}
	for _, col := range cols {
		values[col] = gorm.Expr("NULL")
	}
	claims := jwt.Claims(c)
	if err := orms.UpdatesByID[db.User](db.Tx, claims.ID, values); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, System: err}
	}
	return nil
}
