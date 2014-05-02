package main

import (
	"bytes"
	"log"
	"os/exec"

	"github.com/shizeeg/xmpp"
)

// RunPlugin is a generic interface for external commands.
func (s *Session) RunPlugin(stanza xmpp.Stanza, filename string, params ...string) {
	message, ok := stanza.Value.(*xmpp.ClientMessage)
	if !ok {
		log.Println("Wrong Stanza type!")
		return
	}
	lang := "-lang=en"
	if len(message.Lang) > 0 {
		lang = "-lang="+message.Lang
	}
	plugin := exec.Command(filename, lang)
	plugin.Args = append(plugin.Args, params...)

	var out bytes.Buffer
	plugin.Stdout = &out
	if err := plugin.Run(); err != nil {
		log.Println(err)
		return
	}
	s.Say(stanza, out.String(), false)
}
