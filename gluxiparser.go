package main

import (
	"errors"
	"github.com/shizeeg/xmpp"
	"strings"
)

type GluxiParser struct {
	Prefix       string
	IsForMe      bool
	index        int
	Separator    string
	OwnNick      string
	NickSuffixes string
	FirstEOLpos  int
	Tokens       []string
}

// IsMyMessage checks if message starts with our nick
func (p *GluxiParser) IsMyMessage(cmsg *xmpp.ClientMessage) bool {
	body := strings.TrimSpace(cmsg.Body)
	nlen := len(p.OwnNick)

	if nlen > 0 {
		if strings.HasPrefix(body, p.OwnNick) && nlen < len(body) {
			suf := string(body[nlen])
			if strings.Contains(p.NickSuffixes, suf) {
				return true
			}
		}
	}
	return false
}

// Init inits parser.
func (p *GluxiParser) Init(cmsg *xmpp.ClientMessage) error {
	p.IsForMe = false
	/*	if cmsg.Type == "groupchat" {
			p.IsForMe = true
		}
	*/
	body := strings.TrimSpace(cmsg.Body)
	if strings.HasPrefix(body, p.Prefix) {
		p.IsForMe = true
		body = strings.TrimPrefix(body, p.Prefix)
		// firsttoken remove prefix
		// ?
	}

	nlen := len(p.OwnNick)
	if nlen > 0 {
		if p.IsMyMessage(cmsg) {
			suf := string(body[nlen])
			body = strings.TrimPrefix(body, p.OwnNick)
			body = strings.TrimPrefix(body, suf)
			body = strings.TrimSpace(body)
			body = strings.TrimPrefix(body, p.Prefix)
			p.IsForMe = true
		}
	}

	if p.Separator == "" && strings.ContainsRune(body, '\n') {
		a := strings.Split(body, "\n")
		b := strings.Split(a[0], " ")
		p.FirstEOLpos = len(b)
		p.Tokens = append(b, a[1:]...)
		p.Separator = "\n"
	} else { // command without newlines
		if p.Separator == "" {
			p.Separator = " "
			p.Tokens = strings.Split(body, p.Separator)
			p.FirstEOLpos = len(p.Tokens)
		} else { // with separator supplied
			p.Tokens = strings.Split(body, p.Separator)
		}

	}
	p.Separator = "" // ready for a new message
	return nil
}

// Token returns token
// You can use negative values to get tokens from last
func (p *GluxiParser) Token(index int) (token string, err error) {
	length := len(p.Tokens)
	abs := func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	}
	if length > abs(index) {
		if index > 0 {
			token = p.Tokens[index]
			if len(strings.TrimSpace(token)) == 0 {
				err = errors.New("Empty token!")
			}
			return
		}
		token = p.Tokens[length-index]
		if len(strings.TrimSpace(token)) == 0 {
			err = errors.New("Empty token")
		}
		return
	}
	return "", errors.New("Index out of bound!")
}
