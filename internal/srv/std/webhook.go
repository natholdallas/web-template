package std

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/dbg"
	"webtplmst/internal/client"
	"webtplmst/internal/srv/internal"
)

func WechatWebhook(c fiber.Ctx) error {
	log.Info("wechat webhooking...")
	stdReq, err := internal.FasthttpToHTTP(c.RequestCtx())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed", "code": "200"})
	}
	// verify signature
	d, err := client.WxVerify(stdReq)
	if err != nil {
		log.Warn("wechat verify signature failed: ", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Failed", "code": "200"})
	}
	dbg.Dump(d)
	return nil
}
