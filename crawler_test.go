package main

import (
	"strconv"
	"testing"
)

func TestIsPageRSS(t *testing.T) {
	// Test for an HTML page
	htmlURL := "https://www.nrk.no/"

	ok := isPageRSS(htmlURL)
	if ok {
		t.Errorf("Page " + htmlURL + " should have returned false on isPageRSS(), but returned true")
	}

	// Test for an XML page
	xmlURL := "https://www.nrk.no/toppsaker.rss"

	ok = isPageRSS(xmlURL)
	if !ok {
		t.Errorf("Page " + xmlURL + " should have returned true on isPageRSS(), but returned false")
	}
}

// func TestFetchRSSLinks

func TestGoogleSearch(t *testing.T) {
	// No Google results
	keyword := "asdlk234%&dfgdfgkljhdfgkljh123321dfg"
	rssLinks := googleSearchRssLinks(keyword)

	if len(rssLinks) != 0 {
		t.Errorf("Expected 0 results, got " + strconv.Itoa(len(rssLinks)))
	}

	// First Google result is .xml or .rss file type
	keyword = "nrk+filetype:rss"
	rssLinks = googleSearchRssLinks(keyword)

	if len(rssLinks) != 1 {
		t.Errorf("Expected one 1 result, got " + strconv.Itoa(len(rssLinks)))
	}

	// Google returns the desired results
	keyword = "nrk"
	rssLinks = googleSearchRssLinks(keyword)

	if len(rssLinks) <= 5 {
		t.Errorf("Expected multiple feeds, got " + strconv.Itoa(len(rssLinks)))
	}
}

func TestCrawler(t *testing.T) {
	var testCases = []string{
		"https://www.nrk.no",
		"www.nrk.no",
		"nrk.no",
		"nrk",
		"n",
		"bbc.com",
		"http://rss.nytimes.com/services/xml/rss/nyt/HomePage.xml",
	}

	// Successful crawl
	for _, URL := range testCases {
		// Crawler should always find rss links
		links := Crawl(URL)

		if len(links) == 0 {
			t.Errorf("Crawler did not return any links when it actually should.")
		}
	}
}
