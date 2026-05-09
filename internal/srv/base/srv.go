// Package base to setup base route
package base

import (
	"github.com/gofiber/fiber/v3"
)

func Setup(api fiber.Router) {
	api.Group("/rate").
		Get("/:code", FindRate)
	api.Group("/webhook").
		All("/wechat", WechatWebhook)
}
