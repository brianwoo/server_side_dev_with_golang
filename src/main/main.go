package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"confusion.com/bwoo/favoriteDishes"

	"confusion.com/bwoo/config"
	"confusion.com/bwoo/cors"
	"confusion.com/bwoo/misc"
	"confusion.com/bwoo/oauth2"

	"confusion.com/bwoo/upload"

	"confusion.com/bwoo/auth"
	"confusion.com/bwoo/comments"
	"confusion.com/bwoo/database"
	"confusion.com/bwoo/dishes"
	"confusion.com/bwoo/leaders"
	"confusion.com/bwoo/promotions"
	"github.com/julienschmidt/httprouter"

	_ "github.com/go-sql-driver/mysql"
)

const serverListenPort = "0.0.0.0:3000"
const sslPort = ":3443"
const serverListenSslPort = "0.0.0.0" + sslPort
const certPath = "../../certs/www.confusion.com.crt"
const keyPath = "../../certs/www.confusion.com.key"

func getIndex(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprint(w, "Welcome to ConFusion!\n")
}

func setupDefaultRoutes(router *httprouter.Router) {
	router.GET("/", getIndex)
}

func redirectToSecurePort(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	host := strings.Split(req.Host, ":")[0]
	target := "https://" + host + sslPort + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}

func listenOnInsecurePortAndRedirect() {

	log.Println("Server starting. Listening on " + serverListenPort)
	// redirect every http request to https
	go http.ListenAndServe(serverListenPort, http.HandlerFunc(redirectToSecurePort))
}

func listenOnSecurePort(router *httprouter.Router) {

	log.Println("Server starting. Listening on " + serverListenSslPort)

	// start server on https port
	server := http.Server{
		Addr:    serverListenSslPort,
		Handler: router,
		TLSConfig: &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
		},
	}

	certFilePath, _ := filepath.Abs(certPath)
	keyFilePath, _ := filepath.Abs(keyPath)
	err := server.ListenAndServeTLS(certFilePath, keyFilePath)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	configFilePath := misc.GetConfigFilePath()
	config := config.ReadDbConfig(configFilePath)

	database.SetupDatabase(config)

	router := httprouter.New()
	cors.SetupCors(router)
	dishes.SetupRoutes(router)
	comments.SetupRoutes(router)
	leaders.SetupRoutes(router)
	promotions.SetupRoutes(router)
	auth.SetupRoutes(router)
	upload.SetupRoutes(router, config)
	oauth2.SetupRoutes(router, config)
	favoriteDishes.SetupRoutes(router)
	setupDefaultRoutes(router)

	listenOnInsecurePortAndRedirect()
	listenOnSecurePort(router)

}
