package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func Test_addRss(t *testing.T) {
	DBInitTest()
	r := mux.NewRouter()
	r.HandleFunc("/api/{serverid}/rss", addRss).Methods("POST")
	ts := httptest.NewServer(r)
	defer ts.Close()

	url := "https://www.nrk.no/rogaland/toppsaker.rss"
	server := "123456789012345678"

	// Add a new Discord
	var d Discord
	d.ServerID = server
	d, _ = db.addDiscord(d)

	// Test empty body
	resp, err := http.Post(ts.URL+"/api/"+server+"/rss", "application/json", nil)
	if err != nil {
		t.Errorf("Error creating the POST request, %s", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected StatusCode %d, received %d", http.StatusBadRequest, resp.StatusCode)
	}

	// Test malformed URL body
	badTest := "{\"url\":\"https://www.nrk.no/rogaland/toppsaker\"}"

	resp, err = http.Post(ts.URL+"/api/"+server+"/rss", "application/json", strings.NewReader(badTest))
	if err != nil {
		t.Errorf("Error creating the POST request, %s", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		all, _ := ioutil.ReadAll(resp.Body)
		t.Errorf("Expected StatusCode %d, received %d, Body: %s", http.StatusBadRequest, resp.StatusCode, all)
	}

	// Test with proper body, new subscription
	goodTest := "{\"url\":\"" + url + "\"}"
	resp, err = http.Post(ts.URL+"/api/"+server+"/rss", "application/json", strings.NewReader(goodTest))
	if err != nil {
		t.Errorf("Error creating the POST request, %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		all, _ := ioutil.ReadAll(resp.Body)
		t.Errorf("Expected StatusCode %d, received %d, Body: %s", http.StatusOK, resp.StatusCode, all)
	}

	a, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading the body. Got %s", err)
	}

	if !strings.Contains(string(a), "now subscribed") {
		t.Errorf("The body should contain \"now subscribed\" but instead it posts: %s", string(a))
	}

	// Test with proper body, already subscribed
	resp, err = http.Post(ts.URL+"/api/"+server+"/rss", "application/json", strings.NewReader(goodTest))
	if err != nil {
		t.Errorf("Error creating the POST request, %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		all, _ := ioutil.ReadAll(resp.Body)
		t.Errorf("Expected StatusCode %d, received %d, Body: %s", http.StatusOK, resp.StatusCode, all)
	}

	a, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading the body. Got %s", err)
	}

	if !strings.Contains(string(a), "already subscribed") {
		t.Errorf("Error reading the body. Got %s", err)
	}

	// Clean tests
	db.deleteAllDiscord()
	db.deleteAllRSS()
}
func Test_listRss(t *testing.T) {
	DBInitTest()
	r := mux.NewRouter()
	r.HandleFunc("/api/{serverid}/rss", listRss).Methods("GET")
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Reused values
	server := "123456789012345678"
	url1 := "https://www.nrk.no/rogaland/toppsaker.rss"
	url2 := "https://www.nrk.no/finnmark/siste.rss"

	// Add a new Discord
	var d Discord
	d.ServerID = server
	d, _ = db.addDiscord(d)

	// Add a couple RSS and subscribe the server to both of them
	r1, _ := db.addRSS(url1)
	r2, _ := db.addRSS(url2)
	r1.DiscordServers = append(r1.DiscordServers, server)
	r2.DiscordServers = append(r2.DiscordServers, server)
	db.updateRSS(r1)
	db.updateRSS(r2)

	resp, err := http.Get(ts.URL + "/api/" + server + "/rss")
	if err != nil {
		t.Errorf("Error making the GET request, %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected StatusCode %d, received %d", http.StatusOK, resp.StatusCode)
		return
	}

	var a []RSS
	err = json.NewDecoder(resp.Body).Decode(&a)
	if len(a) != 2 {
		t.Errorf("Expected two elements, got %v", len(a))
	}

	// Clean tests
	db.deleteAllDiscord()
	db.deleteAllRSS()
}
func Test_listAllRss(t *testing.T) {
	DBInitTest()
	r := mux.NewRouter()
	r.HandleFunc("/api/rss", listAllRss).Methods("GET")
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Reused values
	url1 := "https://www.nrk.no/rogaland/toppsaker.rss"
	url2 := "https://www.nrk.no/finnmark/siste.rss"

	// Add a couple RSS
	db.addRSS(url1)
	db.addRSS(url2)

	resp, err := http.Get(ts.URL + "/api/rss")
	if err != nil {
		t.Errorf("Error making the GET request, %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected StatusCode %d, received %d", http.StatusOK, resp.StatusCode)
		return
	}

	var a []RSS
	err = json.NewDecoder(resp.Body).Decode(&a)
	if len(a) != 2 {
		t.Errorf("Expected two elements, got %v", len(a))
	}

	// Clean tests
	db.deleteAllDiscord()
	db.deleteAllRSS()
}
