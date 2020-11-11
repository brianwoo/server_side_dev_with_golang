package upload

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"confusion.com/bwoo/auth"
	"confusion.com/bwoo/config"
	"confusion.com/bwoo/cors"
	"confusion.com/bwoo/misc"

	"github.com/julienschmidt/httprouter"
)

const fileUploadFormFileKey = "imageFile"

var imageDirectory string

// only accept jpg,jpeg,png and gif
// store in public/images
// use original name from client
// only allow post.  Get PUT and DELETE not allowed

func SetupRoutes(router *httprouter.Router, config config.Config) {

	imageDirectory = config.PublicImagesDir
	//imageDirectoryFull = config.GetPublicImagesDir()

	// auth methods
	router.POST("/imageUpload", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(postImageUpload))))
	router.PUT("/imageUpload", methodNotSupported)
	router.GET("/imageUpload", methodNotSupported)
	router.DELETE("/imageUpload", methodNotSupported)

	router.GET("/images/:imageName", getImage)
}

func methodNotSupported(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	httpMethod := r.Method
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(httpMethod + " operation not supported on /imageUpload"))
}

func getImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	fileName := ps.ByName("imageName")
	file, err := os.Open(filepath.Join(imageDirectory, fileName))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer file.Close()

	// Get Content-type of the file
	fHeader := make([]byte, 512)
	file.Read(fHeader)
	fContentType := http.DetectContentType(fHeader)

	// get file size
	stat, err := file.Stat()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fSize := strconv.FormatInt(stat.Size(), 10)
	// include the following line if you want to download the image
	// w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", fContentType)
	w.Header().Set("Content-Length", fSize)

	file.Seek(0, 0)
	io.Copy(w, file)
}

func postImageUpload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	r.ParseMultipartForm(5 * 1024 * 1024) // 5MB
	file, handler, err := r.FormFile(fileUploadFormFileKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create an empty file on filesystem
	f, err := os.OpenFile(filepath.Join(imageDirectory, handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()

	// Copy the file to the images directory
	io.Copy(f, file)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	uploadResult := UploadResult{
		FieldName:    fileUploadFormFileKey,
		OriginalName: handler.Filename,
		Encoding:     handler.Header.Get("Encoding"),
		MimeType:     handler.Header.Get("Content-Type"),
		Destination:  imageDirectory,
		Filename:     handler.Filename,
		Path:         imageDirectory + "/" + handler.Filename,
		Size:         handler.Size}

	resultJson, _ := misc.GetJsonFromJsonObjs(uploadResult)
	w.Write(resultJson)
}
