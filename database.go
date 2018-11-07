package main

import (
	"fmt"
	"strconv"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// DBInfo stores the details of the DB connection
type DBInfo struct {
	DBURL             string
	DBName            string
	CollectionDiscord string
	CollectionRSS     string
}

// db stores the credentials of our database
var db DBInfo

// DBInit fills DBInfo with the information about our database
func DBInit() {
	db.DBURL = "mongodb://rssbot:rssbot1@ds253243.mlab.com:53243/rssbot"
	db.DBName = "rssbot"
	db.CollectionDiscord = "Discord"
	db.CollectionRSS = "RSS"
}

/*
--------------------------------------------Discord--------------------------------------------
- addDiscord(d Discord) Discord
- getDiscord(s string) Discord
- deleteDiscord(d Discord)
- updateDiscord(d Discord)
--------------------------------------------Discord--------------------------------------------
*/

/*
addDiscord adds
*/
func (db *DBInfo) addDiscord(d Discord) Discord {

	// Creates a connection
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	// Inserts the user into the database
	err = session.DB(db.DBName).C(db.CollectionDiscord).Insert(d)
	if err != nil {
		fmt.Printf("Error in addDiscord(): %v", err.Error())
	}
	return d
}

/*
getDiscord gets
*/
func (db *DBInfo) getDiscord(s string) (Discord, error) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	d := Discord{}
	err = session.DB(db.DBName).C(db.CollectionDiscord).Find(bson.M{"serverid": s}).One(&d)

	//Let other functions handle errors
	return d, err
}

/*
deleteDiscord deletes
*/
func (db *DBInfo) deleteDiscord(d Discord) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DBName).C(db.CollectionDiscord).Remove(bson.M{"serverid": d.ServerID})
	if err != nil {
		fmt.Printf("Error in deleteDiscord(): %v", err.Error())
	}

	/*Rs := db.getAllRSS()
	// Loop through every RSS file
	for _, r := range Rs {
		// Loop through every discord server array
		for i, j := range r.DiscordServers {
			// If you find the server ID we're deleting in said array then remove it
			if j == d.ServerID {
				r.DiscordServers = append(r.DiscordServers[:i], r.DiscordServers[i+1:]...)
				db.updateRSS(r)
			}
		}
	}*/
}

/*
updateDiscord updates
*/
func (db *DBInfo) updateDiscord(d Discord) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	err = session.DB(db.DBName).C(db.CollectionDiscord).Update(bson.M{"serverid": d.ServerID}, bson.M{"$set": bson.M{"channelid": d.ChannelID}})
	if err != nil {
		fmt.Printf("Error in updateDiscord(): %v", err.Error())
	}
}

/*
----------------------------------------------RSS----------------------------------------------
- addRSS(u string) RSS
- getRSS(u string) RSS
- deleteRSS(u string)
- updateRSS(r RSS)
- getAllRSS() []RSS
----------------------------------------------RSS----------------------------------------------
*/

/*
addRSS adds
*/
func (db *DBInfo) addRSS(u string) RSS {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var r RSS
	r.URL = u
	c := readRSS(u)
	r.LastUpdate = toTime(c.LastBuildDate)

	// Inserts the RSS into the database
	err = session.DB(db.DBName).C(db.CollectionRSS).Insert(r)
	if err != nil {
		fmt.Printf("Error in addRSS(): %v", err.Error())
	}
	return r
}

/*
getRSS gets
*/
func (db *DBInfo) getRSS(u string) RSS {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	r := RSS{}

	err = session.DB(db.DBName).C(db.CollectionRSS).Find(bson.M{"url": u}).One(&r)
	if err != nil {
		fmt.Printf("Error in getRSS(): %v", err.Error())
	}
	return r
}

/*
deleteRSS deletes
*/
func (db *DBInfo) deleteRSS(u string) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DBName).C(db.CollectionRSS).Remove(bson.M{"url": u})
	if err != nil {
		fmt.Printf("Error in deleteRSS(): %v", err.Error())
	}
}

/*
updateRSS updates
*/
func (db *DBInfo) updateRSS(r RSS) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Updates the time of the last update
	err = session.DB(db.DBName).C(db.CollectionRSS).Update(bson.M{"url": r.URL}, bson.M{"$set": bson.M{"lastUpdate": r.LastUpdate}})
	if err != nil {
		fmt.Printf("Error in updateRSS(): %v", err.Error())
	}

	// Updates the discord server array
	err = session.DB(db.DBName).C(db.CollectionRSS).Update(bson.M{"url": r.URL}, bson.M{"$set": bson.M{"discordServers": r.DiscordServers}})
	if err != nil {
		fmt.Printf("Error in updateRSS(): %v", err.Error())
	}
}

/*
getAllRSS gets an array
*/
func (db *DBInfo) getAllRSS() []RSS {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var r []RSS

	err = session.DB(db.DBName).C(db.CollectionRSS).Find(bson.M{}).All(&r)
	if err != nil {
		fmt.Printf("Error in getAllRSS(): %v", err.Error())
	}
	return r
}

/*
Used for "manageSubscription" to decide if we want to add or remove
an rss file from a discord server
*/
const (
	add    = iota
	remove = iota
)

/*
!!!NOT TESTED!!!
manageSubscription adds/removes a particular RSS feed from a server
!!!NOT TESTED!!!
*/
func (db *DBInfo) manageSubscription(rssURL string, serverID string, option int) (success bool) {
	// Check if discord server even exists
	_, err := db.getDiscord(serverID)
	if err != nil {
		fmt.Printf("Error while trying to get discord server with ID: " + serverID + ", " + err.Error())
		return false
	}
	rss := db.getRSS(rssURL)

	// Check if discord server is subscribed to RSS feed or not
	var index int
	for index = range rss.DiscordServers {
		if rss.DiscordServers[index] == serverID {
			break
		}
		index = -1
	}

	// New subscription
	if option == add {
		// If discord server is already subscribed
		if index != -1 {
			fmt.Printf("Discord server " + serverID + " is already subscribed to RSS feed " + rssURL)
			return false
		}
		// Add the new RSS feed to collection
		if len(rss.DiscordServers) == 0 {
			rss = db.addRSS(rssURL)
		}

		// Subscribe server to RSS feed
		rss.DiscordServers = append(rss.DiscordServers, serverID)
		db.updateRSS(rss)

		return true
		// Remove subscription
	} else if option == remove {
		// Can't remove something that doesn't exist
		if index == -1 {
			fmt.Printf("Discord server " + serverID + " is not subscribed to RSS feed " + rssURL + ". So there's nothing to remove.")
			return false
		}

		// Remove subscription of server to RSS feed
		rss.DiscordServers = append(rss.DiscordServers[:index], rss.DiscordServers[index+1:]...)
		return true
	} else {
		panic("Wrong use of 'option' parameter in manageSubscription(), value must be either 0 or 1. Value received: " + strconv.Itoa(option))
	}
}
