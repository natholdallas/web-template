package conf

import (
	"fmt"
	"io"
	"os"

	"github.com/natefinch/lumberjack"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/spew"
	"github.com/natholdallas/natools4go/strs"
	"github.com/natholdallas/natools4go/vipers"

	"github.com/gofiber/fiber/v3"
	flog "github.com/gofiber/fiber/v3/log"
	glog "gorm.io/gorm/logger"
)

type AppConf struct {
	Name  string
	Port  string
	Debug bool
	Nginx bool

	// jwt
	SecretAdm string `validate:"required"`
	SecretUsr string `validate:"required"`

	// log
	LogLevelGorm  glog.LogLevel
	LogLevelFiber flog.Level

	// cors
	CorsDev []string
	CorsPrd []string

	// database
	DBName     string `validate:"required"`
	DBQuery    string `validate:"required"`
	DBPort     string `validate:"required"`
	DBHost     string `validate:"required"`
	DBUsername string `validate:"required"`
	DBPassword string `validate:"required"`

	// redis
	RedisHost  string
	RedisPort  string
	RedisIndex int

	// resources
	RWeb   string `validate:"required"`
	RLog   string `validate:"required"`
	RCache string `validate:"required"`
	RMedia string `validate:"required"`

	// smtp
	SMTPHost     string
	SMTPPort     int
	SMTPFrom     string
	SMTPPassword string
	SMTPAddr     string

	// wechat
	WxSite            string
	WxWebhook         string
	WxAppID           string
	WxSecret          string
	WxMch             string
	WxCert            string
	WxV3Sercret       string
	WxV2Sercret       string
	WxPubKey          string
	WxAPIClientKeyPem string
	WxPubKeyPem       string

	// exchangerate
	RateSite       string
	RateCurrencies []string
}

func (a *AppConf) DebugMiddleware(c fiber.Ctx) error {
	if a.Debug {
		return c.Next()
	}
	return &fext.Fail{Status: fiber.StatusForbidden}
}

func (a *AppConf) NginxMiddleware(c fiber.Ctx) bool {
	return a.Nginx
}

func (a *AppConf) AllowOriginsFunc(origin string) bool {
	if a.Debug {
		return strs.AnyPrefix(origin, a.CorsDev...)
	} else {
		return strs.AnyPrefix(origin, a.CorsPrd...)
	}
}

func (a *AppConf) LogWriter() io.Writer {
	return io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   a.RLog + "/app.log",
		MaxSize:    10,
		MaxBackups: 7,
		MaxAge:     28,
		Compress:   true,
	})
}

func (a *AppConf) MkdirAll() {
	os.MkdirAll(a.RMedia, 0o777)
	os.MkdirAll(a.RLog, 0o777)
}

func LoadApp() {
	vipers.Config(Flag.ConfName, Flag.ConfPath)
	App.Name = vipers.Get("app.name", "app")
	App.Port = vipers.Get("app.port", "8080")
	App.Debug = vipers.Get("app.debug", false)
	App.Nginx = vipers.Get("app.nginx", false)
	App.SecretAdm = vipers.String("secret.adm")
	App.SecretUsr = vipers.String("secret.usr")
	App.LogLevelFiber = flog.Level(vipers.Get("loglevel.fiber", int(flog.LevelTrace)))
	App.LogLevelGorm = glog.LogLevel(vipers.Get("loglevel.gorm", int(glog.Warn)))
	App.CorsDev = vipers.StringSlice("cors.dev")
	App.CorsPrd = vipers.StringSlice("cors.prd")
	App.RWeb = strs.TrimEnd(vipers.Get("resources.web", "web"), strs.Slash)
	App.RLog = strs.TrimEnd(vipers.Get("resources.log", "log"), strs.Slash)
	App.RCache = strs.TrimEnd(vipers.String("resources.cache"), strs.Slash)
	App.RMedia = strs.TrimEnd(vipers.Get("resources.media", "media"), strs.Slash)
	App.SMTPHost = vipers.String("smtp.host")
	App.SMTPPort = vipers.Int("smtp.port")
	App.SMTPFrom = vipers.String("smtp.from")
	App.SMTPPassword = vipers.String("smtp.password")
	App.SMTPAddr = fmt.Sprintf("%s:%d", App.SMTPHost, App.SMTPPort)
	App.DBName = vipers.String("db.name")
	App.DBPort = vipers.String("db.port")
	App.DBQuery = vipers.String("db.query")
	App.DBHost = vipers.String("db.host")
	App.DBUsername = vipers.String("db.username")
	App.DBPassword = vipers.String("db.password")
	App.RedisHost = vipers.Get("redis.host", "localhost")
	App.RedisPort = vipers.Get("redis.port", "6379")
	App.RedisIndex = vipers.Get("redis.index", 0)
	App.WxSite = strs.TrimStart(vipers.String("wechat.site"), strs.Slash)
	App.WxAppID = vipers.String("wechat.appid")
	App.WxSecret = vipers.String("wechat.secret")
	App.WxWebhook = vipers.String("wechat.pay.webhook")
	App.WxMch = vipers.String("wechat.pay.mch")
	App.WxCert = vipers.String("wechat.pay.cert")
	App.WxV2Sercret = vipers.String("wechat.pay.apiv2secret")
	App.WxV3Sercret = vipers.String("wechat.pay.apiv3secret")
	App.WxPubKey = vipers.String("wechat.pay.public-key")
	App.WxAPIClientKeyPem = vipers.String("wechat.pem.apiclient")
	App.WxPubKeyPem = vipers.String("wechat.pem.pub")
	App.RateSite = strs.TrimEnd(vipers.String("exchangerate.site"), strs.Slash)
	App.RateCurrencies = vipers.StringSlice("exchangerate.currencies")

	// xdg support
	if dir, err := os.UserCacheDir(); err == nil {
		App.RCache = dir + strs.ToStart(App.RCache, strs.Slash)
	}

	// mkdir & validate
	App.MkdirAll()
	vipers.Validate(App)

	// init
	spew.SetPrinter(flog.Debugf)
	flog.SetLevel(App.LogLevelFiber)
	flog.SetOutput(App.LogWriter())
	fext.SetDebugMode(App.Debug)
	fext.SetErrorFunc(func(err error) { flog.Error(err) })
}
