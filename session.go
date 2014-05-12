package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/shizeeg/xmpp"
)

type Session struct {
	conn              *xmpp.Conn
	config            *Config
	conferences       map[string]Conference
	pendingSubscribes map[string]string
	// timeouts maps from Cookies (from outstanding requests) to the
	// absolute time when that request should timeout.
	timeouts map[xmpp.Cookie]time.Time
}

func (s *Session) Say(stanza interface{}, msg string, private bool) (err error) {
	typ := "chat"
	var to string
	stnza, ok := stanza.(xmpp.Stanza)
	if !ok {
		msg = "error stanza type!"
	}
	switch st := stnza.Value.(type) {
	case *xmpp.ClientMessage:
		typ = st.Type
		to = st.From
		if !private && typ == "groupchat" {
			var nick string
			to, nick = SplitJID(st.From)
			msg = nick + ": " + msg
		}
		if private {
			typ = "chat"
			to = st.From
		}
	default:
		fmt.Printf("s.Say: unknown stanza: %#v\nmessage: %q\n", stanza, msg)
	}
	fmt.Printf("\nOn %#v\nSAY(to=%s, typ=%s, msg=%s)\n", stanza, to, typ, msg)
	return s.conn.SendMUC(to, typ, msg)
}

// GetInfoReply returns xmpp.DiscoveryReply with Identity and supported features.
func (s *Session) GetInfoReply() *xmpp.DiscoveryReply {
	reply := xmpp.DiscoveryReply{
		Node: NODE, // add verification string later
		Identities: []xmpp.DiscoveryIdentity{
			{
				Category: "client",
				Type:     "pc",
				Name:     BOTNAME,
			},
		},
		Features: []xmpp.DiscoveryFeature{
			{Var: "http://jabber.org/protocol/caps"},
			{Var: "http://jabber.org/disco#info"},
			{Var: "http://jabber.org/protocol/muc"},
			{Var: "jabber:iq:version"},
			{Var: "urn:xmpp:ping"},
			{Var: "urn:xmpp:time"},
			{Var: "jabber:iq:time"},
		},
	}
	if vstr, err := reply.VerificationString(); err == nil {
		reply.Node = NODE + "#" + vstr
	}
	return &reply
}

// readMessages reads stanza from channel and returns it
func (s *Session) readMessages(stanzaChan chan<- xmpp.Stanza) {
	defer close(stanzaChan)

	for {
		stanza, err := s.conn.Next()
		if err != nil {
			log.SetPrefix("s.readMessages() ")
			log.Println(err)
			return
		}
		stanzaChan <- stanza
	}
}

