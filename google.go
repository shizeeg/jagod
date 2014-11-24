package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

var gsearch = "http://ajax.googleapis.com/ajax/services/search/web?v=1.0&oe=utf8&ie=utf8&q=%s&start=%d"

type result struct {
	Title   string `json:"titleNoFormatting"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type page struct {
	Label int    `json:"label"`
	Start string `json:"start"`
}

// GoogleSearch represents json data from Google
type GoogleSearch struct {
	Status       int `json:"responseStatus"`
	ResponseData struct {
		Results []result `json:"results"`
		Cursor  struct {
			Pages []page `json:"pages"`
		}
	}
}

// Google returs parsed Google result in a GoogleSearch struct form
func Google(query string, start uint) (result GoogleSearch, err error) {
	res, err := http.Get(fmt.Sprintf(gsearch, url.QueryEscape(query), start))
	if err != nil {
		return
	}
	defer res.Body.Close()

	jsonData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	var jsbytes = []byte(jsonData)
	err = json.Unmarshal(jsbytes, &result)
	return
}

/*
func main() {
	var response GoogleSearch
	response, err := Google("привет", 0)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("STRUCT: %#v\n", response)
	for _, v := range response.ResponseData.Results {
		fmt.Printf("%s\n%s\n%s\n\n", v.Title, v.URL, v.Content)
	}
	v := response.ResponseData.Cursor.Pages[1]
	fmt.Printf("Label: %d\nStart: %s\n", v.Label, v.Start)

	fmt.Printf("%v", response.ResponseData.Cursor.Pages)
}
*/

/*
var jsn = `{
    "responseData": {
        "cursor": {
            "currentPageIndex": 0,
            "estimatedResultCount": "238000",
            "moreResultsUrl": "http://www.google.com/search?oe=utf8&ie=utf8&source=uds&start=0&hl=en&q=golang",
            "pages": [
                {
                    "label": 1,
                    "start": "0"
                },
                {
                    "label": 2,
                    "start": "4"
                },
                {
                    "label": 3,
                    "start": "8"
                },
                {
                    "label": 4,
                    "start": "12"
                },
                {
                    "label": 5,
                    "start": "16"
                },
                {
                    "label": 6,
                    "start": "20"
                },
                {
                    "label": 7,
                    "start": "24"
                },
                {
                    "label": 8,
                    "start": "28"
                }
            ],
            "resultCount": "238,000",
            "searchResultTime": "0.16"
        },
        "results": [
            {
                "GsearchResultClass": "GwebSearch",
                "cacheUrl": "http://www.google.com/search?q=cache:rie1WixWbVcJ:golang.org",
                "content": "Documentation, source, and other resources for Google&#39;s Go language.",
                "title": "The Go Programming Language",
                "titleNoFormatting": "The Go Programming Language",
                "unescapedUrl": "https://golang.org/",
                "url": "https://golang.org/",
                "visibleUrl": "golang.org"
            },
            {
                "GsearchResultClass": "GwebSearch",
                "cacheUrl": "http://www.google.com/search?q=cache:cVd07aBzlbkJ:golang.org",
                "content": "File name, Kind, OS, Arch, SHA1 Checksum. go1.3.3.src.tar.gz, Source, \nb54b7deb7b7afe9f5d9a3f5dd830c7dede35393a. go1.3.3.darwin-386-osx10.6.\ntar.gz\u00a0...",
                "title": "Downloads - The Go Programming Language",
                "titleNoFormatting": "Downloads - The Go Programming Language",
                "unescapedUrl": "http://golang.org/dl/",
                "url": "http://golang.org/dl/",
                "visibleUrl": "golang.org"
            },
            {
                "GsearchResultClass": "GwebSearch",
                "cacheUrl": "http://www.google.com/search?q=cache:OZnThM-T9KkJ:tour.golang.org",
                "content": "Welcome to a tour of the Go programming language. The tour is divided into \nthree sections: basic concepts, methods and interfaces, and concurrency.",
                "title": "A Tour of Go",
                "titleNoFormatting": "A Tour of Go",
                "unescapedUrl": "http://tour.golang.org/",
                "url": "http://tour.golang.org/",
                "visibleUrl": "tour.golang.org"
            },
            {
                "GsearchResultClass": "GwebSearch",
                "cacheUrl": "http://www.google.com/search?q=cache:WuYKHPcyrwcJ:www.reddit.com",
                "content": "<b>golang</b>. subscribeunsubscribe11,180 readers. ~23 users here now. Go Home ... \nterminal.com Vanilla Go Snapshot (self.<b>golang</b>). submitted 10 hours ago * by\u00a0...",
                "title": "r/<b>Golang</b> - Reddit",
                "titleNoFormatting": "r/Golang - Reddit",
                "unescapedUrl": "https://www.reddit.com/r/golang",
                "url": "https://www.reddit.com/r/golang",
                "visibleUrl": "www.reddit.com"
            }
        ]
    },
    "responseDetails": null,
    "responseStatus": 200
}`
*/
