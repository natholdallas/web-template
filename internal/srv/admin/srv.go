// Package admin to setup admin route
package admin

import (
	"time"

	"webtplmst/internal/conf"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/natholdallas/natools4go/fext"
)

var jwt = fext.NewJwt(conf.App.SecretAdmin)

func Setup(api fiber.Router) {
	api.Use(logger.New(logger.Config{
		TimeFormat: time.DateTime,
		Format:     "[Admin]" + fext.StdLogFmt,
	}))
	api.Group("/auth").
		Post("/in", SignIn)
	api.Group("/user", jwt.Middleware).
		Get("", ListUser).
		Get("/:id", FindUser).
		Post("", CreateUser).
		Put("/:id", UpdateUser).
		Delete("/:id", RemoveUser)
	api.Group("/admin", jwt.Middleware).
		Get("", ListAdmin).
		Get("/:id", FindAdmin).
		Post("", CreateAdmin).
		Put("/:id", UpdateAdmin).
		Delete("/:id", RemoveAdmin)
}
