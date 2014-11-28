package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var (
	gsearch   = "http://ajax.googleapis.com/ajax/services/search/web?v=1.0&q=%s&oe=%s&ie=%s&start=%d"
	query     string
	lang      string
	ie, oe    string
	startPage uint
)

func init() {
	flag.StringVar(&lang, "lang", "en", "Output language")
	flag.StringVar(&oe, "oe", "utf8", "Output encoding")
	flag.StringVar(&ie, "ie", "utf8", "Input encoding")
	flag.UintVar(&startPage, "start", 0, "Start page")
	flag.Parse()
}

func main() {
	if len(flag.Args()) > 0 && len(flag.Arg(0)) > 0 {
		query = strings.Join(flag.Args(), " ")
	} else {
		usage(lang)
		return
	}

	var response GoogleSearch
	response, err := Google(query, startPage)
	if err != nil {
		fmt.Println(err)
	}
	results := response.ResponseData.Results
	if len(results) <= 0 {
		if lang != "ru" {
			fmt.Println("Empty Response!")
		} else {
			fmt.Println("Ничего не найдено!")
		}
	}
	var res uint
	if len(results) <= 3 {
		res = startPage % 4
	}
	fmt.Printf("%s\n%s\n%s\n", results[res].Title, results[res].URL, results[res].Content)
}

// Google returs parsed Google result in a GoogleSearch struct form
func Google(query string, start uint) (result GoogleSearch, err error) {
	out := fmt.Sprintf(gsearch, url.QueryEscape(query), oe, ie, startPage)
	res, err := http.Get(out)
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

// usage prints usage screen
func usage(lang string) {
	if lang == "ru" {
		fmt.Println("google [-lang=en] [-start=0] <строка поиска>")
		fmt.Println("  -lang=en|ru - выводит сообщения на выбранном языке")
		fmt.Println("  -start=№ - показать результат № (начиная с нуля)")
	} else {
		fmt.Println("google [-lang=en] [-start=0] <query>")
		fmt.Println("  -lang=en|ru - output messages in this language")
		fmt.Println("  -start=# - show result # (starts from 0)")
	}
}

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
