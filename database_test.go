package main

import (
	"testing"
)

/*
The following function tests all of these functions that does something with track items in the database
addTrack
countTrack
getTrack
getAllTracks
getTracksAfter
latestTicker
deleteAllTracks
*/

// DBInitTest fills DBInfo for test purposes
func DBInitTest() {
	db.DBURL = "mongodb://test:test123@ds155833.mlab.com:55833/rssbot_test"
	db.DBName = "rssbot_test"
	db.CollectionDiscord = "discord"
	db.CollectionRSS = "rss"
}

/*
Test_Database_Rss tests that all RSS related database operations works as intended, and prepares some data for the discord test
*/
func Test_Database_Rss(t *testing.T) {
	// Make sure the test runs the test database
	DBInitTest()
	// Test that delete all functions work, and clean up the database for the tests
	db.deleteAllDiscord()
	db.deleteAllRSS()

	url := "https://www.vg.no/rss/feed/forsiden/"
	url2 := "https://www.nrk.no/rogaland/toppsaker.rss"
	url3 := "https://www.nrk.no/finnmark/siste.rss"
	server := "506870206505811969"
	server2 := "506870206505811971"

	// Add new RSS
	r, err := db.addRSS(url)
	rr, _ := db.addRSS(url2)
	_, _ = db.addRSS(url3)
	if err != nil {
		t.Errorf("Problem with addRSS(): %v", err)
	}

	// Check that we can't add another with the same URL
	_, err = db.addRSS(url)
	if err == nil {
		t.Errorf("Expected an error when adding an already existing URL")
	}

	// Check that we added it correctly and got data from the URL
	if r.LastUpdate == 0 || r.URL != url {
		t.Errorf("LastUpdate should not be 0, and got %v\n URL should be %v and got %v", r.LastUpdate, url, r.URL)
	}

	// Check that we do indeed have 3 element now
	i, err := db.countRSS()
	if i != 3 || err != nil {
		t.Errorf("Expected 3 element, got %v: %v", i, err)
	}

	// Add a discord server to the RSS and update it
	r.DiscordServers = append(r.DiscordServers, server)
	rr.DiscordServers = append(rr.DiscordServers, server)
	rr.DiscordServers = append(rr.DiscordServers, server2)
	err = db.updateRSS(r)
	_ = db.updateRSS(rr)
	if err != nil {
		t.Errorf("Problem with updateRSS(): %v", err)
	}

	// Check that we can get the element and make sure the values are correct
	r2, err := db.getRSS(r.URL)
	if err != nil || r.LastUpdate != r2.LastUpdate || r.URL != r2.URL || r.DiscordServers[0] != "506870206505811969" {
		t.Errorf("Problem with getRSS(): %v\nExpected %v, got %v\nExpected %v, got %v\nExpected %v, got %v", err, r.LastUpdate, r2.LastUpdate, r.URL, r2.URL, "506870206505811969", r.DiscordServers[0])
	}

	// Check that delete works
	err = db.deleteRSS(url3)
	if err != nil {
		t.Errorf("Problem with deleteRSS(): %v", err)
	}

	// Check that we do indeed have 2 element now
	i, err = db.countRSS()
	if i != 2 || err != nil {
		t.Errorf("Expected 2 element, got %v: %v", i, err)
	}
}

/*
	Test_Database_Discord is a direct continuation of Test_Database_Rss and relies on the database files created in it
*/
func Test_Database_Discord(t *testing.T) {
	url2 := "https://www.nrk.no/rogaland/toppsaker.rss"
	server := "506870206505811969"

	// Add a new Discord
	var d Discord
	d.ServerID = server
	d, err := db.addDiscord(d)
	if err != nil {
		t.Errorf("Problem with addDiscord(): %v", err)
	}

	// Check that we do indeed have 1 element now
	i, err := db.countDiscord()
	if i != 1 || err != nil {
		t.Errorf("Expected 1 element, got %v: %v", i, err)
	}

	// Add a discord server to the RSS and update it
	d.ChannelID = "506885726546296834"
	err = db.updateDiscord(d)
	if err != nil {
		t.Errorf("Problem with updateRSS(): %v", err)
	}

	// Check that we can get the element and make sure the values are correct
	d2, err := db.getDiscord(server)
	if err != nil || d.ChannelID != d2.ChannelID || d.ServerID != d2.ServerID {
		t.Errorf("Expected %v got %v \nExpected %v got %v", d.ChannelID, d2.ChannelID, d.ServerID, d2.ServerID)
	}

	// Check that we can delete a discord, that it gets deleted from RSS discordServer array and that the RSS file gets deleted if it reaches 0 servers
	err = db.deleteDiscord(d)
	if err != nil {
		t.Errorf("Problem with deleteDiscord(): %v", err)
	}
	// Check that there's no discord servers left
	i, err = db.countDiscord()
	if i != 0 || err != nil {
		t.Errorf("Expected 0 elements, got %v: %v", i, err)
	}
	// Check that we do indeed have 1 RSS element left now (1 deleted with the server)
	i, err = db.countRSS()
	if i != 1 || err != nil {
		t.Errorf("Expected 1 element, got %v: %v", i, err)
	}
	// Check that there's only one server left for rr (was two before deleteDiscord)
	rr, _ := db.getRSS(url2)
	if len(rr.DiscordServers) != 1 {
		t.Errorf("Expected 1 element, got %v: %v", i, err)
	}

	// Clean out the reminder
	err = db.deleteRSS(url2)
	if err != nil {
		t.Errorf("Problem with deleteRSS(): %v", err)
	}
}
