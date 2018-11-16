package main

import (
	"encoding/xml"
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
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`

	//Works for most sites. Not all
	Enclosure struct {
		URL string `xml:"url,attr"`
	} `xml:"enclosure"`

	//NYTIMES and others. DOES NOT WORK ATM
	Media struct {
		URL string `xml:"url,attr"`
	} `xml:"media:content"`

	//VG specific?
	Image         string `xml:"image"`
	PubDateString string `xml:"pubDate"`
}

/*
Channel is used for parsing the RSS file.
*/
type Channel struct {
	OriginalRSSLink string //FallbackMethod
	Title           string `xml:"channel>title"`
	LastUpdate      int64

	Image struct {
		URL string `xml:"url"`
	} `xml:"channel>image"`

	Items []Item `xml:"channel>item"`
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
getRSSName returns the name of a RSS feed
*/
func getRSSName(link string) (name string) {
	channel := readRSS(link)

	return channel.Title
}

/*
postRSS goes through each discord server that subscribes to a RSS and sends a message to it
*/
func postRSS(RSS string) {
	// c is the latest data from the URL
	c := readRSS(RSS)

	// r is used to see when we last sent the message from this URL
	r, err := db.getRSS(RSS)
	if err != nil {
		log.Println(err)
	}
	var discord Discord
	// If the the latest update in the RSS file (c) is not the same as the latest update we sent out (r)
	if c.LastUpdate != r.LastUpdate {
		for _, server := range r.DiscordServers {
			discord, err = db.getDiscord(server)
			if err != nil {
				log.Println(err)
			}

			//Forward channel to function which sends an embeded message to the correct discord channel
			embedMessage(GlobalSession, discord.ChannelID, c)
		}
		r.LastUpdate = c.LastUpdate
		err = db.updateRSS(r)
		if err != nil {
			log.Println(err)
		}
	}
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
