package oauth2

import (
	"encoding/json"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

const userIdNotFound int64 = -1

type FacebookUserInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

func getFacebookUserInfoFromBody(body io.ReadCloser) (FacebookUserInfo, error) {

	var userInfo FacebookUserInfo
	err := json.NewDecoder(body).Decode(&userInfo)
	if err != nil {
		return FacebookUserInfo{}, err
	}

	return userInfo, nil
}

func GetOauthFbConfig(clientID, clientSecret, redirectUrl string) *oauth2.Config {

	var oauthConf = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Scopes:       []string{"public_profile"},
		Endpoint:     facebook.Endpoint,
	}

	return oauthConf
}
