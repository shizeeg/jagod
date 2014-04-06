package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/shizeeg/xmpp"
)

var (
	cfgfile string
	pidfile string
)

func init() {
	flag.StringVar(&cfgfile, "c", "/etc/jagod.cfg", "main configuration file.")
	flag.StringVar(&pidfile, "pid", "/var/run/jagod.pid", "pidfile")
	flag.Parse()
}

func main() {
	// FIXME: oh, shi~ All this pid-stuff is just a mess..
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, os.Kill)

		exitMsg := <-signalChan
		fmt.Printf("Got signal: %q. Exiting.\n", exitMsg)
		f, _ := os.Create(pidfile)
		defer f.Close()
		f.WriteString("")
		os.Exit(0)
	}()

	pid := os.Getpid()
	if f, err := os.Open(pidfile); err != nil {
		log.Println(err)
	} else {
		buf := make([]byte, 256)
		if _, err := f.Read(buf); err != nil {
			// first launch probably
			//log.Printf("WARNING: Can't read pid: %s\n", err)
		}
		if oldPid, err := strconv.Atoi(strings.Trim(string(buf), "\x00")); err == nil {
			if oldPid > 0 {
				proc, err := os.FindProcess(oldPid)
				fmt.Printf("PROC: %#v\n%#v\n", proc, err)
				if err == nil {
					log.Printf("pid %d read from file, check if we're already running...", proc.Pid)
					// check for "no such process" error
					if err := proc.Signal(syscall.Signal(0)); err == nil || err.Error() != "no such process" {
						log.Fatalf("We're running on PID %d\n", proc.Pid)
						// if err := proc.Kill(); err != nil {
						// 	log.Fatalf("Can't kill PID %d: %s\n", proc.Pid, err.Error())
						// }
					} else {
						log.Println("SIG_0: ", err)
					}
				}
			}
		}
		f.Close()

		if f, err = os.Create(pidfile); err != nil {
			log.Printf("WARNING: %s", err)
		} else {
			if _, err := f.WriteString(fmt.Sprintf("%d", pid)); err != nil {
				log.Printf("WARNING: %q", err.Error())
			} else {
				log.Printf("PID: %d\n", pid)
			}
			f.Close()
		}
	}
	s := Session{
		config: &Config{},
	}
	// FIXME: need we respect XDG_CONFIG_DIRS?
	if err := s.config.ReadFile(cfgfile); err != nil {
		log.Fatal(err)
	}
	conn, err := xmpp.Dial(
		s.config.Account.Server+":"+s.config.Account.Port,
		s.config.Account.User,
		s.config.Account.Server,
		s.config.Account.Password,
		s.config.xmppConfig)
	if err != nil {
		fmt.Printf("cant connect! %v", err)
		return
	}
	s.conn = conn
	s.conn.SignalPresence("")

	_, _, err = s.conn.RequestRoster()
	parser := GluxiParser{
		Prefix:       s.config.MUC.Prefix,
		NickSuffixes: s.config.MUC.NickSuffixes,
		OwnNick:      s.config.MUC.Nick,
	}
	stanzaChan := make(chan xmpp.Stanza)
	go s.readMessages(stanzaChan)

	s.timeouts = make(map[xmpp.Cookie]time.Time)
	s.conferences = make(map[string]Conference)

	ticker := time.NewTicker(1 * time.Second)
	pingticker := time.NewTicker(s.config.Account.Keepalive * time.Second)

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
				msg := st.Body
				if st.Type == "groupchat" && !strings.HasSuffix(st.From, parser.OwnNick) {
					conf, ok := s.conferences[xmpp.RemoveResourceFromJid(st.From)]
					msg := conf.ChompNick(msg) // chomp "nick:" first
					if !ok {
						continue
					}
					if !s.config.Cmd.Disable_turn_url && IsContainsURL(msg) {
						if link, err := url.QueryUnescape(msg); err == nil {
							if strings.Count(msg, "%") > 3 {
								s.conn.SendMUC(conf.JID, "groupchat", link)
							}
						}
					} else if !s.config.Cmd.Disable_turn && IsWrongLayout(msg) {
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
						s.JoinMUC(conf, parser.OwnNick, password)
					case "INVITE": // FIXME: need to check XEP-0249 support first
						if s.config.Cmd.Disable_invite {
							continue
						}
						conf, ok := s.conferences[xmpp.RemoveResourceFromJid(st.From)]
						if !ok {
							s.conn.Send(st.From, "it can't be used in roster!")
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
					case "SPELL":
						if s.config.Cmd.Disable_spell {
							continue
						}
						req := strings.Join(parser.Tokens[2:], " ")
						text := YandexSpell(s.config.Yandex.SpellLangs, req)
						if len(text) == 0 {
							text = "no responce from yandex.ru"
						}
						s.Say(stanza, text, false)
					case "TR":
						if s.config.Cmd.Disable_tr {
							continue
						}
						go func() {
							var text, lang string
							if len(parser.Tokens) > 2 {
								lang = parser.Tokens[2]
							}
							if lang == "help" {
								text = strings.Join(GetLangs(s.config.Yandex.DictAPI), ", ")
							} else if len(parser.Tokens) > 3 {
								req := strings.Join(parser.Tokens[3:], " ")
								text = YandexDic(s.config.Yandex.DictAPI, s.config.Yandex.RespLang, lang, req)
							} else {
								text = "tr <lang-lang>|<help> <text>"
							}

							if len(text) == 0 {
								text = "no responce from yandex.ru"
							}
							s.Say(stanza, "http://api.yandex.ru/dictionary/\n"+text, false)
						}()

					case "VERSION":
						if s.config.Cmd.Disable_version {
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
					case "PING":
						if s.config.Cmd.Disable_ping {
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
								s.Say(stanza, msg, false)
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
					s.Say(stanza, msg, false)
				}
			case *xmpp.MUCPresence:
				s.processPresence(st)
			}
		default:
		}
		time.Sleep(150 * time.Millisecond) // release CPU
	}

}
