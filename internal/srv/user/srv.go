// Package user to setup user route
package user

import (
	"time"

	"webtplmst/internal/conf"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/natholdallas/natools4go/fext"
)

var jwt = fext.NewJwt(conf.App.SecretUser)

func Setup(api fiber.Router) {
	api.Use(logger.New(logger.Config{
		TimeFormat: time.DateTime,
		Format:     "[User]" + fext.StdLogFmt,
	}))
	api.Group("/auth").
		Post("/in", SignIn)
	api.Group("/user", jwt.Middleware).
		Get("/", FindUser).
		Post("/reset/password", ResetPassword).
		Put("/", UpdateUser)
}
