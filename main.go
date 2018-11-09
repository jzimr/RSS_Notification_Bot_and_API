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

	"github.com/bwmarrin/discordgo"
)

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

	// Register guildCreate as a callback for the guildDelte events.
	dg.AddHandler(guildDelete)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Discord bot is now running.  Press CTRL-C to exit.")

	//Init and start the router
	routerInit()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()

}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateStatus(0, "!commands")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!commands") { //Example
		s.ChannelMessageSend(m.ChannelID, "!newrss <link>\n!configure <channel_id>")
	}

	// Try to find a webpages' RSS
	if strings.HasPrefix(strings.ToLower(m.Content), "!newrss") {
		getRSSFeeds(s, m)
	}

	//Only for testing purposes. Need to change later.
	if strings.HasPrefix(strings.ToLower(m.Content), "!testembed") {

		//We want to pass a json struct with the function
		embedMessage(s, m)

	}

	if strings.HasPrefix(strings.ToLower(m.Content), "!configure") {
		words := strings.Split(m.Content, " ")
		if len(words) != 2 {
			s.ChannelMessageSend(m.ChannelID, "Invalid syntax. !configure <channel_id> or !configure <channel_name>")
			return
		}

		var index int
		index = -1
		channels, err := s.GuildChannels(m.GuildID)
		if err != nil {
			fmt.Println(err, " something went horribly wrong")
		}
		for i := range channels {
			if channels[i].ID == words[1] || channels[i].Name == words[1] {
				index = i
				break
			}
		}

		if index == -1 {
			s.ChannelMessageSend(m.ChannelID, "Invalid channel id or channel name. Try again")
			return
		}

		var discordServer Discord
		discordServer.ServerID = m.GuildID
		discordServer.ChannelID = channels[index].ID

		db.updateDiscord(discordServer)
		s.ChannelMessageSend(m.ChannelID, "Text channel with name "+channels[index].Name+" are now set as the default notification channel.")

	}

}

/*
getRSSFeeds gets rss feeds and lets the user choose which feeds to
subscribe to
*/
func getRSSFeeds(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Split(m.Content, " ")
	if len(words) == 2 {
		links, err := Crawl(words[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error()) // Perhaps change to log.print instead?
			return
		}

		var message string

		if len(links) == 1 {
			message = "Found a RSS feed: " + links[0]
		} else {
			message = "Found multiple RSS feeds:\n"
			for i := 0; i < 20; i++ {
				message += strconv.Itoa(i+1) + ". " + links[i] + "\n"
			}
			message += "Select multiple feeds by putting a space in-between numbers. E.g. 1 10 23"
		}
		s.ChannelMessageSend(m.ChannelID, message)

		// TODO: Better formatting
		// Add feature to listen to what user types (Maybe prefix should be something like /sub [ids]?)

	} //else statement to give user feedback?
}

//We need to replace all data with the data from the json struct
//We also do not want to use the m variable. Use channel id from db
func embedMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	var testEmbed discordgo.MessageEmbed
	testEmbed.Color = 245 //This should be changed
	testEmbed.URL = "http://localhost"
	testEmbed.Title = "A title of something goes here"
	testEmbed.Description = "Something about something something goes here"

	var testEmbedFooter discordgo.MessageEmbedFooter
	testEmbedFooter.Text = "Article pulled from XXX.XXX"
	testEmbed.Footer = &testEmbedFooter

	var testEmbedImage discordgo.MessageEmbedImage
	testEmbed.Image = &testEmbedImage
	testEmbedImage.URL = "https://i.imgur.com/aSVjtu7.png"

	s.ChannelMessageSendEmbed(m.ChannelID, &testEmbed)
}

/*
This function will be called (due to AddHandler above) every time a new
guild is joined.
*/
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	//The server is down or deleted
	if event.Guild.Unavailable {
		log.Printf("\nWas not able to connect to %s", event.Guild.ID)
		return
	}

	log.Println("serverid:", event.Guild.ID)

	var discordServer Discord
	discordServer.ServerID = event.Guild.ID

	//Check if server already exist
	r, err := db.getDiscord(discordServer.ServerID)

	if err != nil && r.ServerID == "" {
		db.addDiscord(discordServer)
	}

}

func guildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	log.Println("\nServer was deleted OR i was kicked from ", event.Guild.ID)

	var discordServer Discord
	discordServer.ServerID = event.Guild.ID
	db.deleteDiscord(discordServer)

}
