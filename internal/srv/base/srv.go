// Package base to setup base route
package base

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/natholdallas/natools4go/fext"
)

func Setup(api fiber.Router) {
	api.Use(logger.New(logger.Config{
		TimeFormat: time.DateTime,
		Format:     "[Base] " + fext.StdLogFmt,
	}))

	api.Group("/rate").
		Get("/:code", FindRate)
	api.Group("/webhook").
		All("/wechat", WechatWebhook)
}
