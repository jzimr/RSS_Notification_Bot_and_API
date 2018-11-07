//Check out github.com/bwmarrin/discordgo for more indepth information about the discord library

package main

import (
	"fmt"
	"os"
	"os/signal"
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
		words := strings.Split(m.Content, " ")
		if len(words) == 2 {
			link, err := Crawl(words[1])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error()) // Perhaps change to log.print instead?
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Found a RSS link: "+link)
		} //else statement to give user feedback?
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

		s.ChannelMessageSend(m.ChannelID, "Text channel with name "+channels[index].Name+" are now set as the default notification channel.")
		var discordServer Discord
		discordServer.ServerID = m.GuildID
		discordServer.ChannelID = channels[index].ID

		db.updateDiscord(discordServer)

	}

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

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	//The server is down or deleted
	if event.Guild.Unavailable {
		fmt.Printf("\nWas not able to connect to %s", event.Guild.ID)
		return
	}

	var discordServer Discord
	discordServer.ServerID = event.Guild.ID
	db.addDiscord(discordServer)

	/*
		channels, err := s.GuildChannels(event.Guild.ID)
		if err != nil {
			fmt.Println(err, " something went horribly wrong")
		}
		for i := range channels {
			fmt.Println("\n", channels[i].ID, channels[i].Name)
		}
	*/
}

func guildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	fmt.Println("\nDAMN SON. Got rekt ", event.Guild.ID)
}
