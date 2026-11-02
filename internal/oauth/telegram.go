package oauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"webtplmst/internal/conf"
)

// TelegramAuth is the raw signed payload returned by the Telegram Login
// Widget callback. Field names deliberately match the widget's JSON.
type TelegramAuth struct {
	ID        string `json:"id" validate:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date" validate:"required"`
	Hash      string `json:"hash" validate:"required"`
} //	@name	TelegramAuth

// CheckTelegramAuth verifies the Telegram Login Widget signature and the
// freshness of the payload. See https://core.telegram.org/widgets/login.
func CheckTelegramAuth(v *TelegramAuth) error {
	if v.ID == "" || v.Hash == "" {
		return errors.New("missing telegram auth fields")
	}
	if conf.App.OauthTelegramBotToken == "" {
		return errors.New("telegram oauth not configured")
	}
	if time.Since(time.Unix(v.AuthDate, 0)) > 5*time.Minute {
		return errors.New("telegram auth data expired")
	}

	secret := sha256.Sum256([]byte(conf.App.OauthTelegramBotToken))
	fields := map[string]string{
		"id":         v.ID,
		"first_name": v.FirstName,
		"last_name":  v.LastName,
		"username":   v.Username,
		"photo_url":  v.PhotoURL,
		"auth_date":  fmt.Sprintf("%d", v.AuthDate),
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		if fields[k] != "" {
			parts = append(parts, k+"="+fields[k])
		}
	}
	dataCheckString := strings.Join(parts, "\n")

	mac := hmac.New(sha256.New, secret[:])
	mac.Write([]byte(dataCheckString))
	if hex.EncodeToString(mac.Sum(nil)) != v.Hash {
		return errors.New("telegram auth hash mismatch")
	}
	return nil
}
