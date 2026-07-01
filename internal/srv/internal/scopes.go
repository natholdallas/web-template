package internal

import (
	"github.com/natholdallas/natools4go/orms"
	"gorm.io/gorm"
)

func UserScope(id string) orms.GormScope {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where("user_id = ?", id)
	}
}

func IDScope(id any) orms.GormScope {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id = ?", id)
	}
}
