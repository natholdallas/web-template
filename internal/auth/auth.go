// Package auth provides access/refresh JWT pairs backed by Redis revocation.
package auth

import (
	"errors"
	"fmt"
	"time"
	"webtplmst/internal/db"

	"github.com/golang-jwt/jwt/v5"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/rands"
	"github.com/redis/go-redis/v9"
)

// Pair is an access token and its matching refresh token.
type Pair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Auth manages token issuance and refresh-token revocation for one audience.
type Auth struct {
	SecretKey  string
	Prefix     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// New builds an Auth manager. Prefix (e.g. "adm"/"usr") namespaces the Redis keys.
func New(secretKey, prefix string, accessTTL, refreshTTL time.Duration) Auth {
	return Auth{secretKey, prefix, accessTTL, refreshTTL}
}

func (a *Auth) redisKey(jti string) string {
	return fmt.Sprintf("%s:refresh:%s", a.Prefix, jti)
}

// GenPair issues an access token and a refresh token for the given user ID.
// The refresh token is recorded in Redis so it can be revoked or rotated.
func (a *Auth) GenPair(userID string) (Pair, error) {
	access, err := fext.GenToken(userID, a.SecretKey, a.AccessTTL)
	if err != nil {
		return Pair{}, err
	}
	jti := rands.Char(32)
	claims := jwt.RegisteredClaims{
		ID:        jti,
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.RefreshTTL)),
	}
	refresh, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(a.SecretKey))
	if err != nil {
		return Pair{}, err
	}
	if db.Rx != nil {
		db.Rx.Set(db.Rc, a.redisKey(jti), userID, a.RefreshTTL)
	}
	return Pair{access, refresh}, nil
}

// VerifyRefresh validates a refresh token and returns its user ID.
// It rejects tokens whose jti has been revoked or is unknown to Redis.
func (a *Auth) VerifyRefresh(token string) (string, error) {
	claims, err := fext.ParseToken(token, a.SecretKey)
	if err != nil {
		return "", err
	}
	if claims.ID == "" || claims.Subject == "" {
		return "", errors.New("invalid refresh token")
	}
	if db.Rx == nil {
		return "", errors.New("redis unavailable")
	}
	val, err := db.Rx.Get(db.Rc, a.redisKey(claims.ID)).Result()
	if err == redis.Nil {
		return "", errors.New("refresh token revoked")
	}
	if err != nil {
		return "", err
	}
	if val != claims.Subject {
		return "", errors.New("refresh token mismatch")
	}
	return claims.Subject, nil
}

// RevokeRefresh invalidates the given refresh token in Redis.
func (a *Auth) RevokeRefresh(token string) error {
	claims, err := fext.ParseToken(token, a.SecretKey)
	if err != nil {
		return err
	}
	if db.Rx == nil {
		return errors.New("redis unavailable")
	}
	return db.Rx.Del(db.Rc, a.redisKey(claims.ID)).Err()
}
