package main

import (
	"errors"
	"fmt"
	"log"
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
- addDiscord(d Discord) (Discord, error)
- countDiscord() (int, error)
- getDiscord(s string) (Discord, error)
- deleteDiscord(d Discord) error
- updateDiscord(d Discord) error
- deleteAllDiscord()
- getDiscordFromAPI(s string) (Discord, error)
--------------------------------------------Discord--------------------------------------------
*/

/*
addDiscord adds a new discord to the database
*/
func (db *DBInfo) addDiscord(d Discord) (Discord, error) {
	// Creates a connection
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Inserts the user into the database
	err = session.DB(db.DBName).C(db.CollectionDiscord).Insert(d)

	return d, err
}

/*
countDiscord counts how many discord servers there are
*/
func (db *DBInfo) countDiscord() (int, error) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	count, err := session.DB(db.DBName).C(db.CollectionDiscord).Count()

	return count, err
}

/*
getDiscord gets the discord object from a serverId
*/
func (db *DBInfo) getDiscord(s string) (Discord, error) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	d := Discord{}
	err = session.DB(db.DBName).C(db.CollectionDiscord).Find(bson.M{"serverid": s}).One(&d)

	return d, err
}

/*
deleteDiscord is run if a server is deleted or bot kicked. It deletes the discord from the database
	and also removes the serverId from all the RSS feeds it was subscribed to
*/
func (db *DBInfo) deleteDiscord(d Discord) error {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DBName).C(db.CollectionDiscord).Remove(bson.M{"serverid": d.ServerID})
	if err != nil {
		return err
	}

	Rs, err := db.getAllRSS()
	if err != nil {
		return err
	}
	// Loop through every RSS file
	for _, r := range Rs {
		// Try to remove the discord serverId
		err = session.DB(db.DBName).C(db.CollectionRSS).Update(bson.M{"url": r.URL}, bson.M{"$pull": bson.M{"discordServers": d.ServerID}})
		// Check if there's still servers subscribed to this rss, delete it if not
		r, err = db.getRSS(r.URL)
		if err != nil {
			return err
		}
		if len(r.DiscordServers) == 0 {
			log.Println("Delete RSS from DB. Empty DiscordList")
			err = db.deleteRSS(r.URL)
			if err != nil {
				return err
			}
		}
	}

	return err
}

/*
updateDiscord is run when the user uses !configure and updates the channelId and APIKey of a discord server
*/
func (db *DBInfo) updateDiscord(d Discord) error {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	err = session.DB(db.DBName).C(db.CollectionDiscord).Update(bson.M{"serverid": d.ServerID}, bson.M{"$set": bson.M{"channelid": d.ChannelID, "apikey": d.APIKey}})

	return err
}

/*
deleteAllDiscord deletes every Discord object
*/
func (db *DBInfo) deleteAllDiscord() (rError error) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Delete everything from the collection
	_, err = session.DB(db.DBName).C(db.CollectionDiscord).RemoveAll(nil)
	return err
}

/*
getDiscordFromAPI gets the discord object from an apiKey
*/
func (db *DBInfo) getDiscordFromAPI(s string) (Discord, error) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	d := Discord{}
	err = session.DB(db.DBName).C(db.CollectionDiscord).Find(bson.M{"apikey": s}).One(&d)

	return d, err
}

/*
----------------------------------------------RSS----------------------------------------------
- addRSS(u string) (RSS, error)
- countRSS() (int, error)
- getRSS(u string) (RSS, error)
- deleteRSS(u string) error
- updateRSS(r RSS) error
- getAllRSS() ([]RSS, error)
- deleteAllRSS()
----------------------------------------------RSS----------------------------------------------
*/

/*
addRSS adds a new rss to the database
*/
func (db *DBInfo) addRSS(u string) (RSS, error) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Check if already exists
	dd, err := db.getRSS(u)
	if err == nil {
		return dd, errors.New("The RSS already exists")
	}

	var r RSS
	r.URL = u
	c := readRSS(u)
	r.LastUpdate = c.LastUpdate

	// Inserts the RSS into the database
	err = session.DB(db.DBName).C(db.CollectionRSS).Insert(r)

	return r, err
}

/*
countRSS counts how many rss elements there are
*/
func (db *DBInfo) countRSS() (int, error) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	count, err := session.DB(db.DBName).C(db.CollectionRSS).Count()

	return count, err
}

