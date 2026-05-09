package internal

import (
	"net/http"
	"time"

	"webtplmst/internal/conf"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/natholdallas/natools4go/fext"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func FasthttpToHTTP(context *fasthttp.RequestCtx) (*http.Request, error) {
	var v http.Request
	err := fasthttpadaptor.ConvertRequest(context, &v, true)
	return &v, err
}

func FastLogger(prefix string) fiber.Handler {
	return logger.New(logger.Config{
		TimeFormat: time.DateTime,
		Format:     "[" + prefix + "]" + fext.StdLogFmt,
		Stream:     conf.App.LogWriter(),
	})
}
