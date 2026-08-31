// Package pwd provides argon2id password hashing helpers.
package pwd

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	prefix  = "$argon2id$v=19$m=65536,t=3,p=2$"
	saltLen = 16
	keyLen  = 32
)

// Hash derives an argon2id hash from the given plaintext password.
func Hash(plain string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(plain), salt, 3, 64*1024, 2, keyLen)
	return prefix + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(key), nil
}

func TryHash(plain string) string {
	hash, err := Hash(plain)
	if err != nil {
		return ""
	}
	return hash
}

// Verify reports whether the plaintext password matches the stored hash.
func Verify(plain, encoded string) bool {
	if !IsHashed(encoded) {
		return false
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false
	}
	var memory, timeCost, parallelism int
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &parallelism); err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	derived := argon2.IDKey([]byte(plain), salt, uint32(timeCost), uint32(memory), uint8(parallelism), uint32(len(key)))
	return subtle.ConstantTimeCompare(key, derived) == 1
}

// IsHashed reports whether the stored value is an argon2id hash.
func IsHashed(s string) bool {
	return strings.HasPrefix(s, prefix)
}
