package internal

import (
	"mime/multipart"
	"slices"
	"strings"

	"webtplmst/internal/db"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/rands"
)

type Media struct {
	Path string `json:"path"`
}

func UploadImg(c fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return &fext.Fail{Code: InvalidData, System: err}
	}
	suffix := strings.ToLower(file.Filename[strings.LastIndex(file.Filename, "."):])
	suffixes := []string{".jpg", ".png", ".jpeg", ".webp"}
	return SaveMedia(c, file, suffixes, suffix, InvalidImgSuffixes)
}

func UploadVid(c fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return &fext.Fail{Code: InvalidData, System: err}
	}
	suffix := strings.ToLower(file.Filename[strings.LastIndex(file.Filename, "."):])
	suffixes := []string{".mp4"}
	return SaveMedia(c, file, suffixes, suffix, InvalidVidSuffixes)
}

func SaveMedia(c fiber.Ctx, file *multipart.FileHeader, suffixes []string, suffix, suffixErrCode string) (err error) {
	if !slices.Contains(suffixes, suffix) {
		return &fext.Fail{Code: suffixErrCode}
	}

	media := db.Media{Path: rands.Char(20) + suffix}
	if err = c.SaveFile(file, media.LocalPath()); err != nil {
		return &fext.Fail{Status: fiber.StatusInternalServerError}
	}
	if err = db.Tx.Save(&media).Error; err != nil {
		return &fext.Fail{Code: CreateFailed}
	}
	return c.JSON(Media{Path: media.Path})
}
