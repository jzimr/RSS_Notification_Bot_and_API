package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	//Works for most sites. Not all
	Enclosure struct {
		Url string `xml:"url,attr"`
	} `xml:"enclosure"`

	//NYTIMES and others?
	Media struct {
		Url string `xml:"url,attr"`
	} `xml:"media:container"`

	//VG specific?
	Image string `xml:"image"`

	PubDate string `xml:"pubDate"`
}

/*
Channel is used for parsing the RSS file.
*/
type Channel struct {
	Title         string `xml:"channel>title"`
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
	log.Println(RSS)
	// Convert string to time
	lastBuild, err := toTime(c.LastBuildDate)
	if err != nil {
		log.Println("Error in postRss()", RSS, err)
		return
	}

	r, err := db.getRSS(RSS)
	if err != nil {
		fmt.Printf("%v", err.Error())
	}

	//This if statement does currently not work

	//HACKY FIX
	log.Println(r.LastUpdate.String())
	//Remove GMT/UTC suffix
	parsedTime := r.LastUpdate.String()[:len(r.LastUpdate.String())-4]
	//Remove extra timezone info
	//if strings.()

	if !strings.HasPrefix(lastBuild.String(), parsedTime) {
		for _, server := range r.DiscordServers {
			// NOT FINISHED
			// Post to discord servers here

			discord, err := db.getDiscord(server)
			if err != nil {
				fmt.Println(err)
			}

			//Forward channel to function which sends an embeded message to the correct discord channel
			embedMessage(GlobalSession, discord.ChannelID, c)
			log.Printf("Channel ID: %v", discord.ChannelID)
		}
		r.LastUpdate = lastBuild
		db.updateRSS(r)
	}
}

/*
toTime converts from the RFC1123 format to time.Time
*/
func toTime(s string) (time.Time, error) {

	//Remove extra space on the right side
	newS := strings.TrimRight(s, " ")

	t, err := time.Parse(time.RFC1123, newS)
	if err != nil {
		t, err = time.Parse(time.RFC1123Z, newS)
		if err != nil {
			return t, err
		}
	}
	return t, nil
}