/*
getRSS gets a rss object from its url
*/
func (db *DBInfo) getRSS(u string) (RSS, error) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	r := RSS{}

	err = session.DB(db.DBName).C(db.CollectionRSS).Find(bson.M{"url": u}).One(&r)

	return r, err
}

/*
deleteRSS deletes a rss object from its url
*/
func (db *DBInfo) deleteRSS(u string) error {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DBName).C(db.CollectionRSS).Remove(bson.M{"url": u})

	return err
}

/*
updateRSS updates the time of the last update and the array of discord servers
*/
func (db *DBInfo) updateRSS(r RSS) error {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DBName).C(db.CollectionRSS).Update(bson.M{"url": r.URL}, bson.M{"$set": bson.M{"lastUpdate": r.LastUpdate, "discordServers": r.DiscordServers}})
	if err != nil {
		fmt.Printf("Error in updateRSS(): %v", err.Error())
	}

	//Check if discord array is empty. If yes delete this rss document
	dbR, err := db.getRSS(r.URL)
	if err != nil {
		fmt.Printf("Error in updateRSS(): %v", err.Error())
	}

	if len(dbR.DiscordServers) == 0 {
		err = db.deleteRSS(dbR.URL)
	}

	return err
}

/*
getAllRSS gets an array of all rss's
*/
func (db *DBInfo) getAllRSS() ([]RSS, error) {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var r []RSS

	err = session.DB(db.DBName).C(db.CollectionRSS).Find(bson.M{}).All(&r)

	return r, err
}

/*
deleteAllDiscord deletes every RSS object
*/
func (db *DBInfo) deleteAllRSS() {
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Delete everything from the collection
	session.DB(db.DBName).C(db.CollectionRSS).RemoveAll(nil)
}

/*
----------------------------------------------HELPERFUNCTIONS----------------------------------------------
- manageSubscription(string, string, int) bool
- getAllSubscribed(string) []RSS
----------------------------------------------HELPERFUNCTIONS----------------------------------------------
*/

/*
Used for "manageSubscription" to decide if we want to add or remove
an rss file from a discord server
*/
const (
	add    = iota
	remove = iota
)

/*
manageSubscription adds/removes a particular RSS feed from a server
*/
func (db *DBInfo) manageSubscription(rssURL string, serverID string, option int) (success bool) {
	// Check if discord server even exists
	_, err := db.getDiscord(serverID)
	if err != nil {
		log.Println("Error while trying to get discord server with ID: " + serverID + ", " + err.Error())
		return false
	}
	rss, err := db.getRSS(rssURL)
	if err != nil {
		fmt.Printf("manageSubscriptions(): %v", err.Error())
	}

	// Check if discord server is subscribed to RSS feed or not
	index := -1
	for i := range rss.DiscordServers {
		if rss.DiscordServers[i] == serverID {
			index = i
			break
		}
	}

	// New subscription
	if option == add {
		// If discord server is already subscribed
		if index != -1 {
			log.Println("Discord server " + serverID + " is already subscribed to RSS feed " + rssURL)
			return false
		}
		// Add the new RSS feed to collection
		if len(rss.DiscordServers) == 0 {
			rss, err = db.addRSS(rssURL)
			log.Println(err)
		}

		// Subscribe server to RSS feed
		rss.DiscordServers = append(rss.DiscordServers, serverID)
		db.updateRSS(rss)

		return true
		// Remove subscription
	} else if option == remove {
		// Can't remove something that doesn't exist
		if index == -1 {
			log.Println("Discord server " + serverID + " is not subscribed to RSS feed " + rssURL + ". So there's nothing to remove.")
			return false
		}

		// Remove subscription of server to RSS feed
		rss.DiscordServers = append(rss.DiscordServers[:index], rss.DiscordServers[index+1:]...)
		db.updateRSS(rss)

		return true
	} else {
		panic("Wrong use of 'option' parameter in manageSubscription(), value must be either 0 or 1. Value received: " + strconv.Itoa(option))
	}
}

/*
getAllSubscribed returns a list of all RSS files the server has subscribed to
*/
func getAllSubscribed(serverID string) []RSS {
	var subscribedTo []RSS

	allRSS, err := db.getAllRSS()
	if err != nil {
		log.Println("Error while trying to get all RSS from database" + err.Error())
		return subscribedTo
	}

	for i, rss := range allRSS {
		for _, sID := range rss.DiscordServers {
			if sID == serverID {
				subscribedTo = append(subscribedTo, allRSS[i])
			}
		}
	}

	return subscribedTo
}
