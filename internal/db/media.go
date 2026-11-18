package db

import (
	"errors"
	"io/fs"
	"os"

	"github.com/natholdallas/natools4go/orms"
	"gorm.io/gorm"

	"webtplmst/internal/conf"
)

// Media is a generic static resource (image/video/file) stored on local disk
// and served through the configured media base URL (route or domain).
type Media struct {
	orms.Model[uint]
	Name   string `gorm:"column:name;size:256;comment:Original filename" json:"name"`             // Original filename
	Path   string `gorm:"column:path;size:255;not null;unique;comment:Relative path" json:"path"` // Relative path under the media dir
	URL    string `gorm:"column:url;size:512;comment:Public URL" json:"url"`                      // Public URL (media base + path)
	MIME   string `gorm:"column:mime;size:64;comment:MIME type" json:"mime"`                      // MIME type
	Kind   string `gorm:"column:kind;size:16;comment:Media kind image/video/file" json:"kind"`    // Media kind
	Size   int64  `gorm:"column:size;comment:Byte size" json:"size"`                              // Byte size
	UserID uint   `gorm:"column:user_id;comment:Owner user ID (0=system)" json:"user_id"`         // Owner user ID (0=system)
} //	@name	Media

// AfterDelete removes the physical file when the record is deleted.
//
// GORM calls hooks on the value being deleted. Deleting by primary key
// (orms.DeleteByID, tx.Delete(&Media{}, id)) passes a zero-valued instance, so
// Lookup/Remove must load the row first — see RemoveByID.
func (s *Media) AfterDelete(_ *gorm.DB) error {
	if s.Path == "" {
		return nil
	}
	if err := os.Remove(s.LocalPath()); err != nil && !errors.Is(err, fs.ErrNotExist) {
		// The row is already gone; a missing file is fine, anything else is not.
		return err
	}
	return nil
}

// RemoveByID deletes a media record and its file. It loads the row first so the
// AfterDelete hook receives a populated Path; deleting straight by primary key
// would leave the file orphaned on disk.
func RemoveByID(tx *gorm.DB, id uint) error {
	v, err := orms.First[Media](tx, id)
	if err != nil {
		return err
	}
	return tx.Delete(&v).Error
}

// LocalPath returns the absolute path of the file on disk.
func (s *Media) LocalPath() string {
	return conf.App.RMedia + "/" + s.Path
}

// OpenFile opens the media file for reading.
func (s *Media) OpenFile() (*os.File, error) {
	return os.Open(s.LocalPath())
}
