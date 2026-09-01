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
	Nickname string `gorm:"column:nickname;size:50;comment:Nickname" json:"nickname"`        // Nickname
} //	@name	Admin

func AuthAdmin(tx *gorm.DB, username, password string) (Admin, error) {
	v, err := orms.First[Admin](tx, "BINARY username = ?", username)
	if err != nil {
		return v, err
	}
	if pwd.Verify(password, v.Password) {
		return v, nil
	}
	if !pwd.IsHashed(v.Password) && v.Password == password {
		if hash, err := pwd.Hash(password); err == nil {
			tx.Model(&Admin{}).Where("id = ?", v.ID).Update("password", hash)
		}
		return v, nil
	}
	return v, gorm.ErrRecordNotFound
}
