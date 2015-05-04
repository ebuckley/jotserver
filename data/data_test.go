package data

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"testing"
)

var testDbFile = "/tmp/jotservertest.db"

func mockConnection(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", testDbFile)
	if err != nil {
		t.Fatal("Error creating db")
	}
	return db
}

func TestGetHash(t *testing.T) {
	testVal := Password("somevalule")
	hashed := testVal.GetHash()
	if len(hashed) < len(testVal) {
		t.Errorf("expected a hashed value but got: %v \n", hashed)
	}
}

func TestUserWorkflow(t *testing.T) {
	db := mockConnection(t)

	// Ignore error from temoving testDb incase it doesn't exist anymore
	os.Remove(testDbFile)

	CreateDB(db)
	defer db.Close()

	var name = "ersin"
	var rawPw = "coffee"
	isUser := IsRegisteredUser(db, name)
	if isUser != -1 {
		t.Errorf("Error: user should not be created %v %v \n", name, isUser)
	}

	didCreateUser := CreateUser(db, name, Password(rawPw))
	if !didCreateUser {
		t.Errorf("Error: did not create user %v %v %v", name, Password(rawPw), didCreateUser)
	}

	if IsRegisteredUser(db, "bob") != -1 {
		t.Errorf("Error: phantom login can happen")
	}
	if IsRegisteredUser(db, name) == -1 {
		t.Errorf(" %v should exist as a user but does not", name)
	}

	if IsValidLogin(db, name, Password(rawPw)) == -1 {
		t.Errorf("Login should be valid for %v %v", name, rawPw)
	}

	if IsValidLogin(db, name, Password("ahuuehueh")) != -1 {
		t.Errorf("Login should not be valid for %v %v", name, "ahuehhueuhe")
	}

	if leval == -1 {
		t.Errorf("Login should be valid for %v %v", name, rawPw)
	}
}
