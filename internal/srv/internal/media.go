package internal

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/rands"

	"webtplmst/internal/conf"
	"webtplmst/internal/db"
)

// Kind classifies an uploaded file and drives its validation rules.
type Kind string

const (
	KindImage Kind = "image"
	KindVideo Kind = "video"
	KindFile  Kind = "file"
)

// kindConf holds the allowed extension and MIME rules for one kind.
type kindConf struct {
	exts  []string
	mimes []string
}

var kindConfs = map[Kind]kindConf{
	KindImage: {
		exts:  []string{".jpg", ".jpeg", ".png", ".webp", ".gif"},
		mimes: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
	},
	KindVideo: {
		exts:  []string{".mp4", ".webm", ".mov", ".m4v"},
		mimes: []string{"video/mp4", "video/webm", "video/quicktime"},
	},
	KindFile: {
		exts: []string{
			".pdf", ".zip", ".gz", ".7z", ".rar", ".doc", ".docx",
			".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".md", ".csv",
		},
	},
}

// safeFileMIME reports whether a sniffed MIME type may be served statically
// without running active content (blocks HTML/SVG/JS/XML that could XSS).
func safeFileMIME(mime string) bool {
	blocked := []string{
		"text/html", "application/xhtml+xml", "image/svg+xml",
		"application/javascript", "text/javascript", "application/x-javascript",
		"application/xml", "text/xml", "application/x-httpd-php",
	}
	if slices.Contains(blocked, mime) {
		return false
	}
	allowed := []string{
		"image/", "video/", "audio/", "text/plain", "application/pdf",
		"application/zip", "application/gzip", "application/x-tar",
		"application/msword", "application/vnd.openxmlformats-officedocument.",
		"application/vnd.ms-", "application/json", "application/octet-stream",
	}
	for _, p := range allowed {
		if strings.HasPrefix(mime, p) {
			return true
		}
	}
	return false
}

// sniffMIME sniffs the real content type of an uploaded file using its first
// bytes instead of trusting the client-provided filename extension.
func sniffMIME(file *multipart.FileHeader) string {
	src, err := file.Open()
	if err != nil {
		return ""
	}
	defer src.Close()
	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil && err != io.EOF {
		return ""
	}
	return http.DetectContentType(buf[:n])
}

// SaveFile validates and stores an uploaded file of the given kind, returning
// the persisted media record. The file is stored under a dated sub-directory
// and removed again if the DB write fails.
func SaveFile(c fiber.Ctx, file *multipart.FileHeader, kind Kind, userID uint) (*db.Media, error) {
	cf, ok := kindConfs[kind]
	if !ok {
		return nil, &fext.Fail{Code: InvalidMedia, Message: "unsupported media kind"}
	}

	suffix := strings.ToLower(filepath.Ext(file.Filename))
	if suffix == "" || !slices.Contains(cf.exts, suffix) {
		return nil, &fext.Fail{Code: InvalidMedia, Message: "invalid file extension"}
	}

	if max := int64(conf.App.MediaMaxSize) * 1024 * 1024; max > 0 && file.Size > max {
		return nil, &fext.Fail{Code: InvalidMedia, Message: "file too large"}
	}

	mime := sniffMIME(file)
	if kind == KindFile {
		if !safeFileMIME(mime) {
			return nil, &fext.Fail{Code: InvalidMedia, Message: "file type not allowed"}
		}
	} else if !slices.Contains(cf.mimes, mime) {
		return nil, &fext.Fail{Code: InvalidMedia, Message: "file content does not match its extension"}
	}

	path := time.Now().Format("2006/01") + "/" + rands.Char(20) + suffix
	dest := conf.App.RMedia + "/" + path
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return nil, &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}
	if err := c.SaveFile(file, dest); err != nil {
		return nil, &fext.Fail{Status: fiber.StatusInternalServerError, System: err}
	}

	v := db.Media{
		Name:   file.Filename,
		Path:   path,
		URL:    conf.App.MediaBase() + "/" + path,
		MIME:   mime,
		Kind:   string(kind),
		Size:   file.Size,
		UserID: userID,
	}
	if err := db.Tx.Save(&v).Error; err != nil {
		if remErr := os.Remove(dest); remErr != nil {
			log.Warn("remove media file after DB failure: ", remErr)
		}
		return nil, &fext.Fail{Code: UploadFailed, System: err}
	}
	return &v, nil
}

// ParseUploadFile reads the "file" field from the multipart request.
func ParseUploadFile(c fiber.Ctx) (*multipart.FileHeader, error) {
	file, err := c.FormFile(UploadFormFile)
	if err != nil {
		return nil, &fext.Fail{Code: InvalidData, System: err}
	}
	return file, nil
}
