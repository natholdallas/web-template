// Package oauth implements third-party login (Google, Telegram) verification
// and authorization URL building backed by the app configuration.
package oauth

import (
	"errors"
	"net/url"

	"webtplmst/internal/client"
	"webtplmst/internal/conf"
)

// GoogleAuthURL builds the Google OAuth2 authorization URL. The state value
// carries frontend context (e.g. {"mode":"login"} or {"mode":"bind"}).
func GoogleAuthURL(state string) (string, error) {
	if !conf.App.OauthGoogleEnabled {
		return "", errors.New("google oauth not enabled")
	}
	if conf.App.OauthGoogleClientID == "" || conf.App.OauthGoogleRedirectURI == "" {
		return "", errors.New("google oauth not configured")
	}
	q := url.Values{}
	q.Set("client_id", conf.App.OauthGoogleClientID)
	q.Set("redirect_uri", conf.App.OauthGoogleRedirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("prompt", "select_account")
	if state != "" {
		q.Set("state", state)
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + q.Encode(), nil
}

// GoogleExchange trades an authorization code for the Google profile using the
// registered redirect URI from the configuration.
func GoogleExchange(code string) (client.GoogleUser, error) {
	if !conf.App.OauthGoogleEnabled {
		return client.GoogleUser{}, errors.New("google oauth not enabled")
	}
	token, err := client.GoogleToken(code, conf.App.OauthGoogleRedirectURI)
	if err != nil {
		return client.GoogleUser{}, err
	}
	return client.GoogleUserInfo(token)
}
