package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/shizeeg/ini"
	"github.com/shizeeg/youtube"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

var (
	duration    string
	encoding    string
	contType    string
	lang        string
	link        string
	exclude     string
	youtubeKey  string
	showHeaders bool
	prefix      = "Заголовок"
	unkPrefix   = "HTTP заголовок"
)

func init() {
	cfg, err := ini.Load("/etc/jagod.cfg")
	if err != nil {
		log.Fatalf("can't find config file %v\n", err)
	}
	flag.StringVar(&youtubeKey, "key", cfg.Section("google").Key("youtubeDataAPIKey").String(), "YouTube Data API key")
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

	if msg, err := url.QueryUnescape(link); err == nil {
		if strings.Count(link, "%") > 3 {
			fmt.Print(msg)
		}
	}

	res, err := http.Get(link)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

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
	if title == "" {
		return
	}
	if youtubeKey != "" && strings.HasSuffix(title, "- YouTube") {
		if ids := youtube.IDs(link); len(ids) > 0 {
			if dur, err := youtube.GetDuration(youtubeKey, ids[0]); err == nil && dur != "" {
				title = fmt.Sprintf("%s [%s]", title, dur)
			}
		}
	}
	fmt.Println(prefix + ": " + title)
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
			return "", errors.New("token error")
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
