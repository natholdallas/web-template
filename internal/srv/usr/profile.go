package usr

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/natholdallas/natools4go/strs"
	"gorm.io/gorm"

	"webtplmst/internal/db"
	"webtplmst/internal/oauth"
	"webtplmst/internal/srv/internal"
)

type Profile struct {
	orms.Model[uint]
	Username       string  `json:"username"`
	Avatar         string  `json:"avatar"`
	GoogleEmail    *string `json:"googleEmail"`
	GoogleName     *string `json:"googleName"`
	GoogleAvatar   *string `json:"googleAvatar"`
	TelegramName   *string `json:"telegramName"`
	TelegramAvatar *string `json:"telegramAvatar"`
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
	Avatar   string `json:"avatar"`
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
	if err := orms.UpdatesByID[db.User](db.Tx, claims.ID, map[string]any{
		"username": d.Username,
		"avatar":   d.Avatar,
	}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, Message: err.Error()}
	}
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
	if err := orms.UpdatesByID[db.User](db.Tx, claims.ID, map[string]any{"password": hash}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, Message: err.Error()}
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
		return &fext.Fail{Code: internal.OauthFailed, System: err}
	}
	if user.Email == "" {
		return &fext.Fail{Code: internal.OauthFailed, Message: "google email missing"}
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
		return &fext.Fail{Code: internal.UpdateFailed, Message: err.Error()}
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
	if err := oauth.CheckTelegramAuth(&d); err != nil {
		return &fext.Fail{Code: internal.OauthFailed, System: err}
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
		return &fext.Fail{Code: internal.UpdateFailed, Message: err.Error()}
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
