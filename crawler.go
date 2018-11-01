package main

/*
crawler.go takes a webpage URL as input, searches through the webpage's source code
after links that contain "rss", and returns them
*/

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

/*
Searches through a webpage's source and returns all links that have "rss" in them
*/
func fetchRSSLinks(URL string) (rssLinks []string) {
	resp, err := http.Get(URL)
	if err != nil {
		log.Println("An error occured while trying to make GET request, " + err.Error())
		return nil
	}
	defer resp.Body.Close()

	fmt.Println(URL)

	z := html.NewTokenizer(resp.Body)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			for _, a := range t.Attr {
				if a.Key == "href" {
					if strings.Contains(a.Val, "rss") && isPageRSS(a.Val) {
						rssLinks = append(rssLinks, a.Val)
					}
				}
			}
		}
	}
	return rssLinks
}

/*
Checks whether a given webpage (URL) is of RSS format
*/
func isPageRSS(URL string) (isRSS bool) {
	resp, err := http.Get(URL)
	if err != nil {
		log.Println("An error occured while trying to make GET request, " + err.Error())
		return false
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	decoder := xml.NewDecoder(strings.NewReader(string(body)))

	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return false
			}
			return false
		}

		// Search for "rss" node
		if v, ok := t.(xml.StartElement); ok {
			if v.Name.Local == "rss" {
				return true
			}
		}
	}
}

/*
---(WIP) Make a google search based on the given URL to find RSS links (WIP)---
*/
func googleSearchRssLinks(keyword string) (rssLinks []string) {
	URL := "https://www.google.no/search?q=" + keyword + "+rss"
	var result []string

	// Google search results
	result = fetchRSSLinks(URL)

	if len(result) == 0 {
		log.Println("Google could not find any search results with the given keyword: " + keyword)
		return
	}

	for i := 0; i < 3; i++ {
		if strings.Contains(result[i], keyword) && strings.Contains(result[i], "rss") {
			tempLinks := fetchRSSLinks(result[i])
			if len(tempLinks) != 0 {
				result = tempLinks
				break
			}
		}
	}
	return result
}

/*
	Todo:
	1. Add support for multiple RSS links, and let the user choose which ones to include
	2. Webcrawl through google search on the main website and check for RSS links (Smart crawler)
	3. Reddit Automatic RSS?
	3.5 Make bot post image on how-to-guide on how to get rss feeds from reddit if RSS search failed

	Improving crawler (How it should check for RSS given a link):
	1. Search through the webpage's source code afteran "rss" link (DONE)
	2.1. Slice the URL to get the main name (between www. and .com)
	2.2. Create a google search URL: "Google/search/" + slicedURL + "%20rss"
	2.3. Get the first x results and add these pages to the webcrawler queue
	2.4. Repeat Step 1 for each page from queue
	3. If no pages found, return "Not Found" to user
*/

/*
Crawl returns a link to the RSS of the URL provided, or err if none found
*/
func Crawl(URL string) (RSSLink string, err error) {
	if !strings.Contains(URL, "http://") && !strings.Contains(URL, "https://") {
		URL = "http://" + URL
	}

	// Check if user already sends us an RSS link
	if isPageRSS(URL) {
		log.Println("\"" + URL + "\" is already a .rss file")
		return URL, nil
	}

	// Search through the webpage's source code after an "rss" link
	links := fetchRSSLinks(URL)

	// If none found, Try the google searching method
	if len(links) == 0 {
		parts := strings.Split(URL, ".")
		var keyword string

		if parts[0] == "http://www" || parts[0] == "https://www" {
			keyword = parts[1]
		} else if strings.Contains(parts[0], "https://") {
			keyword = strings.TrimPrefix(parts[0], "https://")
		} else if strings.Contains(parts[0], "http://") {
			keyword = strings.TrimPrefix(parts[0], "http://")
		}

		// Execute searching method
		links = googleSearchRssLinks(keyword)
	}

	// Still no results? Might as well give up.
	if len(links) == 0 {
		return "", fmt.Errorf("No RSS link found on the given webpage")
	}

	// If there are more links, we just return the first one (Room for improvement here!)
	return links[0], nil
}
