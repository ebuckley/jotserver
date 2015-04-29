package main

import (
	"database/sql"
	"flag"
	"jotserver/data"
	r "jotserver/requests"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	connString *string
	dbFileName *string
)

//create or open database connection
func getDB(filename string) *sql.DB {
	conn := data.DB(filename)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Println("Database Being created %s", filename)
		data.CreateDB(conn)
	}
	return conn
}

func main() {

	connString = flag.String("connection", ":8080", "A string for listening and serving the http server (default: 8080)")
	dbFileName = flag.String("db", "jotserver.db", "the location of the database file")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	db := getDB(*dbFileName)

	log.Printf("starting jotserver on %s", *connString)
	http.HandleFunc("/authenticate", r.AuthHandler(db))
	http.HandleFunc("/restricted", r.RestrictedHandler)
	http.HandleFunc("/register", r.RegisterHandler(db))

	//TODO catch errors and log them
	go func() {
		sig := <-sigs
		log.Println("recieved signal", sig)
		done <- true
	}()

	go func() {
		http.ListenAndServe(*connString, nil)
	}()
	<-done
	db.Close()
	log.Println("exiting")
}
