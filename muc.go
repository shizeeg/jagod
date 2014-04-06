package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shizeeg/xmpp"
)

type Conference struct {
	JID       string
	Password  string
	Parser    GluxiParser
	Joined    bool
	Occupants []Occupant
}

type Occupant struct {
	Nick  string // room nickname. REQUIRED
	JID   string // optional (if available)
	Affil string // owner, admin, member, none
	Role  string // moderator, participant, visitor, none
}

// IsNick checks if given string is actual nick
// returns index in Conference.Occupants and true, -1, false otherwise.
func (c *Conference) NickIndex(nick string) (index int, found bool) {
	for i, o := range c.Occupants {
		if nick == o.Nick {
			return i, true
		}
	}
	return -1, false
}

// GetOccupantByNick returns index and actual Occupant if existed in conference
func (c *Conference) GetOccupantByNick(nick string) (i int, occ Occupant) {
	for i, o := range c.Occupants {
		if nick == o.Nick {
			return i, o
			break
		}
	}
	return -1, Occupant{}
}

// ChompNick cuts off <nick> part from "nick: message"
func (c *Conference) ChompNick(msg string) string {
	if pos := strings.IndexAny(msg, c.Parser.NickSuffixes); pos > 0 {
		nick := msg[:pos]
		for _, o := range c.Occupants {
			if nick == o.Nick {
				return msg[pos+1:]
			}
		}
	}
	return msg
}

// OccupantAdd adds occupant from <presence> stanza sent on somebody new joins the conference
func (c *Conference) OccupantAdd(stanza *xmpp.MUCPresence) (added Occupant, err error) {
	if len(stanza.X) <= 0 {
		return Occupant{}, errors.New("no <x /> child in stanza!")
	}
	x := stanza.X[0]
	var nick string
	if tmp := strings.SplitN(stanza.From, "/", 2); len(tmp) == 2 {
		nick = tmp[1]
	}
	add := func(o Occupant) {
		found := false
		for i, p := range c.Occupants {
			if o.Nick == p.Nick {
				c.Occupants[i] = o
				found = true
				break
			}
		}
		if !found {
			c.Occupants = append(c.Occupants, o)
		}
	}

	for _, item := range x.Items {
		added := Occupant{
			Affil: item.Affil,
			Role:  item.Role,
			Nick:  item.Nick,
			JID:   item.JID,
		}
		if len(item.Nick) == 0 {
			added.Nick = nick
		}
		add(added)
	}
	if stanza.IsCode(xmpp.SELF) {
		c.Parser.OwnNick = nick
	}
	return
}

func (c *Conference) OccupantDel(stanza *xmpp.MUCPresence) (deleted Occupant, err error) {
	if len(stanza.X) == 0 {
		return Occupant{}, errors.New("no <x /> child in stanza!")
	}
	//x := stanza.X[0]
	var nick string
	if tmp := strings.SplitN(stanza.From, "/", 2); len(tmp) == 2 {
		nick = tmp[1]
	}
	for i, o := range c.Occupants {
		if nick == o.Nick {
			deleted := c.Occupants[i]
			copy(c.Occupants[i:], c.Occupants[i+1:])
			c.Occupants = c.Occupants[:len(c.Occupants)-1]
			return deleted, nil
		}
	}
	return Occupant{}, errors.New(fmt.Sprintf("nick: %q not in %q!", nick, c))
}

//func (c *Conference) NickToJids(nick string, last bool) (jids []string, err error) {
//
//}
