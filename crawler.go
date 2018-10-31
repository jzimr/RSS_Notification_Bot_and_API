package main

/*
crawler.go takes a webpage URL as input, searches through the webpage's source code
after links that contain "rss", and returns them
*/

import (
	"fmt"
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
					if strings.Contains(a.Val, "rss") {
						fmt.Println(a.Val)
						rssLinks = append(rssLinks, a.Val)
					}
				}
			}
		}
	}
	return rssLinks
}

/*
Crawl returns a link to the RSS of the URL provided, or err if none found
*/
func Crawl(URL string) (RSSLink string, err error) {
	if !strings.Contains(URL, "http://") || !strings.Contains(URL, "https://") {
		URL = "http://" + URL
	}
	links := fetchRSSLinks(URL)

	if len(links) == 0 {
		return "", fmt.Errorf("No RSS link found on the given webpage")
	}
	// If there are more links, we just return the first one
	return links[0], nil
}
