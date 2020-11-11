package promotions

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

	// promotion
	router.GET("/promotions/:promotionId", cors.CorsAllOrigin(getPromotion))
	router.PUT("/promotions/:promotionId", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(putPromotion))))
	router.POST("/promotions/:promotionId", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(postPromotion))))
	router.DELETE("/promotions/:promotionId", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(deletePromotion))))

	// promotions
	router.GET("/promotions", cors.CorsAllOrigin(getPromotions))
	router.PUT("/promotions", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(putPromotions))))
	router.POST("/promotions", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(postPromotions))))
	router.DELETE("/promotions", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(deletePromotions))))
}

/****************************
* Helper functions
****************************/
func getPromotionFromBody(body io.ReadCloser) (Promotion, error) {

	var promotion Promotion
	err := json.NewDecoder(body).Decode(&promotion)
	if err != nil {
		return Promotion{}, err
	}

	return promotion, nil
}

/****************************
* Promotion operations
****************************/
func getPromotion(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	promotionId := ps.ByName("promotionId")
	promotionIdInt, err := misc.GetInt64FromString(promotionId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	promotion, err := getPromotionFromDb(promotionIdInt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var jsonPromo []byte = make([]byte, 0)
	if promotion == nil {
		jsonPromo, err = misc.GetJsonFromJsonObjs(struct{}{})
	} else {
		jsonPromo, err = misc.GetJsonFromJsonObjs(promotion)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonPromo)
}

func putPromotion(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	promotionId := ps.ByName("promotionId")
	promotionIdInt, err := misc.GetInt64FromString(promotionId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	promotion, err := getPromotionFromBody(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updatedPromotion, err := updatePromotionFromDb(promotionIdInt, promotion)
	if err != nil && updatedPromotion == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// if update was run but no rows updated (i.e. updatedPromotion.ID = 0),
	// we will just return an empty json object
	var updatedPromotionJson []byte
	if updatedPromotion.ID == 0 {
		updatedPromotionJson = misc.GetEmptyJsonByteArray()
	} else {
		updatedPromotionJson, _ = misc.GetJsonFromJsonObjs(updatedPromotion)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(updatedPromotionJson))
}

func postPromotion(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	promotionId := ps.ByName("promotionId")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("POST operation not supported on /promotions/" + promotionId))
}

func deletePromotion(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	promotionId := ps.ByName("promotionId")
	promotionIdInt, err := misc.GetInt64FromString(promotionId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status, err := deletePromotionFromDb(promotionIdInt)
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
* Promotions operations
****************************/
func getPromotions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	isFeaturedStr := r.URL.Query().Get("featured")
	isFeatured, err := strconv.ParseBool(isFeaturedStr)
	if err != nil {
		isFeatured = false
	}

	promotions, err := getPromotionsFromDb(isFeatured)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	promosJson, err := misc.GetJsonFromJsonObjs(promotions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(promosJson)
}

func putPromotions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("PUT operation not supported on /promotions"))
}

func postPromotions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	promotion, err := getPromotionFromBody(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status, err := createPromotionInDb(promotion)
	statusJson, _ := misc.GetJsonFromJsonObjs(status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJson)
		return
	}

	// NOTE: in NodeJS, it returns all the promotion and comments
	w.Header().Set("Content-Type", "application/json")
	w.Write(statusJson)
}

func deletePromotions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	status, err := deletePromotionsFromDb()
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
