package internal

import (
	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/strs"

	"webtplmst/internal/conf"
)

// DebugMiddleware only allows access to the debug mode
func AllowOriginsFunc(origin string) bool {
	if conf.App.Debug {
		return strs.AnyPrefix(origin, conf.App.CorsDev...)
	} else {
		return strs.AnyPrefix(origin, conf.App.CorsPrd...)
	}
}

// DebugMiddleware only allows access to the debug mode
func DebugMiddleware(c fiber.Ctx) error {
	if conf.App.Debug {
		return c.Next()
	}
	return &fext.Fail{Status: fiber.StatusForbidden}
}

// NginxMiddleware only allows access to the nginx
func NginxMiddleware(c fiber.Ctx) bool {
	return conf.App.Nginx
}

// SwaggerMiddleware only allows access to the swagger
func SwaggerMiddleware(c fiber.Ctx) error {
	if conf.App.Swagger {
		return c.Next()
	}
	return &fext.Fail{Status: fiber.StatusForbidden}
}
