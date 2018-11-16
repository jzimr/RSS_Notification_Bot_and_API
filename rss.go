package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

/*
RSS holds the URL of an RSS site, an array of all servers subscribed to it
and a timestamp telling when the last time it was sent to subscribers was
*/
type RSS struct {
	URL            string   `json:"url" bson:"url"`
	LastUpdate     int64    `json:"lastUpdate" bson:"lastUpdate"`
	DiscordServers []string `json:"discordServers" bson:"discordServers"`
}

/*
Item holds articles or anything else the RSS file consists of
It's reached through Channel
*/
type Item struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	PubDateString string `xml:"pubDate"`
}

/*
Channel is used for parsing the RSS file.
*/
type Channel struct {
	OriginalRSSLink string //FallbackMethod
	Title           string `xml:"channel>title"`
	LastUpdate      int64
	Items           []Item `xml:"channel>item"`
}

/*
readRSS takes an RSS file as a parameter and parses it.
It then checks if it's been made any changes since last check,
	and posts to subscribed servers if it has been.
*/
func readRSS(RSS string) Channel {
	// Reads the RSS file
	resp, err := http.Get(RSS)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	// Parses the RSS file to a channel with items
	var channel Channel
	if err = xml.NewDecoder(resp.Body).Decode(&channel); err != nil {
		log.Fatalln(err)
	}

	// Sets the last update time to that of the latest article
	channel.LastUpdate, err = toTime(channel.Items[0].PubDateString)
	if err != nil {
		log.Println("Error in readRSS()", RSS, err)
	}

	// Tries to fix a bug by trimming the article URL of bonus characters
	channel.Items[0].Link = stringTrim(channel.Items[0].Link)
	channel.OriginalRSSLink = RSS

	return channel
}

/*
toTime converts from the RFC1123(z) format to timestamp
*/
func toTime(s string) (int64, error) {

	//Remove potential extra characters
	newS := stringTrim(s)

	t, err := time.Parse(time.RFC1123, newS)
	if err != nil {
		t, err = time.Parse(time.RFC1123Z, newS)
		if err != nil {
			return t.Unix(), err
		}
	}

	return t.Unix(), nil
}

/*
stringTrim removes unwanted characters
*/
func stringTrim(s string) string {
	s = strings.TrimLeft(s, "\"")
	s = strings.TrimRight(s, "\"")
	s = strings.TrimRight(s, "\n")
	s = strings.TrimRight(s, " ")

	return s
}

/*
Checks whether a given webpage (URL) is of RSS format
*/
func isPageRSS(URL string) (isRSS bool) {
	// Make GET request
	baseClient := &http.Client{} //
	req, _ := http.NewRequest("GET", URL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	resp, err := baseClient.Do(req)
	if err != nil {
		log.Println("An error occured while trying to make GET requestsssss, " + err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	doctype := string(body)[0:16] // Fixes error on some RSS feeds

	return strings.Contains(doctype, "<?xml version")

}
