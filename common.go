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
