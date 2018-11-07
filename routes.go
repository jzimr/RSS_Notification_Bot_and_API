package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//Post to ChannelID
func addRss(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)

	// Connect to database
	// If (we are already subscribed to the rss)
	////// return
	// If (database has this rss link already)
	////// add serverid to the array in the db collection
	// else
	////// create new document with the rss link and add serverid to array

	// If success: 201 response code

	// E.g. JSON POST Format:
	/*
		{
			"rss_link": "nrk.no/rss/.rss"
		}
	*/

	err := json.NewEncoder(w).Encode("Not yet implemented")
	if err != nil {
		fmt.Println(err)
	}
}

//Get by ChannelID
func listRss(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)

	// Connect to database
	// for (rss collection)
	////// add rss link to array
	// Respond in json format

	err := json.NewEncoder(w).Encode("Not yet implemented")
	if err != nil {
		fmt.Println(err)
	}
}

//Delete by ID

//Webhooks?
