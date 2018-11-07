package main

/*
Discord
*/
type Discord struct {
	ServerID  string `json:"serverId" bson:"serverId"`
	ChannelID string `json:"channelId" bson:"channelId"`
}
