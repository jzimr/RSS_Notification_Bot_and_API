package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"time"
)

// RSS ,,,
type RSS struct {
	URL            string    `json:"url" bson:"url"`
	LastUpdate     time.Time `json:"lastUpdate" bson:"lastUpdate"`
	DiscordServers []string  `json:"discordServers" bson:"discordServers"`
}

/*
Item holds articles or anything else the RSS file consists of
It's reached through Channel
*/
type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Enclosure   string `xml:"enclosure"` // Relevant image
	PubDate     string `xml:"pubDate"`
}

/*
Channel is used for parsing the RSS file.
*/
type Channel struct {
	LastBuildDate string `xml:"channel>lastBuildDate"`
	Items         []Item `xml:"channel>item"`
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

	return channel
}

/*
postRSS goes through each discord server that subscribes to a RSS and sends a message to it
*/
func postRSS(RSS string) {
	c := readRSS(RSS)

	// Convert string to time
	lastBuild := toTime(c.LastBuildDate)

	r := db.getRSS(RSS)

	if r.LastUpdate != lastBuild {
		for _, server := range r.DiscordServers {
			// NOT FINISHED
			// Post to discord servers here
			discord, err := db.getDiscord(server)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("Channel ID: %v", discord.ChannelID)
		}
		r.LastUpdate = lastBuild
		db.updateRSS(r)
	}
}

/*
toTime converts from the RFC1123 format to time.Time
*/
func toTime(s string) time.Time {
	layout := "Mon, 02 Jan 2006 15:04:05 MST"

	t, err := time.Parse(layout, s)
	if err != nil {
		fmt.Println(err)
	}
	return t
}
