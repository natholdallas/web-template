// Package usr registers the user API routes.
package usr

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

// userKey buckets authenticated rate limits by user ID.
func userKey(c fiber.Ctx) string {
	if claims := jwt.Claims(c); claims != nil {
		return claims.ID
	}
	return ""
}

func Setup(api fiber.Router) {
	jwt = fext.NewJWT(conf.App.SecretUsr)
	api.Group("/auth").
		Use(internal.OriginGuard).
		Use(internal.RateLimit(internal.IPKey, "usr:auth", 30, time.Minute)).
		Post("/login", Login).
		Post("/refresh", Refresh).
		Post("/logout", Logout).
		Post("/google", SignInGoogle).
		Post("/telegram", SignInTelegram)
	api.Group("/profile", jwt.Middleware).
		Get("/me", FindProfile).
		Put("/me", UpdateProfile).
		Put("/password", ResetPassword).
		Post("/avatar", internal.RateLimit(userKey, "usr:avatar", 20, time.Hour), UploadAvatar).
		Delete("/avatar", DeleteAvatar).
		Put("/bind/google", BindGoogle).
		Put("/bind/telegram", BindTelegram).
		Delete("/bind/:platform", UnbindPlatform)
	api.Group("/media", jwt.Middleware).
		Get("", ListMedia).
		Post("", internal.RateLimit(userKey, "usr:media", 30, time.Hour), UploadMedia).
		Delete("/:id", RemoveMedia)
}
