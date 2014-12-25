package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shizeeg/ini"
	"github.com/shizeeg/xmpp"
)

// Config is a configuration file structure
type Config struct {
	// main config file pointer
	c *ini.File
	// pointer to config for xmpp.Dial()
	xmppConfig *xmpp.Config
}

// ReadFile reads config from filename
// cfg := Config{}
// if err := cfg.ReadFile("my.conf"); err != nil {
//	log.Fataln(err)
// }
func (cfg *Config) ReadFile(filename string) error {
	c, err := ini.Load(filename)
	if err != nil {
		return err
	}
	cfg.c = c
	acc := cfg.c.Section("account")
	cfg.xmppConfig = &xmpp.Config{}
	cfg.xmppConfig.Resource = acc.Key("resource").String()
	cfg.xmppConfig.SkipTLS = acc.Key("SkipTLS").MustBool()
	cfg.xmppConfig.TrustedAddress = acc.Key("trusted").MustBool()
	cfg.xmppConfig.ServerCertificateSHA256 = FingerprintToBytes(acc.Key("FingerprintSHA256").String())
	return nil
}

// FingerprintToBytes converts SHA256 fingerprint from "AA:BB:CC:DD" format to bytes array
func FingerprintToBytes(sha256 string) []byte {
	var out []byte
	fingerprint := strings.Trim(sha256, "\n\r\t \"")
	fprintlen := len(fingerprint)

	switch fprintlen {
	case 95: // SHA256 in AB:0C:AD format. 32 bytes w/ delimiters
		for i := 0; i < fprintlen; i += 3 {
			b, err := strconv.ParseUint(fingerprint[i:i+2], 16, 8)
			if err != nil {
				fmt.Printf("illegal byte %x @ offset %d!\n", b, i)
				return []byte(nil)
			}
			out = append(out, uint8(b))
		}
	case 64: // SHA256 in ABBCAD format (w/o delimiters)
		for i := 0; i < fprintlen; i += 2 {
			b, err := strconv.ParseUint(fingerprint[i:i+2], 16, 8)
			if err != nil {
				fmt.Printf("Illegal byte %x @ offset %d\n", b, i)
				return []byte(nil)
			}
			out = append(out, uint8(b))

		}
	default:
		fmt.Printf("Wrong fingerprint! Expect SHA256 hash, 32 bytes long but got %q %d bytes long!\n",
			fingerprint, len(fingerprint))
	}
	return out
}
