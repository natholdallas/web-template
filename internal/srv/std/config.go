package std

import (
	"github.com/gofiber/fiber/v3"

	"webtplmst/internal/conf"
)

// Config is the public app configuration consumed by the frontend so it can
// build media URLs and render the available OAuth login providers.
type Config struct {
	Media     string      `json:"media"`        // media base (route path or full domain)
	MediaSize int         `json:"mediaMaxSize"` // max single upload size (MB)
	Oauth     OauthConfig `json:"oauth"`        // enabled third-party login providers
} //	@name	Config

type OauthConfig struct {
	Google   *GoogleOauth   `json:"google,omitempty"`
	Telegram *TelegramOauth `json:"telegram,omitempty"`
} //	@name	OauthConfig

type GoogleOauth struct {
	Enabled     bool   `json:"enabled"`
	ClientID    string `json:"clientId"`
	RedirectURI string `json:"redirectUri"`
} //	@name	GoogleOauth

type TelegramOauth struct {
	Enabled     bool   `json:"enabled"`
	BotUsername string `json:"botUsername"`
} //	@name	TelegramOauth

// FindConfig godoc
//
//	@Summary		Get app config
//	@Description	Public app configuration: media base URL and enabled OAuth providers
//	@Tags			Config
//	@ID				findConfig
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Config
//	@Router			/api/v1/config [get]
func FindConfig(c fiber.Ctx) error {
	v := Config{
		Media:     conf.App.MediaBase(),
		MediaSize: conf.App.MediaMaxSize,
	}
	if conf.App.OauthGoogleEnabled {
		v.Oauth.Google = &GoogleOauth{
			Enabled:     true,
			ClientID:    conf.App.OauthGoogleClientID,
			RedirectURI: conf.App.OauthGoogleRedirectURI,
		}
	}
	if conf.App.OauthTelegramEnabled {
		v.Oauth.Telegram = &TelegramOauth{
			Enabled:     true,
			BotUsername: conf.App.OauthTelegramBotName,
		}
	}
	return c.JSON(v)
}
