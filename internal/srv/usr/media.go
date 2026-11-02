package usr

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/strs"

	"webtplmst/internal/db"
	"webtplmst/internal/srv/internal"
)

type MediaQueries struct {
	orms.Sorter
	orms.Pagination
} //	@name	UsrMediaQueries

type MediaPage struct {
	Total   int64      `json:"total"`
	Page    int64      `json:"page"`
	Content []db.Media `json:"content"`
} //	@name	UsrMediaPage

// ListMedia godoc
//
//	@Summary	List my media files
//	@Tags		usrMedia
//	@ID			usrListMedia
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		query	query		MediaQueries	false	"Query params"
//	@Success	200		{object}	MediaPage
//	@Router		/usr/api/v1/media [get]
func ListMedia(c fiber.Ctx) error {
	q, err := fext.QueryParser[MediaQueries](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	claims := jwt.Claims(c)
	v := orms.QE[db.Media](db.Tx).
		Where("user_id = ?", claims.ID).
		Scopes(q.Sorter.Scope).
		IPaginate(q.Pagination)
	return c.JSON(v)
}

// UploadMedia godoc
//
//	@Summary	Upload a media file
//	@Tags		usrMedia
//	@ID			usrUploadMedia
//	@Accept		multipart/form-data
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		file	formData	file	true	"File to upload"
//	@Param		kind	formData	string	false	"Media kind: image/video/file (default file)"
//	@Success	200		{object}	db.Media
//	@Failure	400		{object}	Fail
//	@Router		/usr/api/v1/media [post]
func UploadMedia(c fiber.Ctx) error {
	file, err := internal.ParseUploadFile(c)
	if err != nil {
		return err
	}
	claims := jwt.Claims(c)
	kind := internal.Kind(strings.ToLower(c.FormValue("kind")))
	if kind == "" {
		kind = internal.KindFile
	}
	v, err := internal.SaveFile(c, file, kind, strs.ParseUint[uint](claims.ID))
	if err != nil {
		return err
	}
	return c.JSON(v)
}

// RemoveMedia godoc
//
//	@Summary	Remove my media file
//	@Tags		usrMedia
//	@ID			usrDeleteMedia
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		id	path	int	true	"Media ID"
//	@Success	200
//	@Failure	400	{object}	Fail
//	@Router		/usr/api/v1/media/{id} [delete]
func RemoveMedia(c fiber.Ctx) error {
	claims := jwt.Claims(c)
	v := orms.IFirst[db.Media](db.Tx, c.Params("id"))
	if v.ID == 0 {
		return &fext.Fail{Code: internal.RecordNotFound}
	}
	if v.UserID != strs.ParseUint[uint](claims.ID) {
		return &fext.Fail{Status: fiber.StatusForbidden}
	}
	if err := orms.Delete[db.Media](db.Tx, v.ID); err != nil {
		return &fext.Fail{Code: internal.RemoveFailed, System: err}
	}
	return nil
}
