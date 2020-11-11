package auth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

// Create the JWT key used to create the signature
const jwtKey = "12345-67890-09876-54321"

const msgLoginFailed = "Login failed!"
const msgLoginSuccessful = "You are successfully logged in!"

const userIdNotFound int64 = -1
const costOfPwHash = 8

type UserInfo struct {
	ID        int64  `json:"_id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Admin     bool   `json:"admin"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	credentials
}

type checkJwtStatus struct {
	Status    string `json:"status"`
	IsSuccess bool   `json:"success"`
	ErrorMsg  string `json:"err"`
}

func (ui UserInfo) generatePasswordHash() ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(ui.Password), costOfPwHash)
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type passwordHash string

func (ph passwordHash) validateCredentials(creds credentials) error {
	return bcrypt.CompareHashAndPassword([]byte(ph), []byte(creds.Password))
}

type loginResult struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Status  string `json:"status"`
}

type signupResult struct {
	Status string `json:"status"`
	User   string `json:"user"`
}

type claims struct {
	UserId string `json:"_id"`
	Admin  bool   `json:"admin"`
	jwt.StandardClaims
}

func (claims *claims) generateTokenStringForUser() (string, error) {

	expirationTime := time.Now().Add(24 * time.Hour)
	// Create the JWT claims, which includes the username and expiry time
	claims.StandardClaims = jwt.StandardClaims{
		// In JWT, the expiry time is expressed as unix milliseconds
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	return tokenString, err
}

func GetLoginResult(userId int64, isAdmin bool, isSuccess bool) loginResult {

	var tokenString string
	var statusMsg = msgLoginFailed
	if isSuccess {
		statusMsg = msgLoginSuccessful
		userIdString := strconv.FormatInt(userId, 10)
		claims := &claims{UserId: userIdString, Admin: isAdmin}
		var err error
		tokenString, err = claims.generateTokenStringForUser()
		if err != nil {
			tokenString = ""
		}
	}

	result := loginResult{Success: isSuccess, Status: statusMsg, Token: tokenString}
	return result
}

func GetJwtTokenFromRequest(r *http.Request) (string, error) {

	// header is a string: KEY: Authorization VALUE: Bearer tokenstringInBase64
	tokenInHeaderVal := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	if len(tokenInHeaderVal) != 2 {
		return "", fmt.Errorf("Malformed token")
	}

	return tokenInHeaderVal[1], nil
}

func validateToken(jwtToken string) (claims, bool) {

	claims := &claims{}

	token, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})
	if err != nil {
		return *claims, false
	}

	if !token.Valid {
		return *claims, false
	}

	return *claims, true
}

func VerifyUser(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		token, err := GetJwtTokenFromRequest(r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		claims, ok := validateToken(token)
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// store the claims in the request as a context obj
		r = r.WithContext(context.WithValue(r.Context(), "claims", claims))
		next(w, r, ps)
	}
}

func VerifyAdmin(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		claims := GetClaimsFromRequest(r)
		if !claims.Admin {
			http.Error(w, "You are not authorized to perform this operation!", http.StatusUnauthorized)
			return
		}
		next(w, r, ps)
	}
}

func GetClaimsFromRequest(r *http.Request) claims {

	claims, _ := r.Context().Value("claims").(claims)
	return claims
}
