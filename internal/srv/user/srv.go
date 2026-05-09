// Package user to setup user route
package user

import (
	"webtplmst/internal/conf"
	"webtplmst/internal/srv/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
)

var jwt = fext.NewJwt(conf.App.SecretUser)

func Setup(api fiber.Router) {
	api.Use(internal.FastLogger("User"))
	api.Group("/auth").
		Post("/in", SignIn)
	api.Group("/user", jwt.Middleware).
		Get("/", FindUser).
		Post("/reset/password", ResetPassword).
		Put("/", UpdateUser)
}
