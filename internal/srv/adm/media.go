package adm

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
	"gorm.io/gorm"

	"webgx/internal/db"
	"webgx/internal/srv/internal"
)

type MediaQueries struct {
	orms.Sorter
	orms.Pagination
} //	@name	AdmMediaQueries

type MediaPage struct {
	Total   int64      `json:"total"`
	Page    int64      `json:"page"`
	Content []db.Media `json:"content"`
} //	@name	AdmMediaPage

// ListMedia godoc
//
//	@Summary	List all media files
//	@Tags		admMedia
//	@ID			admListMedia
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		query	query		MediaQueries	false	"Query params"
//	@Success	200		{object}	MediaPage
//	@Router		/adm/api/v1/media [get]
func ListMedia(c fiber.Ctx) error {
	q, err := fext.QueryParser[MediaQueries](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	v := orms.QE[db.Media](db.Tx).
		Scopes(q.Sorter.Scope).
		IPaginate(q.Pagination)
	return c.JSON(v)
}

// UploadMedia godoc
//
//	@Summary	Upload a media file
//	@Tags		admMedia
//	@ID			admUploadMedia
//	@Accept		multipart/form-data
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		file	formData	file	true	"File to upload"
//	@Param		kind	formData	string	false	"Media kind: image/video/file (default file)"
//	@Success	200		{object}	db.Media
//	@Failure	400		{object}	Fail
//	@Router		/adm/api/v1/media [post]
func UploadMedia(c fiber.Ctx) error {
	file, err := internal.ParseUploadFile(c)
	if err != nil {
		return err
	}
	kind := internal.Kind(strings.ToLower(c.FormValue("kind")))
	if kind == "" {
		kind = internal.KindOf(file.Filename)
	}
	v, err := internal.SaveFile(c, file, kind, 0)
	if err != nil {
		return err
	}
	return c.JSON(v)
}

// RemoveMedia godoc
//
//	@Summary	Remove a media file
//	@Tags		admMedia
//	@ID			admDeleteMedia
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		id	path	int	true	"Media ID"
//	@Success	200
//	@Failure	400	{object}	Fail
//	@Router		/adm/api/v1/media/{id} [delete]
func RemoveMedia(c fiber.Ctx) error {
	id, err := internal.ParseID(c)
	if err != nil {
		return err
	}
	v, err := orms.First[db.Media](db.Tx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
		}
		return &fext.Fail{Code: internal.RemoveFailed, System: err}
	}
	if internal.AvatarInUse(v.URL) {
		return &fext.Fail{Status: fiber.StatusConflict, Code: internal.MediaInUse}
	}
	if err := db.RemoveByID(db.Tx, v.ID); err != nil {
		return &fext.Fail{Code: internal.RemoveFailed, System: err}
	}
	return nil
}
