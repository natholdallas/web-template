// Package srv to setup server
package srv

import (
	"webtplmst/docs"
	"webtplmst/internal/conf"
	"webtplmst/internal/srv/admin"
	"webtplmst/internal/srv/base"
	"webtplmst/internal/srv/user"

	"github.com/fsnotify/fsnotify"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
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

	fext.SetDebugMode(conf.App.Debug)
	fext.SetErrorFunc(func(err error) { log.Error(err) })
	fext.SetLogLevel(conf.App.LogLevelFiber)

	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: AllowOriginsFunc,
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

	base.Setup(app.Group("/api/v1"))
	admin.Setup(app.Group("/admin/api/v1"))
	user.Setup(app.Group("/user/api/v1"))
	fext.Listen(app, ":"+conf.App.Port)
}

func AllowOriginsFunc(origin string) bool {
	if conf.App.Debug {
		return strs.AnyPrefix(origin, conf.App.CorsDev...)
	} else {
		return strs.AnyPrefix(origin, conf.App.CorsPrd...)
	}
}

func Reload(fsnotify.Event) {
	fext.SetLogLevel(conf.App.LogLevelFiber)
	fext.SetDebugMode(conf.App.Debug)
}
