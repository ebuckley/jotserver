package requests

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"jotserver/data"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var goodAuthRequest = `{
	"user": "ersin",
	"password": "daedbeef"
}`

func TestNewAuth(t *testing.T) {
	tok, err := newAuthToken("huehuehueh")

	if err != nil {
		t.Errorf("Failed because function returned an error: %v", err)
	}

	if len(tok) == 0 {
		t.Errorf("Failed because a token wasn't returned ")
	}
}

func TestAuthHandler(t *testing.T) {
	db := mockdb()
	defer db.Close()
	registerUser(db, t)

	requestBody := strings.NewReader(goodAuthRequest)
	req, err := http.NewRequest("POST", "http://localhost.com/authenticate", requestBody)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	AuthHandler(db)(w, req)

	if w.Code != 200 {
		t.Errorf("response not successfull")
	}

	var responseBody map[string]interface{}

	err = json.Unmarshal([]byte(w.Body.String()), &responseBody)
	if err != nil {
		t.Errorf("response unparsable")
	}

	token := responseBody["Token"].(string)
	if len(token) < 1 {
		t.Error(token)
	}

}

func TestAuthAndRestrictedVisit(t *testing.T) {
	db := mockdb()
	defer db.Close()
	registerUser(db, t)

	requestBody := strings.NewReader(goodAuthRequest)
	req, err := http.NewRequest("POST", "http://localhost.com/authenticate", requestBody)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	AuthHandler(db)(w, req)
	if w.Code != 200 {
		t.Errorf("response not successfull")
	}

	var responseString = w.Body.String()
	var responseBody map[string]interface{}
	err = json.Unmarshal([]byte(responseString), &responseBody)
	if err != nil {
		t.Errorf("response unparsable")
	}

	token := responseBody["Token"].(string)
	if len(token) < 1 {
		t.Error(token)
	}

	//use token from previous request for the next request
	restrictedReq, err := http.NewRequest("POST", "http://localhost.com/restricted", strings.NewReader(responseString))
	if err != nil {
		log.Fatal(err)
	}

	restrictedW := httptest.NewRecorder()
	RestrictedHandler(restrictedW, restrictedReq)

	log.Printf("response: %v", restrictedW.Body.String())

	var restrictedResponse map[string]interface{}
	err = json.Unmarshal([]byte(restrictedW.Body.String()), &restrictedResponse)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%v", restrictedResponse)
}

func mockdb() *sql.DB {
	src := rand.NewSource(time.Now().Unix())
	rs := rand.New(src)
	dbPath := fmt.Sprintf("/tmp/%X-test.db", rs.Intn(1000000))

	log.Println("Creating mock db: %s", dbPath)
	conn := data.DB(dbPath)
	data.CreateDB(conn)
	return conn
}
func TestRegisterUser(t *testing.T) {
	db := mockdb()
	defer db.Close()
	registerUser(db, t)
}

func registerUser(db *sql.DB, t *testing.T) {

	requestBody := strings.NewReader(goodAuthRequest)
	req, err := http.NewRequest("POST", "http://localhost.com/register", requestBody)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	RegisterHandler(db)(w, req)
	if w.Code != 201 {
		t.Errorf("response not successfull didn't get 201: %v %v", w.Code, w.Body.String())
		return
	}

	var responseString = w.Body.String()
	var responseBody map[string]interface{}
	err = json.Unmarshal([]byte(responseString), &responseBody)
	if err != nil {
		t.Errorf("response unparsable")
		return
	}

	if len(responseBody["Token"].(string)) < 32 {
		t.Errorf("msg not parsed from json %v \n", responseBody)
	}
}
