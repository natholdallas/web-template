// Package srv to setup server
package srv

import (
	"webtplmst/docs"
	"webtplmst/internal/conf"
	"webtplmst/internal/srv/adm"
	"webtplmst/internal/srv/internal"
	"webtplmst/internal/srv/std"
	"webtplmst/internal/srv/usr"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/strs"
	"github.com/yokeTH/gofiber-scalar/scalar/v3"
)

func Setup() {
	app := fiber.New(fiber.Config{
		AppName:      conf.App.Name,
		ErrorHandler: fext.ErrorHandler,
	})
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: conf.App.AllowOriginsFunc,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
	}))
	app.Use("/", static.New(conf.App.RWeb, static.Config{
		Next: conf.App.NginxMiddleware,
	}))
	app.Use("/media", static.New(conf.App.RMedia, static.Config{
		Next: conf.App.NginxMiddleware,
	}))
	app.Get("/docs/*", scalar.New(scalar.Config{
		Theme:             scalar.ThemeSaturn,
		FileContentString: docs.SwaggerInfo.ReadDoc(),
		Title:             "API Documentation",
	}))
	std.Setup(app.Group("/api/v1").Use(internal.Log("Base")))
	adm.Setup(app.Group("/adm/api/v1").Use(internal.Log("Admin")))
	usr.Setup(app.Group("/usr/api/v1").Use(internal.Log("User")))
	fext.Listen(app, strs.ToStart(conf.App.Port, ":"))
}
