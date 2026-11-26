// Package srv to setup server
package srv

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/strs"

	"webgx/internal/conf"
	"webgx/internal/srv/adm"
	"webgx/internal/srv/internal"
	"webgx/internal/srv/std"
	"webgx/internal/srv/swg"
	"webgx/internal/srv/usr"
)

// Run serves HTTP until ctx is cancelled and then shuts the listener down
// gracefully. A failure to listen is returned to the caller.
func Run(ctx context.Context) error {
	return fext.Listen(newApp(), strs.ToStart(conf.App.Port, strs.Colon), fiber.ListenConfig{
		GracefulContext: ctx,
	})
}

// newApp builds the Fiber app: middleware first, then one group per API surface.
func newApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      conf.App.Name,
		ErrorHandler: fext.ErrorHandler,
		BodyLimit:    conf.App.BodyLimit * 1024 * 1024,
		// Only a loopback proxy (the bundled nginx) may set the client IP;
		// a direct client cannot forge X-Forwarded-For for rate-limit buckets.
		ProxyHeader:        fiber.HeaderXForwardedFor,
		TrustProxy:         true,
		TrustProxyConfig:   fiber.TrustProxyConfig{Loopback: true},
		EnableIPValidation: true,
		// Bound slow clients: without these a single stalled connection holds a
		// goroutine and a file descriptor indefinitely.
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	})
	// Without an explicit recover middleware a single panicking handler takes
	// the whole process down: Fiber v3 does not install one by default.
	app.Use(recover.New())
	app.Use(func(c fiber.Ctx) error {
		c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
		c.Set(fiber.HeaderXFrameOptions, "DENY")
		c.Set(fiber.HeaderReferrerPolicy, "no-referrer")
		c.Set(fiber.HeaderPermissionsPolicy, "camera=(), microphone=(), geolocation=()")
		return c.Next()
	})
	app.Use(requestid.New())
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: internal.AllowOriginsFunc,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		// The refresh token is a cookie, so the browser only stores/echoes it
		// when the response allows credentials. Origin echoing stays exact
		// through AllowOriginsFunc above.
		AllowCredentials: true,
	}))
	std.Setup(app.Group("/api/v1").Use(internal.Log("GEN")))
	adm.Setup(app.Group("/adm/api/v1").Use(internal.Log("ADM")))
	usr.Setup(app.Group("/usr/api/v1").Use(internal.Log("USR")))
	swg.Setup(app.Group("/doc/api/v1").Use(internal.Log("DOC")))
	// Filesystem middleware last: an API path must never be answered by a file
	// that happens to sit at the same route.
	app.Use("/", static.New(conf.App.RWeb, static.Config{
		Next: internal.NginxMiddleware,
	}))
	app.Use(conf.App.MediaRoute, static.New(conf.App.RMedia, static.Config{
		Next: internal.NginxMiddleware,
	}))
	// SendFile, not static.New: static.New takes a root directory, so pointing it
	// at a file made every request look for <file>/favicon.ico.
	app.Get("/favicon.ico", func(c fiber.Ctx) error {
		return c.SendFile("./assets/favicon.ico")
	})
	return app
}
