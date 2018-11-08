//Check out github.com/bwmarrin/discordgo for more indepth information about the discord library

package main

import (
	"fmt"
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
		// TODO: Add more commands to the list
		s.ChannelMessageSend(m.ChannelID, "!newrss <link>\n!configure <channel_id>")
	}

	// Try to find a webpages' RSS
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

var tempFeeds = make(map[int]string)

/*
getRSSFeeds gets and lists rss feeds based on a search
*/
func getRSSFeeds(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Split(m.Content, " ")
	if len(words) == 2 {
		links := Crawl(words[1])

		var message string

		if len(links) == 0 {
			message = "No RSS link found on the given webpage: '" + words[1] + "'"
		} else if len(links) == 1 {
			message = "Found one RSS feed: " + links[0] + ".\n"

			ok := db.manageSubscription(links[0], m.GuildID, add)
			if !ok {
				message += "Already subscribed to this RSS feed. So nothing new added."
			} else {
				message += "Added RSS feed to the subscription list."
			}
			s.ChannelMessageSend(m.ChannelID, message)

		} else {
			// Reset map
			tempFeeds = make(map[int]string)
			message = "Found multiple RSS feeds:\n"
			// Max feeds listed per search is currently 20
			for i := 0; i < 20; i++ {
				tempFeeds[i+1] = links[i] // Add feeds to a map temporarily
				message += strconv.Itoa(i+1) + ". " + links[i] + "\n"
			}
			message += "Use !addrss <numbers> to select multiple feeds by putting a space in-between numbers. E.g. \"!addrss 3 7 19\""
		}
		s.ChannelMessageSend(m.ChannelID, message)

		// TODO: Better formatting
	} else {
		s.ChannelMessageSend(m.ChannelID, "Error! Command is of type \"!newrss <link/searchphrase>\"")
	}
}

/*
addRSSFeeds lets the user choose which feeds to subscribe to
*/
func addRSSFeeds(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Split(m.Content, " ")

	var message string

	if len(tempFeeds) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Error! There are no feeds to choose from!")
	} else if len(words) >= 2 {
		// Go through each feed number the user has selected
		for i := 1; i < len(words); i++ {
			num, err := strconv.Atoi(words[i])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error! The numbers you entered were not numbers at all!")
				return
			}
			// Subscribe the server to the RSS feeds
			ok := db.manageSubscription(tempFeeds[num], m.GuildID, add)

			if ok {
				message += "Added " + tempFeeds[num] + " to the subscription list.\n"
			} else {
				message += "Already subscribed to RSS feed " + tempFeeds[num] + "\n"
			}
		}
		s.ChannelMessageSend(m.ChannelID, message)
	} else if len(words) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Error! You did not specify which RSS feeds you want to subscribe to!")
	}
}

/*
removeRSSFeeds lets the user choose which feeds to unsubscribe to
*/
func removeRSSFeeds(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Split(m.Content, " ")

	var message string

	if len(words) >= 2 {
		// TODO: Implement
		s.ChannelMessageSend(m.ChannelID, message)
	} else if len(words) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Error! You did not specify which RSS feeds you want to unsubscribe from!")
	}
}

/*
listRSSFeeds lists all the feeds currently subscribed to
*/
func listRSSFeeds(s *discordgo.Session, m *discordgo.MessageCreate) {
	subscribedTo := getAllSubscribed(m.GuildID)

	if len(subscribedTo) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Nothing to show :(")
		return
	}

	var message string
	message += "Current RSS feeds subscribed to:\n"

	for i, v := range subscribedTo {
		if i >= 20 {
			break
		}

		// Add LastUpdate so the user knows how long since he received last message? (In order to filter out e.g. discontinued RSS's)
		message += strconv.Itoa(i+1) + ". " + v.URL + "\n"
	}
	s.ChannelMessageSend(m.ChannelID, message)
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

// TODO:
// func messageReactions(s *discordgo.Session, channelID, messageID, emojiID string, limit int) (st []*User, err error){
// }

/*
This function will be called (due to AddHandler above) every time a new
guild is joined.
*/
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	//The server is down or deleted
	if event.Guild.Unavailable {
		fmt.Printf("\nWas not able to connect to %s", event.Guild.ID)
		return
	}

	fmt.Println("serverid:", event.Guild.ID)

	var discordServer Discord
	discordServer.ServerID = event.Guild.ID

	//Check if server already exist
	r, err := db.getDiscord(discordServer.ServerID)

	if err != nil && r.ServerID == "" {
		db.addDiscord(discordServer)
	}

}

func guildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	fmt.Println("\nDAMN SON. Got rekt ", event.Guild.ID)

	var discordServer Discord
	discordServer.ServerID = event.Guild.ID
	db.deleteDiscord(discordServer)

}
