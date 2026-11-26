package db

import (
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"gorm.io/gorm"
)

type Admin struct {
	orms.Model[uint]
	Username string `gorm:"column:username;size:50;unique;comment:Username" json:"username"` // Username
	Password string `gorm:"column:password;size:255;comment:Password" json:"-"`              // Password
	Avatar   string `gorm:"column:avatar;size:512;comment:Avatar URL" json:"avatar"`         // Avatar URL
} //	@name	Admin

func AuthAdmin(tx *gorm.DB, username, password string) (Admin, error) {
	v, err := orms.First[Admin](tx, "username = ?", username)
	if err != nil {
		return v, err
	}
	if !pwd.Verify(password, v.Password) {
		return v, gorm.ErrRecordNotFound
	}
	return v, nil
}
