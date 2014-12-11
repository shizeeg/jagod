package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

var (
	encoding    string
	contType    string
	lang        string
	link        string
	exclude     string
	showHeaders bool
	prefix      = "Заголовок"
	unkPrefix   = "HTTP заголовок"
)

func init() {
	flag.StringVar(&encoding, "oe", "utf8", "Output Encoding")
	flag.StringVar(&lang, "lang", "ru", "Language to output in")
	flag.BoolVar(&showHeaders, "headers", false, "Show HTTP Headers")
	flag.StringVar(&exclude, "exclude", "image/*", "Exclude these content types using regexp")
	flag.Parse()
}

func main() {
	if lang != "ru" {
		prefix = "Title"
		unkPrefix = "HTTP Header"
	}
	if len(flag.Args()) > 0 && len(flag.Arg(0)) > 0 {
		link = flag.Arg(0)
	}
	if len(link) <= 4 || link[0:4] != "http" {
		usage(lang)
		return
	}
	res, err := http.Get(link)

	contType = res.Header.Get("Content-Type")
	if showHeaders || len(contType) >= 9 && contType[0:9] != "text/html" {
		if ok, _ := regexp.MatchString(exclude, contType); ok && exclude != "" {
			return
		}
		fmt.Print(unkPrefix + ":")
		for k, v := range res.Header {
			if showHeaders || (k == "Content-Type" || k == "Content-Length") {
				fmt.Printf("\n%s: %s", k, v)
			}
		}
		return
	}

	if err != nil {
		log.Fatal(err)
	}
	var title string
	if text, err := charset.NewReader(res.Body, contType); err == nil {
		title, _ = getTag(text, "title")
	}
	title = strings.Trim(title, "\n ")
	if len(title) > 0 {
		fmt.Print(prefix + ": " + title)
	}
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// getTag gets data (textToken) from requested html-tag.
func getTag(r io.Reader, tag string) (data string, err error) {
	d := html.NewTokenizer(r)
	var currTag html.Token
	for {
		// token type
		tokenType := d.Next()
		if tokenType == html.ErrorToken {
			return "", errors.New("Token Error!")
		}
		token := d.Token()
		switch tokenType {
		case html.StartTagToken: // <tag>
			// type Token struct {
			//     Type     TokenType
			//     DataAtom atom.Atom
			//     Data     string
			//     Attr     []Attribute
			// }
			//
			// type Attribute struct {
			//     Namespace, Key, Val string
			// }

			if token.Data == tag {
				currTag = token
			}
		case html.TextToken: // text between start and end tag
			if currTag.Data == tag {
				return token.Data, nil
			}
		case html.EndTagToken: // </tag>
			if token.Data == tag {
				currTag = token
			}
		case html.SelfClosingTagToken: // <tag/>
		}
	}
}

func usage(lang string) {
	if lang != "ru" {
		fmt.Println("gettitle [-lang=ru|en] [-oe=utf8] [-headers] [-filter=\"\"] <http(s)://www.server.com/>")
		flag.PrintDefaults()
	} else {
		fmt.Println("gettitle [-lang=ru|en] [-oe=utf8] [-headers] [-filter\"\"] <http(s)://сервер.рф/>")
	}
}
