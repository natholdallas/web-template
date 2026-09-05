// Package srv to setup server
package srv

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/strs"

	"webtplmst/internal/conf"
	"webtplmst/internal/srv/adm"
	"webtplmst/internal/srv/internal"
	"webtplmst/internal/srv/std"
	"webtplmst/internal/srv/swg"
	"webtplmst/internal/srv/usr"
)

func Setup() {
	app := fiber.New(fiber.Config{
		AppName:      conf.App.Name,
		ErrorHandler: fext.ErrorHandler,
		BodyLimit:    conf.App.BodyLimit * 1024 * 1024,
	})
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: internal.AllowOriginsFunc,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
	}))
	app.Use("/", static.New(conf.App.RWeb, static.Config{
		Next: internal.NginxMiddleware,
	}))
	app.Use("/media", static.New(conf.App.RMedia, static.Config{
		Next: internal.NginxMiddleware,
	}))
	app.Get("/favicon.ico", static.New("./assets/favicon.ico"))
	std.Setup(app.Group("/api/v1").Use(internal.Log("GEN")))
	adm.Setup(app.Group("/adm/api/v1").Use(internal.Log("ADM")))
	usr.Setup(app.Group("/usr/api/v1").Use(internal.Log("USR")))
	swg.Setup(app.Group("/doc/api/v1").Use(internal.Log("DOC")))
	fext.MustListen(app, strs.ToStart(conf.App.Port, strs.Colon))
}
