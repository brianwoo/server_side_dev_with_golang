package favoriteDishes

import (
	"encoding/json"
	"io"
	"net/http"

	"confusion.com/bwoo/auth"
	"confusion.com/bwoo/cors"
	"confusion.com/bwoo/misc"
	"github.com/julienschmidt/httprouter"
)

func SetupRoutes(router *httprouter.Router) {

	// /favorites
	router.GET("/favorites", cors.Cors(auth.VerifyUser(getFavoriteDishes)))
	router.POST("/favorites", cors.Cors(auth.VerifyUser(postFavoriteDishes)))
	router.DELETE("/favorites", cors.Cors(auth.VerifyUser(deleteFavoriteDishes)))

	// /favorites/:dishId
	router.GET("/favorites/:dishId", cors.Cors(auth.VerifyUser(getFavoriteDish)))
	router.POST("/favorites/:dishId", cors.Cors(auth.VerifyUser(postFavoriteDish)))
	router.DELETE("/favorites/:dishId", cors.Cors(auth.VerifyUser(deleteFavoriteDish)))
}

/****************************
* Helper functions
****************************/
func getFavoriteDishesFromBody(body io.ReadCloser) (favoriteDishes, error) {

	var favDishes favoriteDishes
	err := json.NewDecoder(body).Decode(&favDishes)
	if err != nil {
		return favoriteDishes{}, err
	}

	return favDishes, nil
}

func getFavoriteDishesAndReply(w http.ResponseWriter, userId int64) {

	favDishes, err := getFavoriteDishesFromDb(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// NOTE: in NodeJS, it returns the userid and all the fav dishes
	favDishesJson, _ := misc.GetJsonFromJsonObjs(favDishes)
	w.Header().Set("Content-Type", "application/json")
	w.Write(favDishesJson)
}

/****************************
* /favorites operations
****************************/
func getFavoriteDishes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	claims := auth.GetClaimsFromRequest(r)
	userId, _ := misc.GetInt64FromString(claims.UserId)

	getFavoriteDishesAndReply(w, userId)
}

func postFavoriteDishes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	favDishes, err := getFavoriteDishesFromBody(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := auth.GetClaimsFromRequest(r)
	userId, _ := misc.GetInt64FromString(claims.UserId)

	_, err = createFavoriteDishesInDb(userId, favDishes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	getFavoriteDishesAndReply(w, userId)

	// NOTE: in NodeJS, it returns the userid and all the fav dishes
	// statusJson, _ := misc.GetJsonFromJsonObjs(status)
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(statusJson)
}

func deleteFavoriteDishes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	claims := auth.GetClaimsFromRequest(r)
	userId, _ := misc.GetInt64FromString(claims.UserId)

	_, err := deleteFavoriteDishesFromDb(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	getFavoriteDishesAndReply(w, userId)

	// NOTE: in NodeJS, it returns the userid and all the fav dishes deleted
	// statusJson, _ := misc.GetJsonFromJsonObjs(status)
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(statusJson)
}

/****************************
* /favorites/:dishId operations
****************************/
func getFavoriteDish(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	dishIdInt, err := misc.GetInt64FromString(dishId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := auth.GetClaimsFromRequest(r)
	userId, _ := misc.GetInt64FromString(claims.UserId)

	status, err := getFavoriteDishFromDb(userId, dishIdInt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// NOTE: in NodeJS, it returns the userid and all the fav dishes
	statusJson, _ := misc.GetJsonFromJsonObjs(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(statusJson)
}

func postFavoriteDish(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	dishIdInt, err := misc.GetInt64FromString(dishId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := auth.GetClaimsFromRequest(r)
	userId, _ := misc.GetInt64FromString(claims.UserId)

	_, err = createFavoriteDishInDb(userId, dishIdInt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	getFavoriteDishesAndReply(w, userId)
	// NOTE: in NodeJS, it returns the userid and all the fav dishes deleted
	// statusJson, _ := misc.GetJsonFromJsonObjs(status)
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(statusJson)
}

func deleteFavoriteDish(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	dishIdInt, err := misc.GetInt64FromString(dishId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := auth.GetClaimsFromRequest(r)
	userId, _ := misc.GetInt64FromString(claims.UserId)

	_, err = deleteFavoriteDishFromDb(userId, dishIdInt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	getFavoriteDishesAndReply(w, userId)

	// NOTE: in NodeJS, it returns the userid and all the fav dish deleted
	// statusJson, _ := misc.GetJsonFromJsonObjs(status)
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(statusJson)
}
