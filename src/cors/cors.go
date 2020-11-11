package cors

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var allowedOrigins = make(map[string]bool)

const allowedMethods = "GET, POST, PUT, DELETE, OPTIONS"

func setupAllowedOrigins() {
	allowedOrigins["http://localhost:3000"] = true
	allowedOrigins["https://localhost:3443"] = true
	allowedOrigins["http://localhost:4200"] = true
}

func SetupCors(router *httprouter.Router) {

	setupAllowedOrigins()
	setupDefaultHttpOptions(router)
}

func setupDefaultHttpOptions(router *httprouter.Router) {
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		wHeader := w.Header()
		setAccessControlAllowOriginIfValid(r.Header, wHeader)
		wHeader.Add("Access-Control-Allow-Credentials", "true")
		wHeader.Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		wHeader.Set("Access-Control-Allow-Methods", allowedMethods)

		// Adjust status code to 204
		w.WriteHeader(http.StatusNoContent)
	})
}

func CorsAllOrigin(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		wHeader := w.Header()
		wHeader.Add("Access-Control-Allow-Origin", "*")
		next(w, r, ps)
	}
}

func Cors(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		wHeader := w.Header()
		setAccessControlAllowOriginIfValid(r.Header, wHeader)
		wHeader.Add("Access-Control-Allow-Credentials", "true")
		wHeader.Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		wHeader.Add("Access-Control-Allow-Methods", allowedMethods)

		next(w, r, ps)
	}
}

func setAccessControlAllowOriginIfValid(rHeader http.Header, wHeader http.Header) {

	origin := rHeader.Get("Origin")
	if origin != "" {
		_, ok := allowedOrigins[origin]
		if ok {
			wHeader.Add("Access-Control-Allow-Origin", origin)
		}
	}
}
