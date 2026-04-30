package client

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func GetGoogleUserInfo(token string) (GoogleUser, error) {
	result := GoogleUser{}
	_, err := client.R().
		SetQueryParam("access_token", token).
		SetResult(&result).
		Get("https://www.googleapis.com/oauth2/v1/userinfo")
	return result, err
}
