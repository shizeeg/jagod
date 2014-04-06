package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
)

type TrTag struct {
	Pos  string   `xml:"pos,attrib,omitempty"`
	Text string   `xml:"text"`
	Syn  []string `xml:"syn>text,omitempty"`
	Mean []string `xml:"mean>text,omitempty"`
	Ex   []string `xml:"ex>text,omitempty"`
	Extr []string `xml:"ex>tr>text,omitempty"`
}

type DefTag struct {
	Pos  string  `xml:"pos,attr,omitempty"`
	Ts   string  `xml:"ts,attr,omitempty"`
	Text string  `xml:"text"`
	Tr   []TrTag `xml:"tr"`
}

type DicResult struct {
	Def []DefTag `xml:"def"`
}

type LangResult struct {
	String []string `xml:"string"`
}

type SpellResult struct {
	// XMLName xml.Name `xml:"SpellResult"`
	Errors []SpellError `xml:"error,omitempty"`
}

type SpellError struct {
	Code  string   `xml:"code,attr"`
	Pos   string   `xml:"pos,attr"`
	Row   string   `xml:"col,attr"`
	Len   string   `xml:"len,attr"`
	Words []string `xml:"word"`
	S     []string `xml:"s,omitempty"`
}

// GetLangs returns possible language pairs for YandexDic()
func GetLangs(apikey string) []string {
	langsurl := fmt.Sprintf("https://dictionary.yandex.net/api/v1/dicservice/getLangs?key=%s", apikey)
	resp, err := http.Get(langsurl)
	if err != nil {
		return []string{err.Error()}
	}
	defer resp.Body.Close()
	dec := xml.NewDecoder(resp.Body)
	var r LangResult
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
	var r SpellResult
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
	var d DicResult
	if err := dec.Decode(&d); err != nil {
		return err.Error()
	}
	return d.String()
}

// String is a pretty printer for DicResult
func (r *DicResult) String() string {
	tr := func(d *DefTag) string {
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
