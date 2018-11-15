package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func determineListenAddress() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return "", fmt.Errorf("$PORT not set")
	}
	return ":" + port, nil
}

func routerInit() {
	//Setup router for APIw
	r := mux.NewRouter()
	addr, err := determineListenAddress()
	if err != nil {
		log.Fatal(err)
	}

	addRoutes(r)

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(addr, r))
}

func addRoutes(r *mux.Router) {

	//Routes located in routes.go
	r.HandleFunc("/api/{serverid}/rss", addRss).Methods("POST") // Subscribe to a new rss feed on {serverid}
	r.HandleFunc("/api/{serverid}/rss", listRss).Methods("GET") // List all rss feeds currently subscribed on {serverid}
	r.HandleFunc("/api/{apiKey}", addRss).Methods("POST")
	r.HandleFunc("/api/rss", listAllRss).Methods("GET")
	r.HandleFunc("/api/rss/{apiKey}", listRss).Methods("GET")
}
