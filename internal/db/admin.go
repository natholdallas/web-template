package db

import (
	"github.com/natholdallas/natools4go/orms"
	"gorm.io/gorm"
)

type Admin struct {
	orms.Model[uint]
	Username string `gorm:"column:username;size:50;unique;comment:Username" json:"username"` // Username
	Password string `gorm:"column:password;size:50;comment:Password" json:"password"`        // Password
} //	@name	Admin

func AuthAdmin(tx *gorm.DB, username, password string) (Admin, error) {
	return orms.First[Admin](tx, "BINARY username = ? AND BINARY password = ?", username, password)
}
