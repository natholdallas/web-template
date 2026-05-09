package base

import (
	"webtplmst/internal/client"
	"webtplmst/internal/srv/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/jsons"
)

func WechatWebhook(c fiber.Ctx) error {
	log.Info("wechat webhooking...")
	stdReq, err := internal.FasthttpToHTTP(c.RequestCtx())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	// verify signature
	d, err := client.WxVerify(stdReq)
	if err != nil {
		log.Warn("wechat verify signature failed: ", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Failed", "code": "200"})
	}
	// logging
	log.Debug(jsons.IString(d, true))

	// add your logic here
	return c.Next()
}
