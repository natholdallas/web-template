// Package std registers the public (unauthenticated) API routes.
package std

import (
	"github.com/gofiber/fiber/v3"
)

func Setup(api fiber.Router) {
	api.Group("/rate").
		Get("/:code", FindRate)
	api.Group("/webhook").
		All("/wechat", WechatWebhook)
	api.Group("/config").
		Get("", FindConfig)
}
