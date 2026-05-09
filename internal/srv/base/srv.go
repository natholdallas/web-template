// Package base to setup base route
package base

import (
	"webtplmst/internal/srv/internal"

	"github.com/gofiber/fiber/v3"
)

func Setup(api fiber.Router) {
	api.Use(internal.FastLogger("Base"))
	api.Group("/rate").
		Get("/:code", FindRate)
	api.Group("/webhook").
		All("/wechat", WechatWebhook)
}
