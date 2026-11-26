// Package adm registers the admin API routes.
package adm

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"

	"webgx/internal/conf"
	"webgx/internal/srv/internal"
)

// jwt is built in Setup so the package never reads the configuration at
// initialization time.
var jwt fext.JWT

// userKey buckets authenticated rate limits by admin ID.
func userKey(c fiber.Ctx) string {
	if claims := jwt.Claims(c); claims != nil {
		return claims.ID
	}
	return ""
}

func Setup(api fiber.Router) {
	jwt = fext.NewJWT(conf.App.SecretAdm)
	api.Group("/auth").
		Use(internal.OriginGuard).
		Use(internal.RateLimit(internal.IPKey, "adm:auth", 30, time.Minute)).
		Post("/login", Login).
		Post("/refresh", Refresh).
		Post("/logout", Logout)
	api.Group("/profile", jwt.Middleware).
		Get("/me", FindProfile).
		Put("/me", UpdateProfile).
		Put("/password", ResetPassword).
		Post("/avatar", internal.RateLimit(userKey, "adm:avatar", 20, time.Hour), UploadAvatar).
		Delete("/avatar", DeleteAvatar)
	api.Group("/user", jwt.Middleware).
		Get("", ListUser).
		Get("/:id", FindUser).
		Post("", CreateUser).
		Put("/:id", UpdateUser).
		Delete("/:id", RemoveUser).
		Post("/:id/reset-password", ResetUserPassword)
	api.Group("/admin", jwt.Middleware).
		Get("", ListAdmin).
		Get("/:id", FindAdmin).
		Post("", CreateAdmin).
		Put("/:id", UpdateAdmin).
		Delete("/:id", RemoveAdmin).
		Post("/:id/reset-password", ResetAdminPassword)
	api.Group("/stat", jwt.Middleware).
		Get("", GetStats)
	api.Group("/media", jwt.Middleware).
		Get("", ListMedia).
		Post("", internal.RateLimit(userKey, "adm:media", 60, time.Hour), UploadMedia).
		Delete("/:id", RemoveMedia)
	api.Group("/rate", jwt.Middleware).
		Get("", ListRate).
		Post("/sync", SyncRate)
	api.Group("/config", jwt.Middleware).
		Get("/status", GetConfigStatus)
}
