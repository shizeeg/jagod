package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

var (
	encoding    string
	contType    string
	lang        string
	prefix      = "Заголовок"
	link        string
	showHeaders bool
)

func init() {
	flag.StringVar(&encoding, "oe", "utf8", "Output Encoding")
	flag.StringVar(&lang, "lang", "ru", "Language to output in")
	flag.BoolVar(&showHeaders, "headers", false, "Show HTTP Headers")
	flag.Parse()
}

func main() {
	if lang != "ru" {
		prefix = "Title"
	}
	if len(flag.Args()) > 0 && len(flag.Arg(0)) > 0 {
		link = flag.Arg(0)
	}
	if len(link) <= 4 || link[0:4] != "http" {
		usage(lang)
		return
	}
	if headers, err := http.Head(link); err == nil {
		contType = headers.Header.Get("Content-Type")
		if showHeaders || len(contType) >= 9 && contType[0:9] != "text/html" {
			for k, v := range headers.Header {
				if showHeaders || (k == "Content-Type" || k == "Content-Length") {
					fmt.Printf("\n%s: %s", k, v)
				}
			}
			return
		}
	}
	res, err := http.Get(link)
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
		fmt.Println("gettitle [-lang=ru|en] [-oe=utf8] [-headers] <http(s)://www.server.com/>")
	} else {
		fmt.Println("gettitle [-lang=ru|en] [-oe=utf8] [-headers] <http(s)://сервер.рф/>")
	}
}
