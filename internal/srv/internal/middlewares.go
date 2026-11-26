package internal

import (
	"net/url"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"

	"webgx/internal/conf"
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

// OriginGuard rejects browser requests from a foreign origin on the auth
// endpoints.
//
// The refresh token travels in an HttpOnly cookie, which the browser attaches
// to any request to this host regardless of who initiated it. Without this
// guard a third-party page could POST to /auth/refresh and rotate the victim's
// session. Requests with no Origin header (curl, server-to-server) are allowed
// because they cannot carry a victim's cookie, and same-origin requests are
// allowed even when the deployment does not list itself in [cors] — a
// reverse-proxied single-host install and the backend-fronted SPA both post
// same-origin.
func OriginGuard(c fiber.Ctx) error {
	origin := c.Get(fiber.HeaderOrigin)
	if origin == "" {
		return c.Next()
	}
	if AllowOriginsFunc(origin) {
		return c.Next()
	}
	if u, err := url.Parse(origin); err == nil && strings.EqualFold(u.Host, c.Host()) {
		return c.Next()
	}
	return &fext.Fail{Status: fiber.StatusForbidden}
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
