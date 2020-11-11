package auth

import (
	"encoding/json"
	"io"
	"net/http"

	"confusion.com/bwoo/cors"
	"confusion.com/bwoo/misc"

	"github.com/julienschmidt/httprouter"
)

func SetupRoutes(router *httprouter.Router) {

	// auth methods
	router.POST("/users/login", cors.Cors(login))
	router.POST("/users/signup", cors.Cors(signup))
	router.GET("/users", cors.Cors(VerifyUser(VerifyAdmin(getUsers))))
	router.GET("/users/checkJWTtoken", cors.Cors(checkJwtToken))
}

func writeCheckJwtErrorStatus(w http.ResponseWriter, statusMsg, errMsg string) {

	w.Header().Set("Content-Type", "application/json")
	jwtStatus := checkJwtStatus{IsSuccess: false, Status: statusMsg, ErrorMsg: errMsg}
	statusJson, _ := misc.GetJsonFromJsonObjs(jwtStatus)
	w.Write(statusJson)
	w.WriteHeader(http.StatusUnauthorized)
}

func checkJwtToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	jwtStr, err := GetJwtTokenFromRequest(r)
	if err != nil {
		writeCheckJwtErrorStatus(w, "JWT invalid!", "JWT invalid!")
		return
	}

	if _, ok := validateToken(jwtStr); !ok {
		writeCheckJwtErrorStatus(w, "JWT invalid!", "JWT invalid!")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jwtStatus := checkJwtStatus{IsSuccess: true, Status: "JWT valid!", ErrorMsg: "JWT valid!"}
	statusJson, _ := misc.GetJsonFromJsonObjs(jwtStatus)
	w.Write(statusJson)
	w.WriteHeader(http.StatusOK)
}

func getCredentialsFromBody(body io.ReadCloser) (credentials, error) {

	var creds credentials
	err := json.NewDecoder(body).Decode(&creds)
	if err != nil {
		return credentials{}, err
	}

	return creds, nil
}

func getUserInfoInfoFromBody(body io.ReadCloser) (UserInfo, error) {

	var signupInfo UserInfo
	err := json.NewDecoder(body).Decode(&signupInfo)
	if err != nil {
		return UserInfo{}, err
	}

	return signupInfo, nil
}

func signup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	signupInfo, err := getUserInfoInfoFromBody(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, ok := createUserInDb(signupInfo)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	signupResult := signupResult{Status: "Registration Successful!", User: signupInfo.Username}
	resultJson, _ := misc.GetJsonFromJsonObjs(signupResult)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJson)
}

func login(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	creds, err := getCredentialsFromBody(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	userId, isAdmin, isUserAuth := validateUserInDb(creds)
	loginResult := GetLoginResult(userId, isAdmin, isUserAuth)
	resultJson, _ := misc.GetJsonFromJsonObjs(loginResult)

	w.Header().Set("Content-Type", "application/json")
	if !isUserAuth {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(resultJson)
		return
	}

	// return jwt token
	w.WriteHeader(http.StatusOK)
	w.Write(resultJson)
}

func getUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userInfo, err := getUsersFromDb()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resultJson, _ := misc.GetJsonFromJsonObjs(userInfo)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resultJson)
}
