// Package srv to setup server
package srv

import (
	"webtplmst/docs"
	"webtplmst/internal/conf"
	"webtplmst/internal/srv/admin"
	"webtplmst/internal/srv/base"
	"webtplmst/internal/srv/internal"
	"webtplmst/internal/srv/user"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/natholdallas/natools4go/fext"
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
	base.Setup(app.Group("/api/v1").Use(internal.Recorder("Base")))
	admin.Setup(app.Group("/admin/api/v1").Use(internal.Recorder("Admin")))
	user.Setup(app.Group("/user/api/v1").Use(internal.Recorder("User")))
	fext.Listen(app, ":"+conf.App.Port)
}
