package main

import (
	"fmt"

	"github.com/globalsign/mgo"
)

// DBInfo stores the details of the DB connection
type DBInfo struct {
	DBURL          string
	DBName         string
	CollectionUser string
	CollectionRSS  string
}

// db stores the credentials of our database
var db DBInfo

// DBInit fills DBInfo with the information about our database
func DBInit() {
	db.DBURL = "mongodb://rssbot:rssbot1@ds253243.mlab.com:53243/rssbot"
	db.DBName = "rssbot"
	db.CollectionUser = "user"
	db.CollectionRSS = "RSS"
}

// addUser adds new users to the database
func (db *DBInfo) addUser(u User) User {
	// Creates a connection
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Inserts the user into the database
	err = session.DB(db.DBName).C(db.CollectionUser).Insert(u)
	if err != nil {
		fmt.Printf("Error in addUser(): %v", err.Error())
	}

	return u
}

/*
-----------------------------------------------------------------------------------------------
----------------------------------------------RSS----------------------------------------------
-----------------------------------------------------------------------------------------------
*/

// addRSS adds new RSS to the database
func (db *DBInfo) addRSS(r RSS) RSS {
	// Creates a connection
	session, err := mgo.Dial(db.DBURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Inserts the user into the database
	err = session.DB(db.DBName).C(db.CollectionRSS).Insert(r)
	if err != nil {
		fmt.Printf("Error in addRSS(): %v", err.Error())
	}

	return r
}
