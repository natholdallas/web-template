// Package user to setup user route
package usr

import (
	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"

	"webtplmst/internal/conf"
)

var jwt = fext.NewJWT(conf.App.SecretUsr)

func Setup(api fiber.Router) {
	api.Group("/auth").
		Post("/login", Login).
		Post("/refresh", Refresh).
		Post("/logout", Logout)
	api.Group("/profile", jwt.Middleware).
		Get("/me", FindProfile).
		Put("/me", UpdateProfile).
		Put("/password", ResetPassword)
}
