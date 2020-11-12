package dishes

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"confusion.com/bwoo/auth"
	"confusion.com/bwoo/cors"
	"confusion.com/bwoo/misc"
	"github.com/julienschmidt/httprouter"
)

func SetupRoutes(router *httprouter.Router) {

	// dish
	router.GET("/dishes/:dishId", cors.CorsAllOrigin(getDish))
	router.PUT("/dishes/:dishId", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(putDish))))
	router.POST("/dishes/:dishId", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(postDish))))
	router.DELETE("/dishes/:dishId", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(deleteDish))))

	// dishes
	router.GET("/dishes", cors.CorsAllOrigin(getDishes))
	router.PUT("/dishes", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(putDishes))))
	router.POST("/dishes", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(postDishes))))
	router.DELETE("/dishes", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(deleteDishes))))
}

/****************************
* Helper functions
****************************/
func getDishFromBody(body io.ReadCloser) (Dish, error) {

	var dish Dish
	err := json.NewDecoder(body).Decode(&dish)
	if err != nil {
		return Dish{}, err
	}

	return dish, nil
}

/****************************
* Dish operations
****************************/
func getDish(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	dishIdInt, err := misc.GetInt64FromString(dishId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dish, err := getDishFromDb(dishIdInt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var jsonDish []byte = make([]byte, 0)
	if dish == nil {
		jsonDish, err = misc.GetJsonFromJsonObjs(struct{}{})
	} else {
		jsonDish, err = misc.GetJsonFromJsonObjs(dish)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonDish)
}

func putDish(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	dishIdInt, err := misc.GetInt64FromString(dishId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dish, err := getDishFromBody(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updatedDish, err := updateDishFromDb(dishIdInt, dish)
	if err != nil && updatedDish == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// if update was run but no rows updated (i.e. updatedDish.ID = 0),
	// we will just return an empty json object
	var updatedDishJson []byte
	if updatedDish.ID == 0 {
		updatedDishJson = misc.GetEmptyJsonByteArray()
	} else {
		updatedDishJson, _ = misc.GetJsonFromJsonObjs(updatedDish)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(updatedDishJson))
}

func postDish(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("POST operation not supported on /dishes/" + dishId))
}

func deleteDish(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	dishIdInt, err := misc.GetInt64FromString(dishId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status, err := deleteDishFromDb(dishIdInt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	statusJson, err := misc.GetJsonFromJsonObjs(status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(statusJson))
}

/****************************
* Dishes operations
****************************/
func getDishes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	isFeaturedStr := r.URL.Query().Get("featured")
	isFeatured, err := strconv.ParseBool(isFeaturedStr)
	if err != nil {
		isFeatured = false
	}

	dishes, err := getDishesFromDb(isFeatured)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dishesJson, err := misc.GetJsonFromJsonObjs(dishes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(dishesJson)
}

func putDishes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("PUT operation not supported on /dishes"))
}

func postDishes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dish, err := getDishFromBody(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status, err := createDishInDb(dish)
	statusJson, _ := misc.GetJsonFromJsonObjs(status)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJson)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(statusJson)
}

func deleteDishes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	status, err := deleteDishesFromDb()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	statusJson, err := misc.GetJsonFromJsonObjs(status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(statusJson))
}
