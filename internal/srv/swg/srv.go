package swg

import (
	"webtplmst/internal/conf"
	"webtplmst/internal/srv/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/yokeTH/gofiber-scalar/scalar/v3"
)

func Setup(app fiber.Router) {
	if !conf.App.Swagger {
		return
	}
	app.Get("/*", internal.SwaggerMiddleware, scalar.New(scalar.Config{
		Title:             "API Documentation",
		BasePath:          "/doc/api/v1",
		Theme:             scalar.ThemeSaturn,
		FileContentString: displayDoc(),
	}))
}