// processPresence handle incoming presences
// it also handles MUC (XEP-0045) "onJoin" presences.
func (s *Session) processPresence(stanza *xmpp.MUCPresence) {
	//	if out, err := xml.Marshal(stanza); err != nil {
	//		log.SetPrefix("!!!ERROR!!! ")
	//		log.Printf("PRESENCE: %#v\n", err)
	//	} else {
	//		log.SetPrefix("PRESENCE ")
	//		log.Printf("%#v\n%s\n-- \n", stanza, out)
	//	}
	confJID := xmpp.RemoveResourceFromJid(stanza.From)
	switch stanza.Type {
	case "unavailable":
		if conf, ok := s.conferences[confJID]; ok {
			occupant, err := conf.OccupantDel(stanza)
			if err != nil {
				log.Println(err)
			}
			s.conferences[confJID] = conf
			// We has left conference
			if occupant.Nick == conf.Parser.OwnNick {
				conf, err := s.ConfDel(stanza)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Printf("We're %q has quit %q!\n", occupant.Nick, conf.JID)
			}
		}
	case "": // empty <presence>
		if len(stanza.X) <= 0 {
			return
		}
		x := stanza.X[0]
		if len(stanza.X) > 1 {
			fmt.Printf("SECOND X: %#v\n\n", stanza.X[1])
		}
		fmt.Printf("%#v", stanza.X)
		switch x.XMLName.Space + " " + x.XMLName.Local {
		case "http://jabber.org/protocol/muc#user x":
			fromJid := xmpp.RemoveResourceFromJid(stanza.From)
			//	fromJid = strings.ToLower(fromJid)
			for jid, conf := range s.conferences {
				if !conf.Joined && conf.JID == fromJid {
					conf.Joined = true
					s.conferences[jid] = conf
					msg := fmt.Sprintf("I have joined to %#v", conf)
					fmt.Println(msg) // FIXME: send it to requester
				}
			}
			if conf, ok := s.conferences[fromJid]; ok {
				conf.OccupantAdd(stanza)
				s.conferences[fromJid] = conf
			}
		default:
			fmt.Println("WTF?: ", stanza)
		}
	case "subscribe": // http://xmpp.org/rfcs/rfc6121.html#sub-request-gen
		jid := xmpp.RemoveResourceFromJid(stanza.From)
		if err := s.conn.SendPresence(jid, "subscribed", ""); err != nil {
			s.conn.Send(stanza.From, "Subscription error")
		}
		s.conn.SendPresence(jid, "subscribe", "")
	case "subscribed":
		s.conn.Send(stanza.From, "Hi!")
	case "unsubscribe":
		s.conn.Send(stanza.From, "Fuck you, then!")
	case "error":
		var msg string
		conf, err := s.ConfDel(stanza)
		if err != nil {
			fmt.Println(err)
		}

		bareJid, nick := SplitJID(stanza.From)
		switch stanza.Error.Any.Space + " " + stanza.Error.Any.Local {
		case "urn:ietf:params:xml:ns:xmpp-stanzas conflict":
			msg = fmt.Sprintf("Can't join %q with nick %q. Error %s: Nickname conflict!",
				bareJid, nick, stanza.Error.Code)
		case "urn:ietf:params:xml:ns:xmpp-stanzas not-authorized":
			msg = fmt.Sprintf("I can't join %q: %#v", conf.JID, stanza.Error)
		case "urn:ietf:params:xml:ns:xmpp-stanzas forbidden":
			msg = fmt.Sprintf("Can't join %q with nick %q. Error %s: I'm banned in this conference!",
				bareJid, nick, stanza.Error.Code)
		default:
			msg = fmt.Sprintf("We got error presence: type: %q, code %q: %#v",
				stanza.Error.Type, stanza.Error.Code, stanza.Error)
		}

		fmt.Println(msg)
		for _, j := range s.config.Access.Owners {
			if IsValidJID(j) {
				s.conn.Send(j, msg)
			}
		}
	default:
		log.SetPrefix("WARNING: ")
		if out, err := xml.Marshal(stanza); err == nil {
			log.Printf("Unknown presence stanza:\n %s\n", out) // FIXME: send it to requester
		}
	}
}

func (s *Session) awaitVersionReply(ch <-chan xmpp.Stanza, reqFrom xmpp.Stanza) {
	stanza, ok := <-ch
	reply, ok := stanza.Value.(*xmpp.ClientIQ)
	if !ok {
		return
	}
	bareJID, nick := SplitJID(reply.From)
	fromUser := reply.From
	replyType := "chat"
	if len(nick) > 0 {
		if conf, ok := s.conferences[bareJID]; ok {
			replyType = "groupchat"
			if i, ok := conf.NickIndex(nick); ok {
				fromUser = conf.Occupants[i].Nick
			}
		}
	}

	// fmt.Printf("awaitVersionReply()\nReqFrom: %#v\nReply: %#v\n", reqFrom, reply)
	_ = replyType
	// if !ok {
	// msg := fmt.Sprintf("Version request to %q timed out.", user)
	// say(msg, replyType)
	// return
	// }
	// //	if !ok {
	// msg := fmt.Sprintf("Version request to %q resulted in bad reply type.", user)
	// say(msg, replyType)
	// return
	// }

	if reply.Type == "error" {
		msg := fmt.Sprintf("%s %s", reply.Error.Code, reply.Error.Any.Local)
		if len(reply.Error.Text) > 0 {
			msg = fmt.Sprintf("%s %s", reply.Error.Code, reply.Error.Text)
		}
		s.Say(reqFrom, msg, false)
		return

	} else if reply.Type != "result" {
		msg := fmt.Sprintf("Version request to %q resulted in response with unknown type: %v", nick, reply.Type)
		s.Say(reqFrom, msg, false)
		return
	}

	buf := bytes.NewBuffer(reply.Query)

	var versionReply xmpp.VersionReply
	if err := xml.NewDecoder(buf).Decode(&versionReply); err != nil {
		msg := fmt.Sprintf("Failed to parse version reply from %q: %v", nick, err)
		s.Say(reqFrom, msg, false)
		return
	}
	var msg string
	if len(versionReply.Name) > 0 {
		msg += fmt.Sprintf("\nName:    %q", versionReply.Name)
	}
	if len(versionReply.Version) > 0 {
		msg += fmt.Sprintf("\nVersion: %q", versionReply.Version)
	}
	if len(versionReply.OS) > 0 {
		msg += fmt.Sprintf("\nOS:      %q", versionReply.OS)
	}
	if len(msg) == 0 {
		msg = "no data in reply to iq:version!"
	}
	msg = fmt.Sprintf("Version reply from %q: %s", fromUser, msg)
	fmt.Printf("From: %q\nTo: %q\nMsg: %q", reply.From, reply.To, msg)
	s.Say(reqFrom, msg, false)
}

