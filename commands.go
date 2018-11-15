package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

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
	extraInfo := "Use !remrss <numbers> to remove multiple feeds by putting a space in-between numbers. E.g. \"!remrss 3 7 19\""
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

	if strings.HasPrefix(strings.ToLower(m.Content), "!newkeyrss") {
		// Generate a new key
		discord.APIKey = generateNewKey()
		err = db.updateDiscord(discord)
		if err != nil {
			log.Println("Something went wrong updating the API key in database, " + err.Error())
		}
		message += "New API key: "
	} else if strings.HasPrefix(strings.ToLower(m.Content), "!getkeyrss") {
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
