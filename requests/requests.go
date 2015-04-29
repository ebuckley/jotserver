package requests

import (
	"database/sql"
	"encoding/json"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"jotserver/data"
	"log"
	"net/http"
	"time"
)

var (
	expireTime        = time.Minute * 20
	signKey    []byte = []byte("1deadbeaf")
)

type AuthRequest struct {
	User     string `json:user`
	Password string `json:password`
}

type JotToken struct {
	Token string `json:token`
}

type ErrorResponse struct {
	Msg string `json:msg`
}

func newAuthToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	token.Claims["id"] = username
	token.Claims["exp"] = time.Now().Add(expireTime).Unix()
	tokenString, err := token.SignedString(signKey)
	return tokenString, err
}

func jotTokenResponse(token string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	response := new(JotToken)
	response.Token = token

	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		log.Println("Major failure while encoding JotToken response")
	}
}

func errorResponse(reason string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	response := new(ErrorResponse)
	response.Msg = reason

	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		log.Println("Major fail when encoding errorResponse", err)
	}
}

func isValidPassword(req *AuthRequest) bool {
	if req.User == "ersin" && req.Password == "daedbeef" {
		return true
	}
	return false
}

//return false if it is options call which means we should make a direct response without body
func allowCrossOrigin(rw http.ResponseWriter, r *http.Request) bool {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	rw.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if r.Method == "OPTIONS" {
		return false
	}
	return true
}

/**
 * HTTP handlers
 */

func AuthHandler(w http.ResponseWriter, r *http.Request) {

	if !allowCrossOrigin(w, r) {
		log.Println("cross origin request recieved")
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		reason := fmt.Sprintf("ERROR: expected a POST but recieved ", r.Method)
		errorResponse(reason, w)
		return
	}

	//parse request body
	decoder := json.NewDecoder(r.Body)

	authorizationMsg := new(AuthRequest)
	err := decoder.Decode(authorizationMsg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		reason := fmt.Sprintf("Error: request was not parsable %v", err)
		errorResponse(reason, w)
		return
	}

	log.Printf("Authenticate Request: user[%s]", authorizationMsg.User, authorizationMsg.Password)

	if !isValidPassword(authorizationMsg) {
		w.WriteHeader(http.StatusForbidden)
		errorResponse("Username or Password is not valid", w)
		return
	}

	var token string
	token, err = newAuthToken(authorizationMsg.User)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		reason := fmt.Sprintf("Error with token signing:%v", err)
		log.Println(reason)
		errorResponse(reason, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	jotTokenResponse(token, w)
}

func RestrictedHandler(w http.ResponseWriter, r *http.Request) {

	if !allowCrossOrigin(w, r) {
		log.Println("cross origin request recieved")
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse("Expected post", w)
		return
	}

	//parse request body
	decoder := json.NewDecoder(r.Body)

	tokenReq := new(JotToken)
	err := decoder.Decode(tokenReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		reason := fmt.Sprintf("Error: expected a better request %v", err)
		errorResponse(reason, w)
		return
	}

	// validate the token
	token, err := jwt.Parse(tokenReq.Token, func(token *jwt.Token) (interface{}, error) {
		return signKey, nil
	})

	switch err.(type) {
	case nil:
		if !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			errorResponse("Token not valid", w)
			return
		}
		log.Printf("Access restricted area! Token: %v\n", token)
		w.WriteHeader(http.StatusOK)
		jotTokenResponse(tokenReq.Token, w)
		return

	case *jwt.ValidationError:
		vErr := err.(*jwt.ValidationError)

		switch vErr.Errors {
		case jwt.ValidationErrorExpired:
			w.WriteHeader(http.StatusUnauthorized)
			errorResponse("Token Expired, get a new one.", w)
			return

		default:
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("ValidationError error: %v \n", vErr.Errors)
			errorResponse("Error parsing token!", w)
			return
		}
	default:
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Token parse error: %v", err)
		errorResponse("Token parse error", w)
	}
}

func RegisterHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		if !allowCrossOrigin(w, r) {
			log.Println("cross origin request recieved")
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse("Expected http POST", w)
		}

		decoder := json.NewDecoder(r.Body)
		registration := new(AuthRequest)
		err := decoder.Decode(registration)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := fmt.Sprintf("Error parsing request: %s", err)
			errorResponse(msg, w)
		}

		if data.IsRegisteredUser(db, registration.User) != -1 {
			w.WriteHeader(http.StatusConflict)
			errorResponse("user already exists", w)
		}

		if !data.CreateUser(db, registration.User, data.Password(registration.Password)) {
			w.WriteHeader(http.StatusInternalServerError)
			errorResponse("Error: could not create user", w)
		}

		userId := data.IsValidLogin(db, registration.User, data.Password(registration.Password))
		if userId == -1 {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Error: registration was not able to login with created user", registration)
			errorResponse("Error: with registration, try again", w)
		}

		var newTkn string
		newTkn, err = newAuthToken(registration.User)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Error creating auth token", err)
			errorResponse("Error creating token", w)
		}
		w.WriteHeader(http.StatusCreated)
		jotTokenResponse(newTkn, w)
	}
}
