package oauth2

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"confusion.com/bwoo/auth"
	"confusion.com/bwoo/config"
	"confusion.com/bwoo/cors"
	"confusion.com/bwoo/misc"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
)

var facebookConfig *oauth2.Config

func SetupRoutes(router *httprouter.Router, config config.Config) {

	facebookConfig = GetOauthFbConfig(config.Oauth2FbClientID,
		config.Oauth2FbClientSecret,
		config.Oauth2FbRedirectUrl)

	// facebook OAuth related
	router.GET("/facebook/login", cors.Cors(loginFacebook))
	router.GET("/facebook/callback", cors.Cors(facebookLoginCallback))
	router.GET("/facebook/token", cors.Cors(loginWithFacebookToken))

}

// This requires the browser to call by the user.
// When the user successfully logged in, facebookLoginCallback() will be called
func loginFacebook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	oauthState := generateStateOauthCookie(w)
	url := facebookConfig.AuthCodeURL(oauthState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func facebookLoginCallback(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// Read oauthState from Cookie
	stateOauthCookie, err := r.Cookie("oauthstate")
	if err != nil {
		fmt.Println("Cannot get state token", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if stateOauthCookie != nil && r.FormValue("state") != stateOauthCookie.Value {
		log.Println("invalid OAuth state")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := facebookConfig.Exchange(oauth2.NoContext, r.FormValue("code"))
	if err != nil {
		fmt.Printf("facebookConfig.Exchange() failed with '%s'\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fmt.Println("Facebook Access Token:", token.AccessToken)
}

func getAccessToken(r *http.Request, ps httprouter.Params) (string, error) {

	// 3 ways the access_token can be passed in like the passport-facebook-token module:
	// 1. As a request param named access_token
	// 2. As a header param named access_token
	// 3. As a Bearer token in the header
	queryValues := r.URL.Query()
	if accessToken := queryValues.Get("access_token"); accessToken != "" {
		return accessToken, nil

	} else if accessToken := r.Header.Get("access_token"); accessToken != "" {
		return accessToken, nil

	} else if accessToken, err := auth.GetJwtTokenFromRequest(r); accessToken != "" && err == nil {
		return accessToken, nil
	}

	return "", fmt.Errorf("AccessToken NOT found")
}

func loginWithFacebookToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	accessToken, err := getAccessToken(r, ps)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp, err := http.Get("https://graph.facebook.com/me?fields=id,name,first_name,last_name,email&access_token=" +
		url.QueryEscape(accessToken))
	if resp.StatusCode >= http.StatusBadRequest {
		fmt.Println("Unable to login to Facebook")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err != nil {
		fmt.Printf("Facebook Graph API error: %s\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	defer resp.Body.Close()

	facebookUserInfo, err := getFacebookUserInfoFromBody(resp.Body)
	if err != nil {
		fmt.Printf("ReadAll: %s\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userInfo, err := findUserByFacebookIdCreateIfNotFound(facebookUserInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	loginResult := auth.GetLoginResult(userInfo.ID, userInfo.Admin, true)
	resultJson, _ := misc.GetJsonFromJsonObjs(loginResult)

	w.Header().Set("Content-Type", "application/json")

	// return jwt token
	w.WriteHeader(http.StatusOK)
	w.Write(resultJson)
}

func findUserByFacebookIdCreateIfNotFound(facebookUserInfo FacebookUserInfo) (*auth.UserInfo, error) {

	userInfo, err := getFacebookUserFromDb(facebookUserInfo.ID)
	if err != nil {
		return nil, err
	} else if userInfo == nil {
		_, ok := createUserInDb(facebookUserInfo)
		if !ok {
			return nil, fmt.Errorf("Unable to create user")
		}
		userInfo, _ = getFacebookUserFromDb(facebookUserInfo.ID)
	}

	return userInfo, nil
}