func (s *Session) awaitTimeReply(ch <-chan xmpp.Stanza, reqFrom xmpp.Stanza) {
	stanza, ok := <-ch
	reply, ok := stanza.Value.(*xmpp.ClientIQ)
	if !ok {
		return
	}
	bareJID, nick := SplitJID(reply.From)
	fromUser := reply.From
	replyType := "chat"
	if len(nick) > 0 {
		if conf, ok := s.conferences[bareJID]; ok {
			replyType = "groupchat"
			if i, ok := conf.NickIndex(nick); ok {
				fromUser = conf.Occupants[i].Nick
			}
		}
	}
	_ = fromUser
	_ = replyType

	if reply.Type == "error" {
		msg := fmt.Sprintf("%s %s", reply.Error.Code, reply.Error.Any.Local)
		if len(reply.Error.Text) > 0 {
			msg = fmt.Sprintf("%s %s", reply.Error.Code, reply.Error.Text)
		}
		s.Say(reqFrom, msg, false)
		return

	} else if reply.Type != "result" {
		msg := fmt.Sprintf("Version request to %q resulted in response with unknown type: %v", nick, reply.Type)
		s.Say(reqFrom, msg, false)
		return
	}

	buf := bytes.NewBuffer(reply.Query)

	var timeReply xmpp.TimeReply
	if err := xml.NewDecoder(buf).Decode(&timeReply); err != nil {
		msg := fmt.Sprintf("Failed to parse time reply from %q: %v", nick, err)
		s.Say(reqFrom, msg, false)
		return
	}
	msg := fmt.Sprintf("It's %s on %s's clock.", timeReply.String(), fromUser)
	s.Say(reqFrom, msg, false)
}

