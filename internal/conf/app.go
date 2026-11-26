package conf

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	flog "github.com/gofiber/fiber/v3/log"
	"github.com/natefinch/lumberjack"
	"github.com/natholdallas/natools4go/narder"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/strs"
	"github.com/natholdallas/natools4go/vipers"
	glog "gorm.io/gorm/logger"
)

// SMTP connection modes accepted by smtp.tls.
const (
	SMTPTLSAuto     = "auto"
	SMTPTLSStartTLS = "starttls"
	SMTPTLSSSL      = "ssl"
	SMTPTLSNone     = "none"
)

type AppConf struct {
	// app
	Name      string
	Port      string
	Debug     bool
	Nginx     bool
	Swagger   bool
	BodyLimit int

	// jwt
	AccessMinutes int    `validate:"required"`
	RefreshHours  int    `validate:"required"`
	SecretAdm     string `validate:"required,len=32"`
	SecretUsr     string `validate:"required,len=32"`

	// log
	LogFilename   string `validate:"required"`
	LogMaxSize    int    `validate:"required"`
	LogMaxBacks   int
	LogMaxAge     int
	LogCompress   bool
	LogLocalTime  bool
	LogLevelGorm  glog.LogLevel
	LogLevelFiber flog.Level

	// cors
	CorsDev []string
	CorsPrd []string

	// database
	// Required-ness is driver-dependent, so only the universally required fields
	// carry a tag here; Validate checks host/credentials for server-based drivers.
	DBName        string `validate:"required"`
	DBQuery       string
	DBPort        string
	DBHost        string
	DBUsername    string
	DBPassword    string
	DBAutoMigrate bool
	DBAutoCreate  bool
	DBDriver      string `validate:"required"`
	DBDsn         string `validate:"required"`

	// redis
	RedisHost  string `validate:"required"`
	RedisPort  string `validate:"required"`
	RedisIndex int
	RedisAddr  string `validate:"required"`

	// resources
	RWeb          string `validate:"required"`
	RLog          string `validate:"required"`
	RCache        string `validate:"required"`
	RMedia        string `validate:"required"`
	MediaRoute    string
	MediaURL      string
	MediaMaxSize  int
	AvatarMaxSize int

	// oauth (google / telegram login)
	OauthGoogleEnabled      bool
	OauthGoogleClientID     string
	OauthGoogleClientSecret string
	OauthGoogleRedirectURI  string
	OauthTelegramEnabled    bool
	OauthTelegramBotToken   string
	OauthTelegramBotName    string

	// smtp
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPFrom     string
	SMTPPassword string
	SMTPAddr     string
	SMTPTLS      string

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

	// request client
	RestyInsecureSkipVerify bool
}

// logWriter is shared by every logger (fiber, narder, gorm). Building one
// lumberjack.Logger per call made several writers race over the same file
// rotation; the first call wins and the rest reuse it.
var (
	logWriterOnce sync.Once
	logWriter     io.Writer
)

func (a *AppConf) LogWriter() io.Writer {
	logWriterOnce.Do(func() {
		logWriter = io.MultiWriter(os.Stdout, &lumberjack.Logger{
			Filename:   filepath.Join(a.RLog, a.LogFilename),
			MaxSize:    a.LogMaxSize,
			MaxBackups: a.LogMaxBacks,
			MaxAge:     a.LogMaxAge,
			Compress:   a.LogCompress,
			LocalTime:  a.LogLocalTime,
		})
	})
	return logWriter
}

// MkdirAll creates the writable resource directories. Media is served
// statically (0755); logs and cache stay private (0750). Chmod runs even when
// the directory already exists so an inherited 0777 is tightened.
func (a *AppConf) MkdirAll() error {
	dirs := []struct {
		path string
		mode os.FileMode
	}{
		{a.RMedia, 0o755},
		{a.RLog, 0o750},
		{a.RCache, 0o750},
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d.path, d.mode); err != nil {
			return fmt.Errorf("create %s: %w", d.path, err)
		}
		if err := os.Chmod(d.path, d.mode); err != nil {
			return fmt.Errorf("chmod %s: %w", d.path, err)
		}
	}
	return nil
}

// MediaBase returns the public media base URL: the configured media domain
// (e.g. https://media.example.com) when set, otherwise the same-origin
// /media route path. Used to build absolute media URLs for the frontend.
func (a *AppConf) MediaBase() string {
	if a.MediaURL != "" {
		return a.MediaURL
	}
	return a.MediaRoute
}

