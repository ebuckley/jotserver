package data

import (
	sha "crypto/sha256"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var hashSalt = "usaltybro?"

var createTableQuery = `
create table users (
	id integer primary key,
	name text,
	password text
);
delete from users;
`

var createUserQuery = `
insert into users(name, password) values (?, ?)
`

var findUserQuery = `
select id from users where name = ? and password = ?
`

var findUsersQuery = `
select id from users where name = ?
`

func DB(location string) *sql.DB {
	db, err := sql.Open("sqlite3", location)
	if err != nil {
		log.Fatal("error creating sqlite3 connection: ", err)
	}
	return db
}

func CreateDB(db *sql.DB) {
	_, err := db.Exec(createTableQuery)
	if err != nil {
		log.Printf("%q: %s \n", err, createTableQuery)
		return
	}
}

func CreateUser(db *sql.DB, name string, pw Password) bool {

	if IsRegisteredUser(db, name) != -1 {
		log.Printf("user already exists", name, pw)
		return false
	}

	_, err := db.Exec(createUserQuery, name, pw.GetHash())
	if err != nil {
		log.Printf("%q: %s \n", err, createUserQuery)
		return false
	}
	log.Printf("New User Registered: %v", name)
	return true
}

func IsRegisteredUser(db *sql.DB, name string) int {
	rows, err := db.Query(findUsersQuery, name)
	if err != nil {
		log.Printf("%q: %s \n", err, findUsersQuery)
	}
	defer rows.Close()

	if !rows.Next() {
		log.Printf("When checking isRegistered, User not found: %v ", name)
		return -1
	}

	var id int
	e := rows.Scan(&id)
	if e != nil {
		log.Fatal("Error parsing ID: %q\n", err)
	}
	return id
}
func IsValidLogin(db *sql.DB, name string, pw Password) int {
	rows, err := db.Query(findUserQuery, name, pw.GetHash())
	if err != nil {
		log.Printf("%q: %s \n", err, findUserQuery)
	}
	defer rows.Close()

	if !rows.Next() {
		log.Printf("Invalid login check for: %v - %X", name, pw.GetHash())
		return -1
	}

	var id int
	e := rows.Scan(&id)
	if e != nil {
		log.Fatal("DB fail: %q\n", err)
	}
	return id
}

func (t Password) GetHash() []byte {

	f := sha.New()
	f.Write([]byte(hashSalt + string(t)))
	return f.Sum(nil)
}

type Password string
