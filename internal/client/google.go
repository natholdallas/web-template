package client

import (
	"errors"
	"fmt"

	"webtplmst/internal/conf"
)

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func GoogleUserInfo(token string) (GoogleUser, error) {
	result := GoogleUser{}
	_, err := client.R().
		SetQueryParam("access_token", token).
		SetResult(&result).
		Get("https://www.googleapis.com/oauth2/v1/userinfo")
	return result, err
}

type googleTokenRes struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// GoogleToken exchanges an OAuth authorization code for an access token using
// the configured Google client credentials and the registered redirect URI.
func GoogleToken(code, redirectURI string) (string, error) {
	if conf.App.OauthGoogleClientID == "" || conf.App.OauthGoogleClientSecret == "" {
		return "", errors.New("google oauth not configured")
	}
	resp := googleTokenRes{}
	r, err := client.R().
		SetFormData(map[string]string{
			"code":          code,
			"client_id":     conf.App.OauthGoogleClientID,
			"client_secret": conf.App.OauthGoogleClientSecret,
			"redirect_uri":  redirectURI,
			"grant_type":    "authorization_code",
		}).
		SetResult(&resp).
		Post("https://oauth2.googleapis.com/token")
	if err != nil {
		return "", err
	}
	if r.IsStatusFailure() {
		return "", fmt.Errorf("google token error: %s %s", resp.Error, resp.ErrorDesc)
	}
	if resp.AccessToken == "" {
		return "", errors.New("google token response missing access token")
	}
	return resp.AccessToken, nil
}
