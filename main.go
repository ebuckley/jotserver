package main

import (
	"database/sql"
	"flag"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/ebuckley/jotserver/data"
	r "github.com/ebuckley/jotserver/requests"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	connString   *string
	dbFileName   *string
	statsdConfig *string
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

func getStatsDClient(cfg string) statsd.Statter {
	conn, err := statsd.NewClient(cfg, "jotserver")
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func requestStatter(name string, statsClient statsd.Statter, handler http.HandlerFunc) http.HandlerFunc {

	return func(rw http.ResponseWriter, req *http.Request) {
		before := time.Now()
		handler(rw, req)
		delta := time.Now().Sub(before)
		statsClient.Gauge("jotserver."+name, int64(delta), 1.0)
	}
}

func main() {

	connString = flag.String("connection", ":8080", "A string for listening and serving the http server (default: 8080)")
	dbFileName = flag.String("db", "jotserver.db", "The location of the database file")
	statsdConfig = flag.String("statsd", "127.0.0.1:8125", "statsd client location")

	flag.Parse()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	db := getDB(*dbFileName)
	statsClient := getStatsDClient(*statsdConfig)
	statsClient.Inc("startServer", 1, 1.0)

	log.Printf("starting jotserver on %s", *connString)

	http.HandleFunc("/authenticate", requestStatter("authenticate", statsClient, r.AuthHandler(db)))
	http.HandleFunc("/restricted", requestStatter("restricted", statsClient, r.RestrictedHandler))
	http.HandleFunc("/register", requestStatter("register", statsClient, r.RegisterHandler(db)))

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
	statsClient.Close()

	log.Println("exiting")
}
