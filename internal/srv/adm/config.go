package adm

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"webgx/internal/conf"
)

// ConfCheck is one read-only configuration probe. Enabled reports whether the
// feature is switched on; Configured reports whether every value it needs is
// present.
type ConfCheck struct {
	Key        string   `json:"key"`
	Group      string   `json:"group"`
	Enabled    bool     `json:"enabled"`
	Configured bool     `json:"configured"`
	Missing    []string `json:"missing"`
} //	@name	AdmConfCheck

type ConfStatus struct {
	Checks []ConfCheck `json:"checks"`
} //	@name	AdmConfStatus

// confValue is one configuration key and whether it currently holds a value.
type confValue struct {
	key     string
	present bool
}

// present reports whether a configured string carries an actual value.
func present(v string) bool {
	return strings.TrimSpace(v) != ""
}

func missingValues(values ...confValue) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if !v.present {
			out = append(out, v.key)
		}
	}
	return out
}

func newCheck(key, group string, enabled bool, values ...confValue) ConfCheck {
	missing := missingValues(values...)
	return ConfCheck{
		Key:        key,
		Group:      group,
		Enabled:    enabled,
		Configured: len(missing) == 0,
		Missing:    missing,
	}
}

// configChecks probes every optional feature in display order.
func configChecks() []ConfCheck {
	return []ConfCheck{
		newCheck("app.nginx", "app", conf.App.Nginx),
		newCheck("app.swagger", "app", conf.App.Swagger),
		newCheck("app.debug", "app", conf.App.Debug),
		newCheck("security.jwt", "security", true,
			confValue{"jwt.secret.adm", len(conf.App.SecretAdm) == 32},
			confValue{"jwt.secret.usr", len(conf.App.SecretUsr) == 32},
		),
		newCheck("oauth.google", "oauth", conf.App.OauthGoogleEnabled,
			confValue{"oauth.google.client-id", present(conf.App.OauthGoogleClientID)},
			confValue{"oauth.google.client-secret", present(conf.App.OauthGoogleClientSecret)},
			confValue{"oauth.google.redirect-uri", present(conf.App.OauthGoogleRedirectURI)},
		),
		newCheck("oauth.telegram", "oauth", conf.App.OauthTelegramEnabled,
			confValue{"oauth.telegram.bot-token", present(conf.App.OauthTelegramBotToken)},
			confValue{"oauth.telegram.bot-username", present(conf.App.OauthTelegramBotName)},
		),
		newCheck("smtp", "mail", present(conf.App.SMTPHost) || present(conf.App.SMTPFrom),
			confValue{"smtp.host", present(conf.App.SMTPHost)},
			confValue{"smtp.port", conf.App.SMTPPort != 0},
			confValue{"smtp.from", present(conf.App.SMTPFrom)},
			confValue{"smtp.password", present(conf.App.SMTPPassword)},
		),
		newCheck("wechat", "wechat", present(conf.App.WxAppID) || present(conf.App.WxSecret),
			confValue{"wechat.appid", present(conf.App.WxAppID)},
			confValue{"wechat.secret", present(conf.App.WxSecret)},
		),
		newCheck("wechat.pay", "wechat",
			present(conf.App.WxMch) || present(conf.App.WxCert) || present(conf.App.WxV3Sercret),
			confValue{"wechat.pay.mch", present(conf.App.WxMch)},
			confValue{"wechat.pay.cert", present(conf.App.WxCert)},
			confValue{"wechat.pay.apiv3secret", present(conf.App.WxV3Sercret)},
			confValue{"wechat.pay.public-key", present(conf.App.WxPubKey)},
			confValue{"wechat.pem.apiclient", present(conf.App.WxAPIClientKeyPem)},
			confValue{"wechat.pem.pub", present(conf.App.WxPubKeyPem)},
		),
		newCheck("exchangerate", "rate", rateEnabled(),
			confValue{"exchangerate.site", present(conf.App.RateSite)},
			confValue{"exchangerate.currencies", len(conf.App.RateCurrencies) > 0},
		),
		newCheck("media.cdn", "media", present(conf.App.MediaURL)),
	}
}

// GetConfigStatus godoc
//
//	@Summary		Read configuration status
//	@Description	Reports every optional feature and the configuration keys it is missing
//	@Tags			admConfig
//	@ID				admGetConfigStatus
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	ConfStatus
//	@Router			/adm/api/v1/config/status [get]
func GetConfigStatus(c fiber.Ctx) error {
	return c.JSON(ConfStatus{Checks: configChecks()})
}
