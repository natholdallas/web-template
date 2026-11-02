package db

import (
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"gorm.io/gorm"
)

type User struct {
	orms.Model[uint]
	Username     string       `gorm:"column:username;size:50;unique;comment:Username" json:"username"` // Username
	Password     string       `gorm:"column:password;size:255;comment:Password" json:"-"`              // Password
	Avatar       string       `gorm:"column:avatar;size:512;comment:Avatar URL" json:"-"`              // Avatar URL
	GoogleInfo   GoogleInfo   `gorm:"embedded;embeddedPrefix:google_" json:"-"`                        // Google login info
	TelegramInfo TelegramInfo `gorm:"embedded;embeddedPrefix:telegram_" json:"-"`                      // Telegram login info
} //	@name	User

// GoogleInfo holds the Google login bindings (columns prefixed google_).
type GoogleInfo struct {
	Email  *string `gorm:"column:email;unique;size:256;comment:Google email" json:"email"` // Google email
	Name   *string `gorm:"column:name;size:256;comment:Google display name" json:"name"`   // Google display name
	Avatar *string `gorm:"column:avatar;size:256;comment:Google avatar URL" json:"avatar"` // Google avatar URL
} //	@name	GoogleInfo

// TelegramInfo holds the Telegram login bindings (columns prefixed telegram_).
type TelegramInfo struct {
	ID     *string `gorm:"column:id;unique;size:256;comment:Telegram user ID" json:"id"`     // Telegram user ID
	Name   *string `gorm:"column:name;size:256;comment:Telegram display name" json:"name"`   // Telegram display name
	Avatar *string `gorm:"column:avatar;size:256;comment:Telegram avatar URL" json:"avatar"` // Telegram avatar URL
} //	@name	TelegramInfo

// FindUserByGoogle returns the user bound to the given Google email.
func FindUserByGoogle(tx *gorm.DB, email string) (User, error) {
	return orms.First[User](tx, "google_email = ?", email)
}

// FindUserByTelegram returns the user bound to the given Telegram numeric ID.
func FindUserByTelegram(tx *gorm.DB, id string) (User, error) {
	return orms.First[User](tx, "telegram_id = ?", id)
}

func AuthUser(tx *gorm.DB, username, password string) (User, error) {
	v, err := orms.First[User](tx, "username = ?", username)
	if err != nil {
		return v, err
	}
	if pwd.Verify(password, v.Password) {
		return v, nil
	}
	if !pwd.IsHashed(v.Password) && v.Password == password {
		if hash, err := pwd.Hash(password); err == nil {
			tx.Model(&User{}).Where("id = ?", v.ID).Update("password", hash)
		}
		return v, nil
	}
	return v, gorm.ErrRecordNotFound
}
