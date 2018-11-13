package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func routerInit() {
	//Setup router for APIw
	router := mux.NewRouter().StrictSlash(true)

	addRoutes(router)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func addRoutes(r *mux.Router) {

	//Routes located in routes.go
	r.HandleFunc("/api/{serverid}/rss", addRss).Methods("POST") // Subscribe to a new rss feed on {serverid}
	r.HandleFunc("/api/{serverid}/rss", listRss).Methods("GET") // List all rss feeds currently subscribed on {serverid}
	r.HandleFunc("/api/{apiKey}", addRss).Methods("POST")
	r.HandleFunc("/api/rss", listAllRss).Methods("GET")
	r.HandleFunc("/api/rss/{apiKey}", listRss).Methods("GET")
}
