// Package admin to setup admin route
package adm

import (
	"webtplmst/internal/conf"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
)

var jwt = fext.NewJwt(conf.App.JwtSecretAdm)

func Setup(api fiber.Router) {
	api.Group("/auth").
		Post("/login", Login).
		Post("/refresh", Refresh).
		Post("/logout", Logout)
	api.Group("/profile", jwt.Middleware).
		Get("/me", FindProfile).
		Put("/me", UpdateProfile).
		Put("/password", ResetPassword)
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
}
