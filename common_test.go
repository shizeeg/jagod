package main

import (
	"testing"
)

var (
	jids = map[string]bool{
		"x@x.com":                  true,
		"вася@джаббер.рф":          true,
		"user@server.tld/resource": true,
		"@x.com":                   false,
		"x.com":                    false,
		"x@":                       false,
		"@":                        false,
		".com/res":                 false,
		"x@.com":                   false,
		".com@server":              false,
		"@s.com/res":               false,
		"s.com/res@a":              false,
	}
)

func TestJid(t *testing.T) {
	for k, v := range jids {
		if ok := IsValidJID(k); ok != v {
			t.Errorf("test failed on: %s=%v, expected %v", k, ok, v)
		}
	}
}
