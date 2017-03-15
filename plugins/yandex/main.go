package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/shizeeg/ini"
)

var (
	lang   string
	apikey string
	banner = "http://api.yandex.ru/dictionary/"
)

func init() {
	cfg, err := ini.Load("/etc/yandex.cfg")
	if err != nil {
		log.Println(err)
	}

	flag.StringVar(&lang, "lang", cfg.Section("").Key("LANG").In("ru", []string{}), "Localization")
	flag.StringVar(&apikey, "apikey", cfg.Section("").Key("APIKEY").String(), "Yandex Dict.API Key")
	flag.Parse()
}

func main() {
	fmt.Println(banner)
	if flag.NArg() < 1 || len(flag.Arg(0)) < 2 {
		usage(lang)
		return
	}

	switch {
	case flag.Arg(0) == "list":
		fmt.Println(strings.Join(GetLangs(apikey), ", "))
		return
	case flag.Arg(0) == "spell":
		fmt.Println(YandexSpell("", strings.Join(flag.Args()[1:], " ")))
		return
	case flag.Arg(0) == "help":
		usage(lang)
		return
	}

	fmt.Println(YandexDic(apikey, lang, flag.Arg(0), flag.Arg(1)))
}

func usage(lang string) {
	if lang != "ru" {
		fmt.Fprintln(os.Stderr, "  tr <from-to> <word>")
		fmt.Fprintln(os.Stderr, "  spell <text>")
	} else {
		fmt.Fprintln(os.Stderr, "  tr <из-в> <слово>")
		fmt.Fprintln(os.Stderr, "  spell <текст>")
	}
}

// GetLangs returns possible language pairs for YandexDic()
func GetLangs(apikey string) []string {
	if apikey == "" {
		log.Fatalln("Can't get \"APIKEY\" environment variable!")
	}
	langsurl := fmt.Sprintf("https://dictionary.yandex.net/api/v1/dicservice/getLangs?key=%s", apikey)
	resp, err := http.Get(langsurl)
	if err != nil {
		return []string{err.Error()}
	}
	defer resp.Body.Close()
	dec := xml.NewDecoder(resp.Body)
	var r langresult
	if err := dec.Decode(&r); err != nil {
		return []string{err.Error()}
	}
	return r.String
}

// YandexSpell returns string with correct spellings if any.
func YandexSpell(langs, text string) string {
	requrl := "https://speller.yandex.net/services/spellservice/checkText"
	values := url.Values{"ie": {"utf-8"}, "lang": {langs}, "text": {text}}
	resp, err := http.PostForm(requrl, values)
	defer resp.Body.Close()

	if err != nil {
		return err.Error()
	}

	dec := xml.NewDecoder(resp.Body)
	var r spellresult
	if err := dec.Decode(&r); err != nil {
		return err.Error()
	}
	var s string
	for _, e := range r.Errors {
		for iw, _ := range e.Words {
			if iw >= len(e.S) {
				break
			}
			s = s + " *" + e.S[iw]
			// str = strings.Replace(text, w, e.S[iw], 1)
		}
	}
	// return text + "\n" + s
	if len(s) == 0 {
		return "empty response from Yandex"
	}
	return s
}

// YandexDic translates message to given lang-lang pair.
// call GetLangs() to get possible lang pairs
func YandexDic(apikey, resplang, lang, text string) string {
	if apikey == "" {
		log.Fatalln("No Yandex dict \"APIKEY\" defined!")
	}

	text = url.QueryEscape(text)
	if len(resplang) < 2 {
		resplang = "en"
	}
	requrl := fmt.Sprintf("https://dictionary.yandex.net/api/v1/dicservice/lookup?key=%s&lang=%s&ui=%s&text=%s", apikey, lang, resplang, text)
	resp, err := http.Get(requrl)
	if err != nil {
		return err.Error()
	}

	defer resp.Body.Close()
	dec := xml.NewDecoder(resp.Body)
	var d dicresult
	if err := dec.Decode(&d); err != nil {
		return err.Error()
	}
	return d.String()
}

// String is a pretty printer for DicResult
func (r *dicresult) String() string {
	tr := func(d *deftag) string {
		var s string
		for _, tr := range d.Tr {
			s = s + fmt.Sprintf("%s, ", tr.Text)
		}
		return s[:len(s)-2] + "\n" // cut last comma.
	}
	var str string
	for i, def := range r.Def {
		if len(def.Ts) > 0 {
			str = str + fmt.Sprintf("%d. %s [%s] (%s)\n   %s",
				i+1, def.Text, def.Ts, def.Pos, tr(&def))
		} else {
			str = str + fmt.Sprintf("%d. %s (%s)\n   %s",
				i+1, def.Text, def.Pos, tr(&def))
		}
	}
	return str
}

type trtag struct {
	Pos  string   `xml:"pos,attrib,omitempty"`
	Text string   `xml:"text"`
	Syn  []string `xml:"syn>text,omitempty"`
	Mean []string `xml:"mean>text,omitempty"`
	Ex   []string `xml:"ex>text,omitempty"`
	Extr []string `xml:"ex>tr>text,omitempty"`
}

type deftag struct {
	Pos  string  `xml:"pos,attr,omitempty"`
	Ts   string  `xml:"ts,attr,omitempty"`
	Text string  `xml:"text"`
	Tr   []trtag `xml:"tr"`
}

type dicresult struct {
	Def []deftag `xml:"def"`
}

type langresult struct {
	String []string `xml:"string"`
}

type spellresult struct {
	// XMLName xml.Name `xml:"SpellResult"`
	Errors []spellerror `xml:"error,omitempty"`
}

type spellerror struct {
	Code  string   `xml:"code,attr"`
	Pos   string   `xml:"pos,attr"`
	Row   string   `xml:"col,attr"`
	Len   string   `xml:"len,attr"`
	Words []string `xml:"word"`
	S     []string `xml:"s,omitempty"`
}
