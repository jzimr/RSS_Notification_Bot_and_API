package main

/*
Discord
*/
type Discord struct {
	ServerID  string `json:"serverid" bson:"serverid"`
	ChannelID string `json:"channelid" bson:"channelid"`
	APIKey    string `json:"apikey" bson:"apikey"`
}
