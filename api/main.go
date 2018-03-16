package main

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var myTime int
var mapmutex *sync.Mutex
var accounts map[string][]string
var db *sql.DB

var (
	httpClient    *http.Client
	lasthashblock string
)

const (
	MaxIdleConnections int = 20
	RequestTimeout     int = 60
)

// init HTTPClient
func init() {
	httpClient = createHTTPClient()
	accounts = make(map[string][]string)
	mapmutex = &sync.Mutex{}
}

// createHTTPClient for connection re-use
func createHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxIdleConnsPerHost: MaxIdleConnections,
		},
		Timeout: time.Duration(RequestTimeout) * time.Second,
	}

	return client
}

//Wrap some panics, because http must run forever
func RecoverWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				log.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func main() {

	//var host, port, user, pass string
	db = NewDB(Config.Database.Host,
		Config.Database.Port,
		Config.Database.Username,
		Config.Database.Password,
		Config.Database.DatabaseName,
			Config.Database.Ssl)

	//Get event from filter routine and channels
	//blockid - get block and send to BlockDetail
	//get pending tx and send to TransactionDetail
	blockIdChan := make(chan string)

	go GetEvents(blockIdChan)

	//Get details from transaction by hash and send to insert channel(also update user)
	TsChan := make(chan []toDB, 100)

	//Get details from block by hash and send bunch of transaction to TransactionDetail
	go BlockDetail(blockIdChan, TsChan)

	//Listen forever TransactionDetail for insert to DB
	go func() {
		for {
			select {
			case toDB := <-TsChan:
				go InsertTransactions(db, toDB)
			}
		}
	}()

	//config is global from config package
	fmt.Printf("Server listening on port %s", Config.Port)
	http.Handle("/gettransactions", RecoverWrap(http.HandlerFunc(commHandler)))
	srv := &http.Server{
		Addr:         ":" + Config.Port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
