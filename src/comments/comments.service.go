package comments

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"confusion.com/bwoo/auth"
	"confusion.com/bwoo/cors"
	"confusion.com/bwoo/misc"
	"github.com/julienschmidt/httprouter"
)

func SetupRoutes(router *httprouter.Router) {

	// dish
	router.GET("/dishes/:dishId/comments/:commentId", cors.CorsAllOrigin(getComment))
	router.PUT("/dishes/:dishId/comments/:commentId", cors.Cors(auth.VerifyUser(putComment)))
	router.POST("/dishes/:dishId/comments/:commentId", cors.Cors(auth.VerifyUser(postComment)))
	router.DELETE("/dishes/:dishId/comments/:commentId", cors.Cors(auth.VerifyUser(deleteComment)))

	// dishes
	router.GET("/dishes/:dishId/comments", cors.CorsAllOrigin(getComments))
	router.PUT("/dishes/:dishId/comments", cors.Cors(auth.VerifyUser(putComments)))
	router.POST("/dishes/:dishId/comments", cors.Cors(auth.VerifyUser(postComments)))
	router.DELETE("/dishes/:dishId/comments", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(deleteComments))))
}

func getCommentFromBody(body io.ReadCloser) (Comment, error) {

	var comment Comment
	err := json.NewDecoder(body).Decode(&comment)
	if err != nil {
		return comment, err
	}
	return comment, nil
}

/****************************
* Comment operations
****************************/
func getComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	dishIdInt, err := misc.GetInt64FromString(dishId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	commentId := ps.ByName("commentId")
	commentIdInt, err := misc.GetInt64FromString(commentId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	comment, err := getCommentFromDb(dishIdInt, commentIdInt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var jsonComment []byte = make([]byte, 0)
	if comment == nil {
		jsonComment = misc.GetEmptyJsonByteArray()
	} else {
		jsonComment, err = misc.GetJsonFromJsonObjs(comment)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonComment)
}

func isCommentBelongsToUser(dishId, commentId, userId int64) bool {

	comment, err := getCommentFromDb(dishId, commentId)
	if err != nil {
		return false
	}

	if comment == nil {
		return false
	}

	return comment.Author.ID == userId
}

func putComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	dishIdInt, err := misc.GetInt64FromString(dishId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	commentId := ps.ByName("commentId")
	commentIdInt, err := misc.GetInt64FromString(commentId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := auth.GetClaimsFromRequest(r)
	userId, _ := misc.GetInt64FromString(claims.UserId)

	isCommentBelongsToUser := isCommentBelongsToUser(dishIdInt, commentIdInt, userId)
	if !isCommentBelongsToUser {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	comment, err := getCommentFromBody(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updatedComment, err := updateCommentFromDb(dishIdInt, commentIdInt, comment, userId)
	if err != nil && updatedComment == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// if update was run but no rows updated (i.e. updatedDish.ID = 0),
	// we will just return an empty json object
	var updatedCommentJson []byte
	if updatedComment.ID == 0 {
		updatedCommentJson = misc.GetEmptyJsonByteArray()
	} else {
		updatedCommentJson, _ = misc.GetJsonFromJsonObjs(updatedComment)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(updatedCommentJson))
}

func postComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	commentId := ps.ByName("commentId")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("POST operation not supported on /dishes/" + dishId + "/comments/" + commentId))
}

func deleteComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	dishIdInt, err := misc.GetInt64FromString(dishId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	commentId := ps.ByName("commentId")
	commentIdInt, err := misc.GetInt64FromString(commentId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := auth.GetClaimsFromRequest(r)
	userId, _ := misc.GetInt64FromString(claims.UserId)

	isCommentBelongsToUser := isCommentBelongsToUser(dishIdInt, commentIdInt, userId)
	if !isCommentBelongsToUser {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	status, err := deleteCommentFromDb(dishIdInt, commentIdInt, userId)
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
* Comments operations
****************************/
func getComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	dishIdInt, err := misc.GetInt64FromString(dishId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	comments, err := getCommentsFromDb(dishIdInt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	commentsJson, err := misc.GetJsonFromJsonObjs(comments)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(commentsJson)
}

func putComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("PUT operation not supported on /dishes/" + dishId + "/comments"))
}

func postComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	claims := auth.GetClaimsFromRequest(r)
	userId, err := misc.GetInt64FromString(claims.UserId)

	dishId := ps.ByName("dishId")
	dishIdInt, err := misc.GetInt64FromString(dishId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	comment, err := getCommentFromBody(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status, err := createCommentInDb(dishIdInt, userId, comment)
	statusJson, _ := misc.GetJsonFromJsonObjs(status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJson)
		return
	}

	// NOTE: in NodeJS, it returns all the dish and comments
	w.Header().Set("Content-Type", "application/json")
	w.Write(statusJson)
}

func deleteComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	dishId := ps.ByName("dishId")
	dishIdInt, err := misc.GetInt64FromString(dishId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status, err := deleteCommentsFromDb(dishIdInt)
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
