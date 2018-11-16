package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

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
		_, err = db.addDiscord(discordServer)
		if err != nil {
			log.Println(err)
		}
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
		_, err = s.ChannelMessageSend(channels[1].ID, "Please configure the bot. Use the following command !configure.")
		if err != nil {
			log.Println(err)
		}
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