func (s *Session) processIQ(stanza *xmpp.ClientIQ) interface{} {
	buf := bytes.NewBuffer(stanza.Query)
	parser := xml.NewDecoder(buf)
	token, _ := parser.Token()

	if token == nil {
		return nil
	}
	startElem, ok := token.(xml.StartElement)
	if !ok {
		return nil
	}
	switch startElem.Name.Space + " " + startElem.Name.Local {
	case "urn:xmpp:ping ping":
		fmt.Printf("URN:XMPP:PING: %#v", stanza)
		return xmpp.EmptyReply{}

	case "urn:xmpp:time time":
		tzo, utc := GetTimeDate()
		fmt.Println("urn:xmpp:time: ", utc, tzo)
		return xmpp.TimeReply{TZO: tzo, UTC: utc}

	case "http://jabber.org/protocol/disco#items query":
		return xmpp.DiscoInfoReply{}

	case "http://jabber.org/protocol/disco#info query":
		return s.GetInfoReply()

	case "jabber:iq:time query":
		tz, utc, disp := GetTimeDateOld()
		return xmpp.TimeReplyOld{
			UTC:     utc,
			TZ:      tz,
			Display: disp,
		}

	case "jabber:iq:version query":
		osver, gover := Version()
		reply := xmpp.VersionReply{
			Name:    BOTNAME,
			Version: gover,
			OS:      osver,
		}
		fmt.Printf("jabber:iq:version %#v\n", reply)
		return reply
		//	case "jabber:iq:roster query":
		//		if len(stanza.From) > 0 /*&& stanza.From != s.account */ {
		//	warn(s.term, "Ignoring roster IQ from bad address: "+stanza.From)
		//			fmt.Printf("WARN: Ignoring roster IQ from bad adress: %s", stanza.From)
		//			return nil
		//		}
		//		var roster xmpp.Roster
		//		if err := xml.NewDecoder(bytes.NewBuffer(stanza.Query)).Decode(&roster); err != nil || len(roster.Item) == 0 {
		//	warn(s.term, "Failed to parse roster push IQ")
		//			fmt.Printf("WARN: Failed to parse roster push IQ")
		//			return nil
		//		}
		//		entry := roster.Item[0]
		//
		//		if entry.Subscription == "remove" {
		//			for i, rosterEntry := range s.roster {
		//				if rosterEntry.Jid == entry.Jid {
		//					copy(s.roster[i:], s.roster[i+1:])
		//					s.roster = s.roster[:len(s.roster)-1]
		//				}
		//			}
		//			return xmpp.EmptyReply{}
		//		}
		//
		//		found := false
		//		for i, rosterEntry := range s.roster {
		//			if rosterEntry.Jid == entry.Jid {
		//				s.roster[i] = entry
		//				found = true
		//				break
		//			}
		//		}
		//		if !found {
		//			s.roster = append(s.roster, entry)
		//			s.input.AddUser(entry.Jid)
		//		}
		//		return xmpp.EmptyReply{}
	default:
		//	info(s.term, "Unknown IQ: "+startElem.Name.Space+" "+startElem.Name.Local)
		msg := fmt.Sprintf("Unknown IQ: %s %s\n", startElem.Name.Space, startElem.Name.Local)
		fmt.Println(msg)
	}

	return nil
}

// JoinMUC joins to a conference with nick & optional password
// init & add Conference to the s.conferences map
func (s *Session) JoinMUC(confJID, nick, password string) error {
	bareJID := xmpp.RemoveResourceFromJid(confJID)
	nick = strings.TrimSpace(nick)
	if len(nick) == 0 {
		nick = BOTNAME
	}

	for _, c := range s.conferences {
		if c.JID == bareJID {
			msg := fmt.Sprintf("I'm already in %q with nick %q", c.JID, c.Parser.OwnNick)
			return errors.New(msg)
		}
	}

	if s.conferences == nil {
		s.conferences = make(map[string]Conference)
	}
	// FIXME: fetch settings from database. Separate for each conference.
	parser := GluxiParser{
		OwnNick:      nick,
		NickSuffixes: s.config.MUC.NickSuffixes,
		Prefix:       s.config.MUC.Prefix,
	}
	conf := Conference{
		JID:      bareJID,
		Password: password,
		Parser:   parser,
	}
	s.conferences[bareJID] = conf
	msg := fmt.Sprintf("Conference %q with nick %q added",
		conf.JID, parser.OwnNick)
	if len(conf.Password) > 0 {
		msg += " and password: " + conf.Password
	}

	fmt.Println(msg + "!")
	ver, err := s.GetInfoReply().VerificationString()
	if err != nil {
		log.Println(err)
	}
	st := xmpp.MUCPresence{
		Lang: s.config.Account.Lang,
		To:   conf.JID + "/" + parser.OwnNick,
		Caps: &xmpp.ClientCaps{
			Hash: "sha-1",
			Node: NODE,
			Ver:  ver,
		},
		X: []*xmpp.X{&xmpp.X{
			XMLName:  xml.Name{Space: "http://jabber.org/protocol/muc", Local: "x"},
			Password: conf.Password,
			History:  &xmpp.History{MaxChars: "0"}}},
	}
	// fmt.Println("DEBUG: JoinMUC: {")
	// PrintXML(st)
	// fmt.Println("\n}")
	return s.conn.SendStanza(st)
}

// ConfDel deletes conference
func (s *Session) ConfDel(stanza *xmpp.MUCPresence) (deleted Conference, err error) {
	bareJID := xmpp.RemoveResourceFromJid(stanza.From)
	if conf, ok := s.conferences[bareJID]; ok {
		deleted = conf
		delete(s.conferences, conf.JID)
		log.Printf("Conference %q deleted!", conf.JID)
		return
	}
	return Conference{}, errors.New("No such conference! " + bareJID)
}
