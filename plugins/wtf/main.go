package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

// i18n is Internatialization string type
type i18n map[int]string

// i18n constants (IDs)
const (
	// IDK I don't know
	IDK = iota
)

var (
	lang string
	path string
)

func init() {
	flag.StringVar(&lang, "lang", "en", "language to output messages in")
	// FIX: Where should we search for those files on Windows and OSX?
	flag.StringVar(&path, "f", filepath.Join("/usr/share/misc", "acronyms*"), "path to acronyms database")
}

func main() {
	flag.Parse()
	if flag.Arg(0) == "" {
		flag.PrintDefaults()
		return
	}
	files, err := filepath.Glob(path)
	userfiles, _ := filepath.Glob("acronyms*")
	files = append(files, userfiles...)

	if err != nil {
		log.Fatalf("no such path or directory: %q: %v\n", path, err)
	}
	// let's mimick standard `wtf` here skip is if used like `wtf is smth`
	acronym := flag.Arg(0)
	if flag.Arg(0) == "is" && flag.NArg() > 1 {
		acronym = flag.Arg(1)
	}
	for _, v := range wtf(acronym, files) {
		fmt.Println(v)
	}
}

// wtf tries to find an acronym in each file in files array
func wtf(acro string, files []string) (acronyms []string) {
	acro = strings.ToUpper(acro)
	// skip the gap here
	re := regexp.MustCompile("^(?P<acronym>" + acro + ")([[:space:]]+)(?P<definition>.+)")
	for _, f := range files {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			continue
		}
		for _, v := range strings.Split(string(data), "\n") {
			if re.MatchString(v) {
				match := fmt.Sprintf("${%s}: ${%s}", re.SubexpNames()[1], re.SubexpNames()[3])
				acronyms = append(acronyms, re.ReplaceAllString(v, match))
			}
		}
	}
	if len(acronyms) == 0 {
		return append(acronyms, localize(IDK, lang, acro))
	}
	return
}

// localize returns a localization string
// fmt.Println(localize(IDK, "en", "wtf"))
// Output: wtf, I don't know what wtf means!
// TODO(shizeeg) split as a generic localization framework
func localize(message int, lang string, vars ...string) string {
	loc := map[string]i18n{
		"en": i18n{IDK: fmt.Sprintf("wtf, I don't know what %s means!", strings.Join(vars, ", "))},
		"ru": i18n{IDK: fmt.Sprintf("чозанах? Я не знаю, что такое %s!", strings.Join(vars, ", "))},
	}
	return loc[lang][message]
}
