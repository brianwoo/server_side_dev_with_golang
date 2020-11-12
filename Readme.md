# Building a Web API Project with GoLang (Examples)

## Introduction
GoLang is a great programming language for building Web API projects. To demonstrate the power of GoLang, I wanted to build a sample Web API project which contains some of the commonly used features.

To get an idea of the features to be included in the project, I went back to one of the courses that I previously took, [Server-side Development with NodeJS, Express and MongoDB](https://www.coursera.org/learn/server-side-nodejs), offered by Coursera / Hong Kong University of Science and Technology. I decided to implement all the functionality taught in the course with GoLang and MySQL.

This article provides a number of examples on how these features can be implemented ([Git Repo](https://github.com/brianwoo/server_side_dev_with_golang)).

## List of Examples
- Implementing basic REST API
- Setting up database connection
- Implementing database CRUD functions
- Implementing database CRUD functions (with transaction)
- Implementing authentication and JSON Web Token (JWT)
- Setting up HTTPS (with HTTP redirection)
- Uploading / downloading files
- Setting up Cross-Origin Resource Sharing (CORS)
- Implementing OAuth2 with Facebook as an alternative way to login


## Dependencies:
```console
go get github.com/julienschmidt/httprouter

go get -u github.com/go-sql-driver/mysql

go get github.com/dgrijalva/jwt-go

go get golang.org/x/crypto

go get golang.org/x/oauth2
```

## Startup MySQL:
```console
docker-compose -f docker_compose.yaml up -d
```

## Implementing Basic REST API

To start off, GoLang provides a router module in its stdlib, but I decided to use [julienschmidt's httprouter](https://github.com/julienschmidt/httprouter) instead of GoLang's built-in module.  The reason is that julienschmidt's httprouter provides a cleaner way to implement the routes.


### Route Setup

Setting up routes is quite simple, you will need to setup three things: HTTP Method, a URL and a handler method.

To tell httprouter the part of the URL is a Resource ID, you will need to add a colon before the Resource ID variable (we have :dishId in this example).
```go
func SetupRoutes(router *httprouter.Router) {

	// dish
	router.GET("/dishes/:dishId", getDish)
	router.PUT("/dishes/:dishId", putDish)
	router.POST("/dishes/:dishId", postDish)
	router.DELETE("/dishes/:dishId", deleteDish)

	// dishes
	router.GET("/dishes", getDishes)
	router.PUT("/dishes", putDishes)
	router.POST("/dishes", postDishes)
	router.DELETE("/dishes", deleteDishes)
}
```

A httprouter handler interface looks very much like GoLang's http handler interface, except the httprouter handler takes an extra parameter ps (third param). To get the dishId resource ID, we can use the method ps.ByName().
```go
func deleteDish(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

    // The :dishId specified in the routing
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
```

## Setting Up Database Connection

The database connection setup is database specific. The following example is for MySQL. You can get more information on mysql driver [here](https://github.com/go-sql-driver/mysql).
```go
var DbConn *sql.DB

func SetupDatabase(config config.Config) {

	connString := config.GetConnString()

	var err error
	DbConn, err = sql.Open(config.DbDriver, connString)
	if err != nil {
		log.Fatal(err)
	}

	DbConn.SetMaxOpenConns(4)
	DbConn.SetMaxIdleConns(4)
	DbConn.SetConnMaxLifetime(60 * time.Second)
}
```

## Implementing Database CRUD Functions

Here is an example of getting data from the database. The QueryRowContext() function takes a prepared statement and a list of values as parameters. After the query is done, the Scan() function extracts the values from the row result.

```go
func getDishFromDb(dishId int64) (*Dish, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	row := database.DbConn.QueryRowContext(ctx, `SELECT 
            id,
            name,
            image,
            category,
            label,
            price,
            CASE WHEN featured = 0 then 'false' ELSE 'true' END,
            description,
            createdAt,
            updatedAt
        FROM dish
        WHERE id = ?`, dishId)

	var dish Dish
	err := row.Scan(&dish.ID,
		&dish.Name,
		&dish.Image,
		&dish.Category,
		&dish.Label,
		&dish.Price,
		&dish.Featured,
		&dish.Description,
		&dish.CreatedAt,
		&dish.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &dish, nil
}
```

## Implementing Database CRUD Functions (with transaction)

There are situations where you might need transaction support when doing multiple INSERTs. To insert a record with a transaction, we will need to use sql.Tx.ExecContext(), instead of sql.DB.ExecContext().  Fortunately, both ExecContext() functions have the same function signature and we can have setup a function to return either ExecContext() function:
```go
func getExecContextFunc(tx *sql.Tx) func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {

	if tx != nil {
		return func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
			return tx.ExecContext(ctx, query, args...)
		}
	} else {
		return func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
			return database.DbConn.ExecContext(ctx, query, args...)
		}
	}
}
```

To avoid duplicating the insert logic, we can extract a shared logic in a separate function createFavoriteDishInDbInternal() to handle both transactional and non-transactional inserts.
```go
func createFavoriteDishInDbInternal(tx *sql.Tx, ctx context.Context, userId, dishId int64) (*misc.Status, error) {

	status := &misc.Status{NumOfRowsAffected: 0, IsOk: 0}
	sqlInsert := `INSERT INTO favoriteDish(
					userId,
					dishId
				)
				VALUES (
					?,?
				)`

	execContextFunc := getExecContextFunc(tx)
	result, err := execContextFunc(ctx, sqlInsert, userId, dishId)
	if err != nil {
		return status, err
	}

	rowsInserted, _ := result.RowsAffected()
	status.NumOfRowsAffected = rowsInserted
	status.IsOk = 1
	return status, nil
}
```

Implementing a transaction is relatively straightforward. It starts with BeginTx() to begin a transaction and tx.Commit() to commit. When needed, we can rollback the insert by calling tx.Rollback(). In the following example, we do not commit unless all records have been inserted successfully.
```go
func createFavoriteDishesInDb(userId int64, favDishes favoriteDishes) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// we use transaction so we can commit atomicly (i.e. all or none)
	tx, err := database.DbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	status := &misc.Status{NumOfRowsAffected: 0, IsOk: 0}
	for _, favDish := range favDishes {

		result, err := createFavoriteDishInDbInternal(tx, ctx, userId, favDish.ID)
		if err != nil {
			tx.Rollback()
			status.NumOfRowsAffected = 0
			return status, err
		}

		status.NumOfRowsAffected += result.NumOfRowsAffected
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	status.IsOk = 1
	return status, nil
}
```

As for inserting only one single dish at a time, we can reuse the same logic without using transaction.
```go
func createFavoriteDishInDb(userId, dishId int64) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return createFavoriteDishInDbInternal(nil, ctx, userId, dishId)
}
```

## Implementing Authentication and JSON Web Token (JWT)

The [dgrijalva/jwt-go](https://github.com/dgrijalva/jwt-go) provides a simple way to generate and validate JWTs. In this example, we will extract a JWT token, stored in the Authorization field in the header:

```go
func GetJwtTokenFromRequest(r *http.Request) (string, error) {

	// header is a string: KEY: Authorization VALUE: Bearer tokenstringInBase64
	tokenInHeaderVal := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	if len(tokenInHeaderVal) != 2 {
		return "", fmt.Errorf("Malformed token")
	}

	return tokenInHeaderVal[1], nil
}
```

The jwt-go module can also parse the JWT and extract the claims. The claims (aka the payload) is usually used to store information such as a User's ID or name. However, do not include sensitive information like a password as the claims can be extracted by anyone:

![Image of JWT decoded](/images/jwt_decode.png)

You can find out more about JWT from [this article](https://dzone.com/articles/what-is-jwt-token).

### Validating JWTs
The claims can be extracted using the function jwt.ParseWithClaims(). The function returns the JWT token which can then be validated by looking at the token.Valid attribute.
```go
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
```
To put this all together, we have a verifyUser() function which acts as a middleware function. This middleware function validates the user and if the user has a valid token, the next() function will be called along with claims. This next() function is the auth.VerifyAdmin() function as you will see below.
```go
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
```

Middleware functions are usually chained together for validation.  For example, we can chain Cors() --> VerifyUser() --> VerifyAdmin() together and execute them sequentially before calling putPromotions(). If VerifyUser() fails, the function will return an error to the consumer. VerifyAdmin() and putPromotions() will NOT be executed.
```go
router.PUT("/promotions", cors.Cors(auth.VerifyUser(auth.VerifyAdmin(putPromotions))))
```

### Generating JWTs
The jwt-go module also provides a convenient way of generating JWTs. The information stored inside the Claims can be defined as a struct, along with the jwt.StandardClaims which includes an expiry time. A new JWT token can be generated by calling the jwt.NewWithClaims() function.
```go
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
```

## Setting up HTTPS (with HTTP Redirection)
In this project, we expose both HTTP (port 3000) and HTTPS (port 3443) to the consumers. To make things more secure, we can redirect an incoming HTTP connection to HTTPS to ensure the communication is secure.
```go
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
```

## Uploading / Downloading Files
File Upload and download are functionality commonly used on a web server. The following examples are to illustrate how to implement these features using GoLang.

### Uploading a File
In this example, we have a handler which can save an uploaded file to disk.  Please note that handler.Filename should be sanitized before passing to filepath.Join() to prevent malicious attacks. 

```go
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
```

### Downloading a File
Similarly, we have the getImage() function to handle image file download from the filesystem. Again, please note that filename should be sanitized before passing to filepath.Join() to prevent malicious attacks.
```go
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
```

## Setting up Cross-Origin Resource Sharing (CORS)

As for implementing CORS, the httprouter provides a convenient option to handle HTTP OPTIONS requests. router.GlobalOPTIONS can be setup as a default handler for handling preflight requests to the webservice.

```go
func setupDefaultHttpOptions(router *httprouter.Router) {
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		wHeader := w.Header()
		setAccessControlAllowOriginIfValid(r.Header, wHeader)
		wHeader.Add("Access-Control-Allow-Credentials", "true")
		wHeader.Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, 
            Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		wHeader.Set("Access-Control-Allow-Methods", allowedMethods)

		// Adjust status code to 204
		w.WriteHeader(http.StatusNoContent)
	})
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
```

Similarly, the Cors() function can be setup as a middleware method to handle PUT, POST & DELETE requests.
```go
func Cors(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		wHeader := w.Header()
		setAccessControlAllowOriginIfValid(r.Header, wHeader)
		wHeader.Add("Access-Control-Allow-Credentials", "true")
		wHeader.Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, 
            Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		wHeader.Add("Access-Control-Allow-Methods", allowedMethods)

		next(w, r, ps)
	}
}
```

## Implementing OAuth2 with Facebook as an Alternative Way to Login
One of the popular ways to login these days is to use one of your social media accounts. In this project, we have also explored this option and we allow a user to login with his regular username and password or Facebook login as an alternative. In order to login via Facebook, we will have to use Facebook's OAuth2. Fortunately, GoLang provides that functionality.

To use Facebook's OAuth2, you will first need to setup an app ([instructions](https://dzone.com/articles/implementing-oauth2-social-login-with-facebook-par)) through the [Facebook Developer Portal](https://developers.facebook.com/).  

A regular OAuth2 Facebook Login workflow would involve a client logging in to Facebook, upon a successful login, an access token will be returned from Facebook ([workflow illustration](https://dzone.com/articles/implementing-oauth2-social-login-with-facebook-par-1)). Since we don't have a web client, I have created a login endpoint and a callback endpoint to receive the access token from Facebook.

As a security precaution, it is recommended to setup a random character string in the login request and to validate that string in the callback. We can accomplish that with a state cookie.

```go
var facebookConfig *oauth2.Config

func loginFacebook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	oauthState := generateStateOauthCookie(w)
	url := facebookConfig.AuthCodeURL(oauthState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func facebookLoginCallback(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// Read oauthState from Cookie
	stateOauthCookie, err := r.Cookie("oauthstate")
	if err != nil {
		fmt.Println("Cannot get state token", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if stateOauthCookie != nil && r.FormValue("state") != stateOauthCookie.Value {
		log.Println("invalid OAuth state")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := facebookConfig.Exchange(oauth2.NoContext, r.FormValue("code"))
	if err != nil {
		fmt.Printf("facebookConfig.Exchange() failed with '%s'\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fmt.Println("Facebook Access Token:", token.AccessToken)
}
```

Once we get an access token from Facebook, we can simulate a login session from a Facebook user. With this token, we have the ability to access the user's Facebook profile. In our example, we access the user's Facebook profile and if this user is not found in our database, we create a new local user account with his Facebook ID; otherwise, we validate the user's Facebook ID in our database and return a new JWT token.
```go
func loginWithFacebookToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	accessToken, err := getAccessToken(r, ps)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp, err := http.Get(
        "https://graph.facebook.com/me?fields=id,name,first_name,last_name,email&access_token=" +
		url.QueryEscape(accessToken))
	if resp.StatusCode >= http.StatusBadRequest {
		fmt.Println("Unable to login to Facebook")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err != nil {
		fmt.Printf("Facebook Graph API error: %s\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	defer resp.Body.Close()

	facebookUserInfo, err := getFacebookUserInfoFromBody(resp.Body)
	if err != nil {
		fmt.Printf("ReadAll: %s\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userInfo, err := findUserByFacebookIdCreateIfNotFound(facebookUserInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	loginResult := auth.GetLoginResult(userInfo.ID, userInfo.Admin, true)
	resultJson, _ := misc.GetJsonFromJsonObjs(loginResult)

	w.Header().Set("Content-Type", "application/json")

	// return jwt token
	w.WriteHeader(http.StatusOK)
	w.Write(resultJson)
}

func findUserByFacebookIdCreateIfNotFound(facebookUserInfo FacebookUserInfo) (*auth.UserInfo, error) {

	userInfo, err := getFacebookUserFromDb(facebookUserInfo.ID)
	if err != nil {
		return nil, err
	} else if userInfo == nil {
		_, ok := createUserInDb(facebookUserInfo)
		if !ok {
			return nil, fmt.Errorf("Unable to create user")
		}
		userInfo, _ = getFacebookUserFromDb(facebookUserInfo.ID)
	}

	return userInfo, nil
}
```

## Conclusion
GoLang is a great programming language to create a Web API and there are a number of helpful modules which make building a project much easier. 

In this project, I have demonstrated how quickly it is to build a Web API with some of the commonly used features in GoLang.