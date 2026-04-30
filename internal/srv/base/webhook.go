package base

import "github.com/gofiber/fiber/v3"

func WechatWebhook(c fiber.Ctx) error {
	return c.Next()
}
