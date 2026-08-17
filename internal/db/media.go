package db

import (
	"os"

	"webtplmst/internal/conf"

	"github.com/natholdallas/natools4go/orms"
	"gorm.io/gorm"
)

type Media struct {
	orms.Model[uint]
	Name string `gorm:"column:name;size:200;not null;comment:Name" json:"name"`
	Path string `gorm:"column:path;size:200;not null;unique;comment:Path" json:"path"`
} //	@name	Media

func (s *Media) AfterDelete(tx *gorm.DB) error {
	return os.Remove(s.LocalPath())
}

func (s *Media) LocalPath() string {
	return conf.App.RMedia + "/" + s.Path
}

func (s *Media) OpenFile() (*os.File, error) {
	return os.Open(s.LocalPath())
}
