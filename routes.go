package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func addRss(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode("Not yet implemented")
	if err != nil {
		fmt.Println(err)
	}
}

func listRss(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode("Not yet implemented")
	if err != nil {
		fmt.Println(err)
	}
}
