package main

/*
crawler.go takes a webpage URL as input, searches through the webpage's source code
after links that contain "rss", and returns them
*/

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

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

	if strings.HasPrefix(string(body), "<?xml version=") {
		return true
	}
	return false
}

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
				if a.Key == "href" && strings.Contains(a.Val, "rss") {
					//Check if link on the page is valid
					_, err := url.ParseRequestURI(a.Val)
					if err != nil {
						continue
					}

					if isPageRSS(a.Val) {
						rssLinks = append(rssLinks, a.Val)
					}
				}
			}
		}
	}
	return rssLinks
}

/*
---(WIP) Make a google search based on the given URL to find RSS links (WIP)---
*/
func googleSearchRssLinks(keyword string) (rssLinks []string) {
	URL := "https://www.google.com/search?q=" + keyword + "+rss"
	var links []string

	// Make GET request
	baseClient := &http.Client{}
	req, _ := http.NewRequest("GET", URL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	resp, err := baseClient.Do(req)
	if err != nil {
		log.Println("An error occured while trying to make GET requestsssss, " + err.Error())
		return
	}
	defer resp.Body.Close()

	// Google search results
	result, err := googleResultParser(resp)
	if err != nil {
		log.Println("An error occured while trying to parse google result page, " + err.Error())
		return
	}

	// Check for results
	if len(result) == 0 {
		log.Println("Google could not find any search results with the given keyword: " + keyword)
		return
	}

	// Go through the first three results and check for RSS
	for i := 0; i < 3; i++ {
		fmt.Println(result[i].ResultURL)
		if strings.Contains(result[i].ResultURL, keyword) {
			// If page we found on google already is .rss page
			if isPageRSS(result[i].ResultURL) {
				fmt.Println("yop")
				links = append(links, result[i].ResultURL)
				break
			}

			tempLinks := fetchRSSLinks(result[i].ResultURL)
			if len(tempLinks) != 0 {
				links = tempLinks
				break
			}
		}
	}

	return links
}

/*
	Todo:
	1. Add support for multiple RSS links, and let the user choose which ones to include
	2. Webcrawl through google search on the main website and check for RSS links (Smart crawler) (DONE)
	3. Reddit Automatic RSS?
	3.5 Make bot post image on how-to-guide on how to get rss feeds from reddit if RSS search failed
	4. Create google searches based on the users geographical location (E.g. google.co.uk, google.de, google.com, google.no, ...)

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
