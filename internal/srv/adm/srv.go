// Package admin to setup admin route
package adm

import (
	"webtplmst/internal/conf"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
)

var jwt = fext.NewJwt(conf.App.SecretAdm)

func Setup(api fiber.Router) {
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
