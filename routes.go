package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

/*
addRss gets your discord servers apiKey from path. The user posts an rss url and his server gets subscribed to it
*/
func addRss(w http.ResponseWriter, r *http.Request) {
	// Get the API Key from the URL
	apiKey := mux.Vars(r)["apiKey"]

	// Find the discord server with said apiKey
	discord, err := db.getDiscordFromAPI(apiKey)
	if err != nil {
		http.Error(w, "This api key was not found", http.StatusNotFound)
		return
	}

	// Read the URL from the request
	var u RSS
	err = json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, "Could not read the request. Make sure you have proper json formating eg \n{\n\t\"url\": \"nrk.no/rss/.rss\"\n}", http.StatusBadRequest)
		return
	}

	// Ensure you got a proper URL
	if !isPageRSS(u.URL) {
		http.Error(w, "Invalid request. Make sure you have proper json formating eg \n{\n\t\"url\": \"nrk.no/rss/.rss\"\n}", http.StatusBadRequest)
		return
	}

	// Check if the RSS exists in the database
	rss, _ := db.getRSS(u.URL)
	// If the RSS exists, add the discord server subscription
	if rss.LastUpdate != 0 {
		success := db.manageSubscription(rss.URL, discord.ServerID, add)
		if !success {
			_, err = fmt.Fprintf(w, "%v is already subscribed to %v\n", discord.ServerID, rss.URL)
			if err != nil {
				http.Error(w, "Couldn't print the result", http.StatusBadRequest)
				return
			}
			return
		}
		_, err = fmt.Fprintf(w, "%v is now subscribed to %v\n", discord.ServerID, rss.URL)
		if err != nil {
			http.Error(w, "Couldn't print the result", http.StatusBadRequest)
			return
		}
		return
	}

	// If the RSS don't exist, create it
	rss, err = db.addRSS(u.URL)
	if err != nil {
		fmt.Println(err)
	}
	db.manageSubscription(rss.URL, discord.ServerID, add)

	_, err = fmt.Fprintf(w, "%v is now subscribed to %v\n", discord.ServerID, rss.URL)
	if err != nil {
		http.Error(w, "Couldn't print the result", http.StatusBadRequest)
		return
	}

}

/*
listRss lists the url of every rss file from the database
*/
func listAllRss(w http.ResponseWriter, r *http.Request) {
	rssList, err := db.getAllRSS()
	if err != nil {
		fmt.Println(err)
	}
	var urls []string
	for _, i := range rssList {
		urls = append(urls, i.URL)
	}
	err = json.NewEncoder(w).Encode(urls)
	if err != nil {
		fmt.Println(err)
	}
}

/*
listRss gets your discord servers apiKey from path.
It then posts all the rss feeds the discord server in question is subscribed to
*/
func listRss(w http.ResponseWriter, r *http.Request) {
	// Get the API Key from the URL
	apiKey := mux.Vars(r)["apiKey"]
	discord, err := db.getDiscordFromAPI(apiKey)
	if err != nil {
		http.Error(w, "This api key was not found", http.StatusNotFound)
		return
	}

	rssList := getAllSubscribed(discord.ServerID)
	var urls []string
	for _, i := range rssList {
		urls = append(urls, i.URL)
	}
	err = json.NewEncoder(w).Encode(urls)
	if err != nil {
		fmt.Println(err)
	}
}
