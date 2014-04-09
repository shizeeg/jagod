package main

import (
	"strings"
)

// SplitJID returns barejid and resource from FullJID
// if jid have no /resource part it returns jid and empty string
func SplitJID(jid string) (barejid, resource string) {
	if tmp := strings.SplitN(jid, "/", 2); len(tmp) == 2 {
		return tmp[0], tmp[1]
	}
	return jid, ""
}

// IsValidJid checks if JID is valid
// fullJid == true adds additional check for /resource
func IsValidJID(jid string) bool {
	if len(jid) < 5 { // a@b.c
		return false
	}
	atpos := strings.Index(jid, "@")
	dotpos := strings.Index(jid, ".")
	// slashpos := strings.Index(jid, "/")
	// fmt.Println(atpos, dotpos, slashpos)
	if atpos > 0 && dotpos > 2 && len(jid) > strings.Index(jid, ".")+1 {
		return true
	}
	return false
}
