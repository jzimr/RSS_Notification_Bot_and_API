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

	//Init and start the router
	routerInit()

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

	//Set the GlobalSession to be equal to current Discord Session.
	GlobalSession = s

}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
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

// Key = ID, Value = Link
var availableFeeds = make(map[int]string)

/*
getRSSFeeds gets and lists rss feeds based on a search
*/
func getRSSFeeds(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Split(m.Content, " ")
	if len(words) == 2 {
		links := Crawl(words[1]) // List of all RSS links

		var used = 0
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
		} else {
			// Reset map
			availableFeeds = make(map[int]string)

			used = 1

			// Max feeds listed per search is currently 20
			if len(links) > 20 {
				links = links[:20]
			}

			for i := range links {
				availableFeeds[i+1] = links[i] // Add feeds to a map temporarily
			}
		}
		if used == 0 {
			s.ChannelMessageSend(m.ChannelID, message)
		} else {
			// linksAndNames := getRSSNamesAndLinks(links) // Map of RSS links (Key) and names (Value)
			extraInfo := "Use !addrss <numbers> to select multiple feeds by putting a space in-between numbers. E.g. \"!addrss 3 7 19\""
			RSSListEmbed(s, m, links, availableFeeds, extraInfo)
		}
		// TODO: Better formatting
	} else {
		s.ChannelMessageSend(m.ChannelID, "Error! Command is of type \"!newrss <link/searchphrase>\"")
	}
}

/*
	configure lets the user choose what text channel the bot should post rss updates to
*/
func configure(s *discordgo.Session, m *discordgo.MessageCreate) {
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

/*
addRSSFeeds lets the user choose which feeds to subscribe to
*/
func addRSSFeeds(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Split(m.Content, " ")

	var message string

	if len(availableFeeds) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Error! There are no feeds to choose from!")
	} else if len(words) >= 2 {
		// Go through each feed number the user has selected
		for i := 1; i < len(words); i++ {
			num, err := strconv.Atoi(words[i])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error! The number(s) you entered were not numbers at all!")
				return
			}
			// Subscribe the server to the RSS feeds
			ok := db.manageSubscription(availableFeeds[num], m.GuildID, add)

			if ok {
				message += "Added " + availableFeeds[num] + " to the subscription list.\n"
			} else {
				message += "Already subscribed to RSS feed " + availableFeeds[num] + "\n"
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
		for i := 1; i < len(words); i++ {
			num, err := strconv.Atoi(words[i])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error! The number(s) you entered were not numbers at all!")
				return
			}
			// Subscribe the server to the RSS feeds
			ok := db.manageSubscription(subbedFeeds[num], m.GuildID, remove)

			if ok {
				message += "Removed " + subbedFeeds[num] + " from the subscription list.\n"
			} else {
				message += "You are not subscribed to " + subbedFeeds[num] + "\n"
			}
		}
		s.ChannelMessageSend(m.ChannelID, message)
	} else if len(words) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Error! You did not specify which RSS feeds you want to unsubscribe from!")
	}
}

// Key = ID, Value = Name/Link
var subbedFeeds = make(map[int]string)

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

	// Reset map
	subbedFeeds = make(map[int]string)

	// Fix links
	var links []string
	for i, v := range subscribedTo {
		if i >= 20 {
			break
		}
		links = append(links, v.URL)
	}

	for i := range links {
		subbedFeeds[i+1] = links[i] // Add feeds to a map temporarily
	}

	// Build embedded message
	extraInfo := "Use !remrss <numbers> to remove multiple feeds by putting a space in-between numbers. E.g. \"!addrss 3 7 19\""
	RSSListEmbed(s, m, links, subbedFeeds, extraInfo)
}

/*
manageAPIKey generates or gets an API key
*/
func manageAPIKey(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		log.Println("Something went wrong trying to get guild, " + err.Error())
	}

	// Only server owner is allowed to create API keys
	if m.Author.ID != guild.OwnerID {
		s.ChannelMessageSend(m.ChannelID, "Only the owner of the server has permission to this command.")
		return
	}

	discord, err := db.getDiscord(m.GuildID)
	if err != nil {
		log.Println("Something went wrong getting serverID from databse, " + err.Error())
	}

	var message string

	if strings.HasPrefix(strings.ToLower(m.Content), "!rssnewkey") {
		// Generate a new key
		discord.APIKey = generateNewKey()
		err = db.updateDiscord(discord)
		if err != nil {
			log.Println("Something went wrong updating the API key in database, " + err.Error())
		}
		message += "New API key: "
	} else if strings.HasPrefix(strings.ToLower(m.Content), "!rssgetkey") {
		// Get key
		message += "API key: "
	}

	// Establish DM channel with owner of server
	channel, err := s.UserChannelCreate(guild.OwnerID)
	if err != nil {
		log.Println("Something went wrong whilst trying to create a DM, " + err.Error())
	}

	// Send a private message to the owner so he can use it
	s.ChannelMessageSend(channel.ID, message+discord.APIKey)
}

//We need to replace all data with the data from the json struct
//We also do not want to use the m variable. Use channel id from db
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
		log.Printf("\nWas not able to connect to %s", event.Guild.ID)
		return
	}

	var discordServer Discord
	discordServer.ServerID = event.Guild.ID

	//Check if server already exist
	r, err := db.getDiscord(discordServer.ServerID)

	if err != nil && r.ServerID == "" {
		db.addDiscord(discordServer)
	}

	if r.ChannelID == "" { //Currently no channel set
		channels, err := s.GuildChannels(event.Guild.ID)
		if err != nil {
			log.Println(err)
		}
		discordServer.ChannelID = channels[1].ID

		err = db.updateDiscord(discordServer)
		if err != nil {
			log.Println("db", err)
		}
		//Post to first channel that the bot needs to be configured
	}

}

func guildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	log.Println("Server was deleted OR i was kicked from ", event.Guild.ID)

	var discordServer Discord
	discordServer.ServerID = event.Guild.ID

	err := db.deleteDiscord(discordServer)
	if err != nil {
		log.Println(err)
	}
}
