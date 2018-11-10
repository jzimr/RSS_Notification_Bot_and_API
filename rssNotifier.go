package main

import (
	"log"
	"time"
)

func schedule(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}

/*
ScanAndPost gets all rss documents and range over them
Post to discord if there are updates
*/
func scanAndPost() {
	log.Println("Firing at all cylinders")
	rss, err := db.getAllRSS()
	if err != nil {
		log.Println("Error in scanAndPost()", err)
	}

	for i := range rss {
		postRSS(rss[i].URL)
	}
}
