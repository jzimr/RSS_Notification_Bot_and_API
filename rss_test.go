package main

import (
	"testing"
)

/*
Test_readRSS checks that you can parse data from an RSS link
*/
func Test_readRSS(t *testing.T) {
	DBInitTest()
	url := "https://www.nrk.no/rogaland/toppsaker.rss"
	c := readRSS(url)
	pub, _ := toTime(c.Items[0].PubDateString)
	// If expected data is wrong or missing
	if !(pub == c.LastUpdate) || c.LastUpdate == 0 || len(c.Items) <= 0 || url != c.OriginalRSSLink {
		t.Errorf("Expected %v, got %v\nExpected more than 0, got %v\nExpected %v, got %v\n", pub, c.LastUpdate, len(c.Items), url, c.OriginalRSSLink)

	}
}

/*
Test_toTime checks that you can read a timestamp from a text string with extra characters
*/
func Test_toTime(t *testing.T) {
	s := "\"Thu, 15 Nov 2018 16:36:40 GMT \n\""
	i, err := toTime(s)
	if err != nil {
		t.Errorf("Problem with parsing toTime(): %v", err)
	}
	if i != 1542299800 {
		t.Errorf("You were supposed to get %v but got %v", 1542299800, i)
	}
}
