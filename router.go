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
	r.HandleFunc("/api/rss", addRss).Methods("POST")
	r.HandleFunc("/api/rss/{id}", listRss).Methods("GET")

}
