//Check out github.com/bwmarrin/discordgo for more indepth information about the discord library

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
)

//This may very well be horribly bad <-- Yup
var GlobalSession *discordgo.Session

func main() {
	// Initialize the database
	DBInit()

	// API Setup
	r := mux.NewRouter()
	r.HandleFunc("/api/{apiKey}", addRss).Methods("POST")
	r.HandleFunc("/api/rss", listAllRss).Methods("GET")
	r.HandleFunc("/api/rss/{apiKey}", listRss).Methods("GET")
	http.ListenAndServe(":8080", r)

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
		if strings.HasPrefix(strings.ToLower(m.Content), "!rssnewkey") ||
			strings.HasPrefix(strings.ToLower(m.Content), "!rssgetkey") {
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
		s.ChannelMessageSend(m.ChannelID, "!newrss <link>\n!configure <channel_id>")
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
			tempFeeds = make(map[int]string)

			used = 1

			// Max feeds listed per search is currently 20
			if len(links) > 20 {
				links = links[:20]
			}

			for i := range links {
				tempFeeds[i+1] = links[i] // Add feeds to a map temporarily

			}
		}
		if used == 0 {
			s.ChannelMessageSend(m.ChannelID, message)
		} else {
			RSSListEmbed(s, m, links)
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
	// Show a list of feeds to choose from
	listRSSFeeds(s, m)

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

func RSSListEmbed(s *discordgo.Session, m *discordgo.MessageCreate, links []string) {

	var RssListEmbed discordgo.MessageEmbed

	RssListEmbed.Title = "Found multiple RSS feeds:"
	//RssListEmbed.Description = "Something about something something goes here"

	for i := range links {
		var RssListEmbedFields discordgo.MessageEmbedField
		RssListEmbedFields.Name = strconv.Itoa(i + 1)
		RssListEmbedFields.Value = links[i]
		RssListEmbed.Fields = append(RssListEmbed.Fields, &RssListEmbedFields)
	}

	var RssListEmbedFooter discordgo.MessageEmbedFooter
	RssListEmbedFooter.Text = "Use !addrss <numbers> to select multiple feeds by putting a space in-between numbers. E.g. \"!addrss 3 7 19\""
	RssListEmbed.Footer = &RssListEmbedFooter

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
	log.Println("Server was deleted OR i was kicked from ", event.Guild.ID)

	var discordServer Discord
	discordServer.ServerID = event.Guild.ID

	err := db.deleteDiscord(discordServer)
	if err != nil {
		log.Println(err)
	}
}
