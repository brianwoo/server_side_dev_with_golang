package leaders

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

	// leader
	router.GET("/leaders/:leaderId", cors.CorsAllOrigin(getLeader))
	router.PUT("/leaders/:leaderId", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(putLeader))))
	router.POST("/leaders/:leaderId", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(postLeader))))
	router.DELETE("/leaders/:leaderId", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(deleteLeader))))

	// leaders
	router.GET("/leaders", cors.CorsAllOrigin(getLeaders))
	router.PUT("/leaders", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(putLeaders))))
	router.POST("/leaders", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(postLeaders))))
	router.DELETE("/leaders", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(deleteLeaders))))
}

/****************************
* Helper functions
****************************/
func getLeaderFromBody(body io.ReadCloser) (Leader, error) {

	var leader Leader
	err := json.NewDecoder(body).Decode(&leader)
	if err != nil {
		return Leader{}, err
	}

	return leader, nil
}

/****************************
* Leader operations
****************************/
func getLeader(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	leaderId := ps.ByName("leaderId")
	leaderIdInt, err := misc.GetInt64FromString(leaderId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	leader, err := getLeaderFromDb(leaderIdInt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var jsonPromo []byte = make([]byte, 0)
	if leader == nil {
		jsonPromo, err = misc.GetJsonFromJsonObjs(struct{}{})
	} else {
		jsonPromo, err = misc.GetJsonFromJsonObjs(leader)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonPromo)
}

func putLeader(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	leaderId := ps.ByName("leaderId")
	leaderIdInt, err := misc.GetInt64FromString(leaderId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	leader, err := getLeaderFromBody(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updatedLeader, err := updateLeaderFromDb(leaderIdInt, leader)
	if err != nil && updatedLeader == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// if update was run but no rows updated (i.e. updatedLeader.ID = 0),
	// we will just return an empty json object
	var updatedLeaderJson []byte
	if updatedLeader.ID == 0 {
		updatedLeaderJson = misc.GetEmptyJsonByteArray()
	} else {
		updatedLeaderJson, _ = misc.GetJsonFromJsonObjs(updatedLeader)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(updatedLeaderJson))
}

func postLeader(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	leaderId := ps.ByName("leaderId")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("POST operation not supported on /leaders/" + leaderId))
}

func deleteLeader(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	leaderId := ps.ByName("leaderId")
	leaderIdInt, err := misc.GetInt64FromString(leaderId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status, err := deleteLeaderFromDb(leaderIdInt)
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
* Leaders operations
****************************/
func getLeaders(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	isFeaturedStr := r.URL.Query().Get("featured")
	isFeatured, err := strconv.ParseBool(isFeaturedStr)
	if err != nil {
		isFeatured = false
	}

	leaders, err := getLeadersFromDb(isFeatured)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	leadersJson, err := misc.GetJsonFromJsonObjs(leaders)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(leadersJson)
}

func putLeaders(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("PUT operation not supported on /leaders"))
}

func postLeaders(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	leader, err := getLeaderFromBody(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status, err := createLeaderInDb(leader)
	statusJson, _ := misc.GetJsonFromJsonObjs(status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJson)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(statusJson)
}

func deleteLeaders(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	status, err := deleteLeadersFromDb()
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
