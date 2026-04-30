package db

import (
	"github.com/natholdallas/natools4go/orms"
	"gorm.io/gorm"
)

type User struct {
	orms.Model[uint]
	Username string `gorm:"column:username;size:50;unique" json:"username"` // Username
	Password string `gorm:"column:password;size:50" json:"password"`        // Password
} //	@name	User

func AuthUser(tx *gorm.DB, username, password string) (User, error) {
	return orms.First[User](tx, "BINARY username = ? AND BINARY password = ?", username, password)
}
