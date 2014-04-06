package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

//func ExtractURL(msg string) string {
//	re := regexp.MustCompile(`(https?)://[^ ]+`)
//}

// Turn converting text typed in wrong layout
// from: ghbdtn/ (qwerty-layout)
// to: привет. ("hello" in Russian) (йцукен-layout)
func Turn(msg string) string {
	lat2cyr := map[rune]rune{
		'q': 'й', 'w': 'ц', 'e': 'у', 'r': 'к', 't': 'е', 'y': 'н',
		'u': 'г', 'i': 'ш', 'o': 'щ', 'p': 'з', '[': 'х', ']': 'ъ',
		'a': 'ф', 's': 'ы', 'd': 'в', 'f': 'а', 'g': 'п', 'h': 'р',
		'j': 'о', 'k': 'л', 'l': 'д', ';': 'ж', '\'': 'э', 'z': 'я',
		'x': 'ч', 'c': 'с', 'v': 'м', 'b': 'и', 'n': 'т', 'm': 'ь',
		',': 'б', '.': 'ю', '/': '.', '`': 'ё', '&': '?',

		'Q': 'Й', 'W': 'Ц', 'E': 'У', 'R': 'К', 'T': 'Е', 'Y': 'Н',
		'U': 'Г', 'I': 'Ш', 'O': 'Щ', 'P': 'З', '{': 'Х', '}': 'Ъ',
		'A': 'Ф', 'S': 'Ы', 'D': 'В', 'F': 'А', 'G': 'П', 'H': 'Р',
		'J': 'О', 'K': 'Л', 'L': 'Д', ':': 'Ж', '"': 'Э', 'Z': 'Я',
		'X': 'Ч', 'C': 'С', 'V': 'М', 'B': 'И', 'N': 'Т', 'M': 'Ь',
		'<': 'Б', '>': 'Ю', '?': ',', '~': 'Ё',
	}
	Lat2Cyr := func(c rune) rune {
		if ch, ok := lat2cyr[c]; ok {
			return ch
		}
		return c
	}
	return strings.Map(Lat2Cyr, msg)
}

// IsWrongLayout try to detect if text is in wrong layout.
// For ex: "ghbdtn lheu!"
func IsWrongLayout(msg string) bool {
	msg = strings.TrimSpace(msg)
	msglen := strings.Count(msg, "") - 1
	if msglen < 4 || msg[0] == '!' {
		return false
	}

	re := regexp.MustCompile(`[-a-zA-Z/'".,;:!?\]\[\<>{}~&\x60) ]`)
	latinSymbols := len(re.FindAllString(msg, -1))
	if latinSymbols < msglen { // cyrillic message, nothing to convert
		return false
	}
	ProcessedMsg := Turn(msg)
	if ProcessedMsg == msg {
		return false
	}
	// check if text is proper English
	enOk, enBad, err := SpellCheck(msg, "en")
	if err != nil { // SpellCHeck error. FIXME: What we need to return?
		return false
	}
	if enOk > 0 && enOk > enBad {
		return false
	}

	// check if text is Russian in wrong layout
	ruOk, _, err := SpellCheck(Turn(msg), "ru")
	if err != nil { // FIXME: SpellCheck error.
		return false
	}
	if ruOk > 0 && ruOk > enOk {
		return true
	}
	return false
}

// IsContainsURL checks if string contains http(s):// or ftp(s)://
func IsContainsURL(msg string) bool {
	re := regexp.MustCompile(`(https?|ftps?)://.+`)
	if re.MatchString(msg) {
		return true
	}
	return false
}

// SpellCheck runs aspell against <msg> with <dict>
// correct == '*'
// bad == '&' + '#' (mispelled + unknown)
func SpellCheck(msg, dict string) (correct, bad int, err error) {
	if dict == "" {
		dict = "ru"
	}
	aspell := exec.Command("enchant", "-a", "-d "+dict)
	aspell.Stdin = strings.NewReader(msg)

	var out bytes.Buffer
	aspell.Stdout = &out

	if err = aspell.Run(); err != nil {
		fmt.Print(err)
		return
	}
	str := out.String()
	correct = strings.Count(str, "*")
	unknown := strings.Count(str, "# ")
	misspelled := strings.Count(str, "& ")
	bad = unknown + misspelled

	//DEBUG: fmt.Printf("\nmsg(%s): %q OK: %d, BAD: %d, UNK: %d\n", str, dict, correct, misspelled, unknown)
	return
}
