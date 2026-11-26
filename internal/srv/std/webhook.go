package std

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"

	"webgx/internal/client"
	"webgx/internal/srv/internal"
)

// WechatWebhook receives WeChat Pay transaction notifications. The body is
// never logged: it carries payer identifiers and payment amounts.
func WechatWebhook(c fiber.Ctx) error {
	stdReq, err := internal.FasthttpToHTTP(c.RequestCtx())
	if err != nil {
		log.Warn("convert wechat notify request: ", err)
		return invalidNotify(c)
	}
	d, err := client.WxVerify(stdReq)
	if err != nil {
		log.Warn("wechat verify signature failed: ", err)
		return invalidNotify(c)
	}
	if d.OutTradeNo != nil {
		log.Info("wechat notify: ", *d.OutTradeNo)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// invalidNotify returns a generic failure so internal details never reach the
// caller.
func invalidNotify(c fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "FAIL", "message": "invalid notify"})
}
