//Check out github.com/bwmarrin/discordgo for more indepth information about the discord library

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

//This may very well be horribly bad <-- Yup
var GlobalSession *discordgo.Session

func main() {
	// Initialize the database
	DBInit()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)

	// Register guildCreate as a callback for the guildDelete events.
	dg.AddHandler(guildDelete)

	stop := schedule(scanAndPost, 3*time.Minute)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Discord bot is now running.  Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	stop <- true

	// Cleanly close down the Discord session.
	dg.Close()

}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateStatus(0, "!commands")

	GlobalSession = s

}

/*
This function will be called (due to AddHandler above) every time a new
message is created on any channel that the autenticated bot has access to.
*/
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		log.Println("Error while trying to retrieve channel in messageCreate(), " + err.Error())
	}

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	/*
		Commands only allowed in a server
	*/
	if channel.Type != discordgo.ChannelTypeDM {
		if strings.HasPrefix(strings.ToLower(m.Content), "!newrss") {
			getRSSFeeds(s, m)
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!addrss") {
			addRSSFeeds(s, m)
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!remrss") {
			removeRSSFeeds(s, m)
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!listrss") {
			listRSSFeeds(s, m)
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!newkeyrss") ||
			strings.HasPrefix(strings.ToLower(m.Content), "!getkeyrss") {
			manageAPIKey(s, m)
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!configure") {
			configure(s, m)
		}
	}

	/*
		Commands allowed in server as well as DM
	*/
	if strings.HasPrefix(strings.ToLower(m.Content), "!commands") {
		// TODO: Add more commands to the list
		s.ChannelMessageSend(m.ChannelID, `
`+"**Basic Commands:**"+`
!commands					`+"\t\t\t\t"+`#Get a list of commands
!configure <channel_id>		`+"\t"+` #Set a new channel where the bot should post RSS updates. Default: First channel in server.
!newrss <keyword/link>		`+"\t"+`#Subscribe to a new RSS feed using a keyword (e.g. bbc) or link (e.g. http://feeds.bbci.co.uk/news/rss.xml)
!addrss <link/number(s)>	`+"\t"+` #If more than one RSS link was found by the bot, you can choose feeds
!listrss					`+"\t\t\t\t\t\t"+`#Get a list of all feeds your server is subscribed to.
!remrss<link/number(s)>	`+"\t"+` #(Call \"!listrss\" first) Remove RSS subscription(s).

`+"**WebAPI (Only for server owner):**"+`
!newkeyrss				    `+"\t\t\t\t"+`#Get a new API key to use for the webAPI. (NOTE: this will replace the old one)
!getkeyrss					`+"\t\t\t\t"+`  #Get the current API key.
	`)
	}
}

// TODO:
// func messageReactions(s *discordgo.Session, channelID, messageID, emojiID string, limit int) (st []*User, err error){
// }

/*
We need to replace all data with the data from the json struct.
We also do not want to use the m variable. Use channel id from db
*/
func embedMessage(s *discordgo.Session, channelid string, rss Channel) {
	var Embed discordgo.MessageEmbed

	Embed.URL = rss.Items[0].Link
	Embed.Title = rss.Items[0].Title

	if len(rss.Items[0].Description) > 1 {
		Embed.Description = rss.Items[0].Description
	} else {
		Embed.Description = "No description is available."
	}

	var EmbedFooter discordgo.MessageEmbedFooter
	if len(rss.Title) > 1 {
		EmbedFooter.Text = rss.Title
	} else {
		EmbedFooter.Text = rss.OriginalRSSLink
	}
	Embed.Footer = &EmbedFooter

	var EmbedImage discordgo.MessageEmbedImage

	if len(rss.Items[0].Enclosure.Url) > 1 { //Standard
		EmbedImage.URL = rss.Items[0].Enclosure.Url
	} else if len(rss.Items[0].Image) > 1 { //Weird websites like VG
		EmbedImage.URL = rss.Items[0].Image
	} else if len(rss.Items[0].Media.Url) > 1 { //"media" container
		EmbedImage.URL = rss.Items[0].Media.Url
	} else {
		//Fallback to website image
		EmbedImage.URL = rss.Image.Url
	}
	Embed.Image = &EmbedImage

	s.ChannelMessageSendEmbed(channelid, &Embed)
}

func RSSListEmbed(s *discordgo.Session, m *discordgo.MessageCreate, rssFeeds []string, numberedFeeds map[int]string, extraInfo string) {

	var RssListEmbed discordgo.MessageEmbed

	RssListEmbed.Title = "Found multiple RSS feeds:"
	//RssListEmbed.Description = "Something about something something goes here"

	for i, link := range rssFeeds {
		var RssListEmbedFields discordgo.MessageEmbedField

		RssListEmbedFields.Value = link
		if numberedFeeds[i+1] != "" {
			RssListEmbedFields.Name = strconv.Itoa(i+1) + ". " + getRSSName(numberedFeeds[i+1])
		} else {
			RssListEmbedFields.Name = strconv.Itoa(i+1) + ". <Empty>"
		}

		RssListEmbed.Fields = append(RssListEmbed.Fields, &RssListEmbedFields)
	}

	var RssListEmbedFooter discordgo.MessageEmbedFooter

	// If the message should show some extra message at the bottom
	if extraInfo != "" {
		RssListEmbedFooter.Text = extraInfo
		RssListEmbed.Footer = &RssListEmbedFooter
	}

	s.ChannelMessageSendEmbed(m.ChannelID, &RssListEmbed)
}
