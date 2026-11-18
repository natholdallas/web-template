package internal

import (
	"slices"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"

	"webtplmst/internal/conf"
)

// AllowOriginsFunc reports whether the request origin may access the API.
//
// The comparison is exact on purpose. A prefix match lets an attacker register
// "https://example.com.attacker.net" (or "https://example.com.evil.tld") and be
// treated as trusted, because that string starts with the whitelisted origin.
// Local development subdomains are covered by listing "http://localhost" and
// friends explicitly, not by loosening the match.
func AllowOriginsFunc(origin string) bool {
	if origin == "" {
		// Non-browser clients send no Origin header; CORS does not apply.
		return false
	}
	if conf.App.Debug {
		return slices.Contains(conf.App.CorsDev, origin)
	}
	return slices.Contains(conf.App.CorsPrd, origin)
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
