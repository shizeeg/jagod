package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/shizeeg/xmpp"
)

var (
	cfgfile string
	pidfile string
	pid     = PID{}
)

func init() {
	flag.StringVar(&cfgfile, "c", "/etc/jagod.cfg", "main configuration file.")
	flag.StringVar(&pidfile, "p", "/var/run/jagod.pid", "pidfile")
	flag.Parse()
	pid.FileName = pidfile
	if err := pid.Write(); err != nil {
		log.Printf("[WARNING] unable to write pidfile: %s", err.Error())
	}
}

func main() {
	s := Session{
		config: &Config{},
	}

	err := s.config.ReadFile(cfgfile)
	if err != nil {
		log.Fatal(err)
	}
	// FIXME: need we respect XDG_CONFIG_DIRS?
	//if err := s.config.ReadFile(cfgfile); err != nil {
	//	log.Fatal(err)
	//}
	account := s.config.c.Section("account")
	conn, err := xmpp.Dial(
		account.Key("server").String()+":"+account.Key("port").String(),
		account.Key("user").String(),
		account.Key("server").String(),
		account.Key("password").String(),
		s.config.xmppConfig)
	if err != nil {
		fmt.Printf("cant connect! %v", err)
		return
	}
	s.conn = conn
	// s.conn.SignalPresence("")
	ver, _ := s.GetInfoReply().VerificationString()

	s.conn.SendStanza(
		xmpp.ClientPresence{
			Lang: account.Key("lang").String(),
			Caps: &xmpp.ClientCaps{
				Hash: "sha-1",
				Node: NODE,
				Ver:  ver,
			},
		},
	)

	_, _, err = s.conn.RequestRoster()
	muc := s.config.c.Section("muc")
	parser := GluxiParser{
		Prefix:       muc.Key("prefix").String(),
		NickSuffixes: muc.Key("nick_suffixes").String(),
		OwnNick:      muc.Key("nick").String(),
	}
	stanzaChan := make(chan xmpp.Stanza)
	go s.readMessages(stanzaChan)

	s.timeouts = make(map[xmpp.Cookie]time.Time)
	s.conferences = make(map[string]Conference)

	ticker := time.NewTicker(1 * time.Second)
	pingticker := time.NewTicker(time.Duration(account.Key("keepalive").MustInt()) * time.Second)

	fmt.Println(muc.Key("autojoin").Strings("\n"))
	for _, joinTo := range muc.Key("autojoin").Strings("\n") {
		confJID := joinTo
		pass := ""
		if tmp := strings.SplitN(joinTo, ",", -1); len(tmp) == 2 {
			confJID = strings.TrimSpace(tmp[0])
			pass = strings.TrimSpace(tmp[1])
		}
		bareJID, nick := SplitJID(confJID)
		if len(nick) == 0 {
			nick = parser.OwnNick
		}
		if err := s.JoinMUC(bareJID, nick, pass); err != nil {
			for _, j := range s.config.c.Section("access").Key("owners").Strings("\n") {
				if IsValidJID(j) { // FIXME: temorary code.
					s.conn.Send(j, "autojoin: "+err.Error())
				}
			}
		}
	}

	plugins := s.config.c.Section("plugins").KeysHash(true)
	filters := s.config.c.Section("filters")
	filterGetTitle, _ := filters.Key("gettitle").Bool()
	filterURLUnescape, _ := filters.Key("url_unescape").Bool()
	filterTurn, _ := filters.Key("turn").Bool()

	for {
		select {
		case now := <-ticker.C:
			haveExpired := false
			for _, expiry := range s.timeouts {
				if now.After(expiry) {
					haveExpired = true
					break
				}
			}
			if !haveExpired {
				continue
			}

			newTimeouts := make(map[xmpp.Cookie]time.Time)
			for cookie, expiry := range s.timeouts {
				if now.After(expiry) {
					s.conn.Cancel(cookie)
				} else {
					newTimeouts[cookie] = expiry
				}
			}
			s.timeouts = newTimeouts

		case tick := <-pingticker.C:
			s.conn.KeepAlive()
			_ = tick
		case stanza, ok := <-stanzaChan:
			if !ok {
				fmt.Printf("Error: %v", err)
				return // Bail out. We're disconnected.
			}

			switch st := stanza.Value.(type) {
			case *xmpp.ClientMessage:
				if st.IsDelayed() || len(st.Subject) > 0 { // ignore history (delayed) and topic messages
					continue
				}
				msg := st.Body
				if filterGetTitle && (!strings.HasSuffix(st.From, parser.OwnNick) && IsContainsURL(msg)) {
					message := strings.Replace(msg, "\n", " ", -1)
					for _, word := range strings.Split(message, " ") {
						if len(word) > 4 && word[0:4] == "http" {
							go s.RunPlugin(stanza, "gettitle", false, word)
							break
						}
					}
				}

				if st.Type == "groupchat" && !strings.HasSuffix(st.From, parser.OwnNick) {
					conf, ok := s.conferences[xmpp.RemoveResourceFromJid(st.From)]
					msg := conf.ChompNick(msg) // chomp "nick:" first
					if !ok {
						continue
					}
					if filterURLUnescape && IsContainsURL(msg) {
						if link, err := url.QueryUnescape(msg); err == nil {
							if strings.Count(msg, "%") > 3 {
								s.conn.SendMUC(conf.JID, "groupchat", link)
							}
						}
					} else if filterTurn && IsWrongLayout(msg) {
						// fmt.Printf("MSG WRONG LAYOUT: %q\n%#v", msg, st)
						s.conn.SendMUC(conf.JID, "groupchat", Turn(msg))
					}
				}
				if parser.Init(st); parser.IsForMe {
					if len(parser.Tokens) <= 1 {
						continue
					}
					CMD := strings.ToUpper(parser.Tokens[1]) // FIXME:
					var toNick, toJID, fromNick, confJID string
					toJID = st.From
					_ = fromNick
					_ = confJID
					bareJID, resource := SplitJID(st.From)
					conf, ok := s.conferences[bareJID]
					if ok {
						confJID = bareJID
						if len(parser.Tokens) >= 3 {
							param := parser.Tokens[2]
							if _, ok := conf.NickIndex(param); ok {
								toJID = conf.JID + "/" + param
							} else if strings.Contains(param, ".") { // requesting from domain
								toJID = param
							}
						} else { // w/o params. User requesting himself
							toJID = st.From
							fromNick = resource
						}
					}
					// if st.Type == "chat" { // actual st stanza received from MUC
					// 	toJID = st.From
					// }
					switch CMD {
					// Run external commands (CLI)
					default:
						if cli, ok := plugins[CMD]; ok {
							plugin := strings.Split(cli, " ")
							params := parser.Tokens[2:]
							if len(plugin) > 1 {
								params = append(plugin[1:], parser.Tokens[2:]...)
							}
							go s.RunPlugin(stanza, plugin[0], true, params...)
						}

					case "JOIN":
						conf := parser.Tokens[2]
						parts := strings.Split(conf, "/")
						var password string

						if len(parts) == 2 {
							conf = parts[0]
							parser.OwnNick = parts[1]
						}
						if len(parser.Tokens) == 4 { // user specify a password
							password = parser.Tokens[3]
						}
						if err := s.JoinMUC(conf, parser.OwnNick, password); err != nil {
							for _, j := range s.config.c.Section("access").Key("owners").Strings("\n") {
								if IsValidJID(j) { // FIXME: temorary code.
									s.conn.Send(j, err.Error())
								}
							}
						}

					case "LEAVE", "EXIT", "QUIT", "PART":
						conf, ok := s.conferences[xmpp.RemoveResourceFromJid(st.From)]
						param1, err := parser.Token(2)
						param1 = xmpp.RemoveResourceFromJid(param1)
						status := "I'm quit!"
						if param2, err := parser.Token(3); err == nil {
							status = param2
						}
						if !ok {
							if err != nil {
								s.Respond(stanza, "I'm not in conference!", false)
								continue
							}
							conf, ok = s.conferences[param1]
							if !ok {
								s.Respond(stanza, "I'm not in "+param1, false)
								continue
							}
						}
						if err == nil {
							tmp, ok := s.conferences[param1]
							if !ok {
								s.Respond(stanza, "I'm not in "+param1, false)
								continue
							}
							conf = tmp
						}

						if err := s.conn.LeaveMUC(conf.JID+"/"+conf.Parser.OwnNick, status); err != nil {
							s.Respond(stanza, err.Error(), false)
							continue
						}
						s.Respond(stanza, "I'm quit from "+conf.JID, false)

					case "INVITE": // FIXME: need to check XEP-0249 support first
						if s.config.c.Section("internal").Key("invite").MustBool() {
							continue
						}
						conf, ok := s.conferences[xmpp.RemoveResourceFromJid(st.From)]
						if !ok {
							s.conn.Send(st.From, "Can't be used in a roster!")
							continue
						}
						var to, reason string
						if len(parser.Tokens) >= 3 {
							to = parser.Tokens[2]
							if len(parser.Tokens) >= 4 {
								reason = parser.Tokens[3]
							}
						}
						// FIXME: resolve jid from nick (db stuff)
						// fmt.Printf("INVITE: to %q\nconf: %q\npass: %q\nreason: %q\n", to, conf.JID, conf.Password, reason)
						s.conn.DirectInviteMUC(to, conf.JID, conf.Password, reason)

					case "VERSION":
						if s.config.c.Section("internal").Key("version").MustBool() {
							continue
						}
						fmt.Printf("TONICK: %q\n", toNick)
						replyChan, cookie, err := s.conn.SendIQ(toJID, "get", xmpp.VersionQuery{})
						if err != nil {
							fmt.Printf("Error sending iq:version request %v %d\n", err, cookie)
						}
						//fmt.Println(st)
						// FIXME: move it to the s.Conferences[jid].timeouts[cookie]
						s.timeouts[cookie] = time.Now().Add(5 * time.Second)
						go s.awaitVersionReply(replyChan, stanza)

					case "TIME":
						if s.config.c.Section("internal").Key("time").MustBool() {
							continue
						}
						replyChan, cookie, err := s.conn.SendIQ(toJID, "get", xmpp.TimeQuery{})
						if err != nil {
							fmt.Printf("Error sending urn:xmpp:time request %#v\n, %d\n", err, cookie)
						}
						s.timeouts[cookie] = time.Now().Add(5 * time.Minute)
						go s.awaitTimeReply(replyChan, stanza)

					case "PING":
						if s.config.c.Section("internal").Key("ping").MustBool() {
							continue
						}
						replyChan, cookie, err := s.conn.SendIQ(toJID, "get", xmpp.PingQuery{})
						if err != nil {
							fmt.Printf("Error sending urn:xmpp:ping request %v %d\n", err, cookie)
						}
						s.timeouts[cookie] = time.Now().Add(10 * time.Minute)
						go func(cookie xmpp.Cookie) {
							r := <-replyChan
							t := s.timeouts[cookie]
							var msg string
							elapsed := time.Now().Sub(t) + 10*time.Minute
							switch iq := r.Value.(type) {
							case *xmpp.ClientIQ:
								switch iq.Type {
								case "error": // FIXME: human readable error
									buf := bytes.NewBuffer(iq.Query)
									msg = fmt.Sprintf("IQ-reply:\n%s\n", buf)
								case "result":
									from := iq.From
									if st.Type == "groupchat" && strings.HasPrefix(iq.From, conf.JID) {
										from = strings.SplitN(iq.From, "/", 2)[1]
									}
									msg = fmt.Sprintf("pong from %q after %.3f seconds.", from, elapsed.Seconds())

								}
								s.Respond(stanza, msg, false)
							}
						}(cookie)
					}
				}
			case *xmpp.ClientIQ:
				if st.Type != "get" && st.Type != "set" {
					continue
				}
				reply := s.processIQ(st)
				if reply == nil {
					reply = xmpp.ErrorReply{
						Type:  "cancel",
						Error: xmpp.ErrorBadRequest{},
					}
				}
				if err := s.conn.SendIQReply(st.From, "result", st.Id, reply); err != nil {
					msg := fmt.Sprintf("Failed to send IQ message: %#v", err)
					s.Respond(stanza, msg, false)
				}
			case *xmpp.MUCPresence:
				s.processPresence(st)
			}
		default:
		}
		time.Sleep(150 * time.Millisecond) // release CPU
	}

}
