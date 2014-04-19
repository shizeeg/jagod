package main

import (
	"strings"
	"time"
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

// GetTImeDate return timezone (tzo) and time (utc) as defined in XEP-0202
// http://xmpp.org/extensions/xep-0202.html
func GetTimeDate() (tzo, utc string) {
	TimeFormat := "2006-01-02T15:04:05Z"
	now := time.Now()
	utc = now.UTC().Format(TimeFormat)
	tzo = now.Format("-07:00")
	return
}
// GetTimeOld returns timezone, datetime and display suggestion as defined in
// XEP-0090 http://xmpp.org/extensions/xep-0090.html
func GetTimeDateOld() (tz, utc, display string) {
	TimeFormat := "20060102T15:04:05"
	now := time.Now()
	utc = now.UTC().Format(TimeFormat)
	display = now.Format(time.RubyDate)
	tz = now.Format("MST")
	return
}
