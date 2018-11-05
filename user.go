package main

// User is added to the database
type User struct {
	UserTag   string `json:"userTag"`
	DiscordID string `json:"discordId"`
	RSSURL    string `json:"rssurl"`
}