func LoadApp() {
	vipers.MustConfig(Flag.ConfName, Flag.ConfPath, Flag.ConfType)

	// app
	App.Name = vipers.Get("app.name", "app")
	App.Port = vipers.Get("app.port", "8080")
	App.Debug = vipers.Get("app.debug", false)
	App.Nginx = vipers.Get("app.nginx", false)
	App.BodyLimit = vipers.Get("app.body-limit", 200)
	App.Swagger = vipers.Get("app.swagger", false)

	// jwt
	App.AccessMinutes = vipers.Get("jwt.access-minutes", 15)
	App.RefreshHours = vipers.Get("jwt.refresh-hours", 720)
	App.SecretAdm = vipers.String("jwt.secret.adm")
	App.SecretUsr = vipers.String("jwt.secret.usr")

	// log
	App.LogFilename = vipers.Get("log.filename", "app.log")
	App.LogMaxSize = vipers.Get("log.max-size", 10)
	App.LogMaxBacks = vipers.Get("log.max-backups", 7)
	App.LogMaxAge = vipers.Get("log.max-age", 28)
	App.LogCompress = vipers.Get("log.compress", true)
	App.LogLocalTime = vipers.Get("log.local-time", false)
	App.LogLevelFiber = flog.Level(vipers.Get("loglevel.fiber", int(flog.LevelTrace)))
	App.LogLevelGorm = glog.LogLevel(vipers.Get("loglevel.gorm", int(glog.Warn)))

	// cors
	App.CorsDev = vipers.StringSlice("cors.dev")
	App.CorsPrd = vipers.StringSlice("cors.prd")

	// resources
	App.RWeb = strs.TrimEnd(vipers.String("resources.web"), strs.Slash)
	App.RLog = strs.TrimEnd(vipers.Get("resources.log", "log"), strs.Slash)
	App.RCache = strs.TrimEnd(vipers.String("resources.cache"), strs.Slash)
	App.RMedia = strs.TrimEnd(vipers.Get("resources.media", "media"), strs.Slash)
	App.MediaRoute = strs.TrimEnd(vipers.Get("resources.media-route", "/media"), strs.Slash)
	App.MediaURL = strs.TrimEnd(vipers.Get("resources.media-url", ""), strs.Slash)
	App.MediaMaxSize = vipers.Get("resources.media-max-size", 10)
	App.AvatarMaxSize = vipers.Get("resources.media-avatar-max-size", 2)

	// oauth
	App.OauthGoogleEnabled = vipers.Get("oauth.google.enabled", false)
	App.OauthGoogleClientID = vipers.String("oauth.google.client-id")
	App.OauthGoogleClientSecret = vipers.String("oauth.google.client-secret")
	App.OauthGoogleRedirectURI = vipers.String("oauth.google.redirect-uri")
	App.OauthTelegramEnabled = vipers.Get("oauth.telegram.enabled", false)
	App.OauthTelegramBotToken = vipers.String("oauth.telegram.bot-token")
	App.OauthTelegramBotName = vipers.String("oauth.telegram.bot-username")

	// smtp
	App.SMTPHost = vipers.String("smtp.host")
	App.SMTPPort = vipers.Int("smtp.port")
	App.SMTPUsername = vipers.String("smtp.username")
	App.SMTPFrom = vipers.String("smtp.from")
	App.SMTPPassword = vipers.String("smtp.password")
	App.SMTPAddr = fmt.Sprintf("%s:%d", App.SMTPHost, App.SMTPPort)
	App.SMTPTLS = strings.ToLower(strings.TrimSpace(vipers.Get("smtp.tls", SMTPTLSAuto)))

	// db
	App.DBName = vipers.String("db.name")
	App.DBPort = vipers.String("db.port")
	App.DBQuery = vipers.String("db.query")
	App.DBHost = vipers.String("db.host")
	App.DBUsername = vipers.String("db.username")
	App.DBPassword = vipers.String("db.password")
	App.DBAutoMigrate = vipers.Get("db.auto-migrate", false)
	App.DBAutoCreate = vipers.Get("db.auto-create", false)
	App.DBDriver = vipers.Get("db.driver", "mysql")
	App.DBDsn = orms.MustDSN(App.DBDriver, App.DBName, App.DBUsername, App.DBPassword, App.DBHost, App.DBPort)

	// redis
	App.RedisHost = vipers.Get("redis.host", "localhost")
	App.RedisPort = vipers.Get("redis.port", "6379")
	App.RedisIndex = vipers.Get("redis.index", 0)
	App.RedisAddr = net.JoinHostPort(App.RedisHost, App.RedisPort)

	// wechat
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

	// rate
	App.RateSite = strs.TrimEnd(vipers.String("exchangerate.site"), strs.Slash)
	App.RateCurrencies = vipers.StringSlice("exchangerate.currencies")

	// resty
	App.RestyInsecureSkipVerify = vipers.Get("resty.insecure-skip-verify", false)

	// xdg support
	if dir, err := os.UserCacheDir(); err == nil {
		App.RCache = filepath.Join(dir, App.RCache)
	}

	// mkdir
	if err := App.MkdirAll(); err != nil {
		panic(err)
	}

	// init
	flog.SetLevel(App.LogLevelFiber)
	flog.SetOutput(App.LogWriter())
	narder.SetDebugMode(App.Debug)
	narder.SetOutput(App.LogWriter())
}
