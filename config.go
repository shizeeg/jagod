package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shizeeg/gcfg"
	"github.com/shizeeg/xmpp"
)

// Commands represents [cmd] config section.
type Commands struct {
	DisableInvite  bool `gcfg:"disable_invite"`
	DisablePing    bool `gcfg:"disable_ping"`
	DisableSpell   bool `gcfg:"disable_spell"`
	DisableTr      bool `gcfg:"disable_tr"`
	DisableTurn    bool `gcfg:"disable_turn"`
	DisableTurnURL bool `gcfg:"disable_turn_url"`
	DisableVersion bool `gcfg:"disable_version"`
	DisableLeave   bool `gcfg:"disable_leave"`
}

// Config configuration file structure
type Config struct {
	// Log struct {
	//	File     string `gcfg:"file"`
	//	Chatlogs string `gcfg:"chatlogs"`
	// }

	Account struct {
		User              string        `gcfg:"user"`
		Password          string        `gcfg:"password"`
		Server            string        `gcfg:"server"`
		Port              string        `gcfg:"port"`
		Resource          string        `gcfg:"resource"`
		FingerprintSHA256 string        `gcfg:"fingerprintsha256"`
		Trusted           bool          `gcfg:"Trusted"`
		SkipTLS           bool          `gcfg:"SkipTLS"`
		Priority          string        `gcfg:"priority"`
		Keepalive         time.Duration `gcfg:"keepalive"`
		JID               string        `gcfg:"jid"`
	}
	Access struct {
		Owners []string `gcfg:"owner"`
	}
	MUC struct {
		Nick         string `gcfg:"nick"`
		Prefix       string `gcfg:"prefix"`
		NickSuffixes string `gcfg:"nick_suffixes"`
		LeaveMinJIDs string `gcfg:"leave_minjids"`
	}
	Database struct {
		Type     string `gcfg:"type"`
		Server   string `gcfg:"server"`
		Port     string `gcfg:"port"`
		Password string `gcfg:"password"`
		User     string `gcfg:"user"`
		Database string `gcfg:"database"`
	}
	Yandex struct {
		DictAPI    string `gcfg:"dictapi"`
		RespLang   string `gcfg:"response_lang"`
		SpellLangs string `gcfg:"spell_langs"`
	}
	Cmd Commands
	// pointer to config for xmpp.Dial()
	xmppConfig *xmpp.Config
}

// ReadFile reads config from filename
// cfg := Config{}
// if err := cfg.ReadFile("my.conf"); err != nil {
//	log.Fataln(err)
// }
func (cfg *Config) ReadFile(filename string) error {
	if err := gcfg.ReadFileInto(cfg, filename); err != nil {
		return err
	}
	cfg.xmppConfig = &xmpp.Config{}
	cfg.xmppConfig.Resource = cfg.Account.Resource
	cfg.xmppConfig.SkipTLS = cfg.Account.SkipTLS
	cfg.xmppConfig.TrustedAddress = cfg.Account.Trusted
	cfg.xmppConfig.ServerCertificateSHA256 = cfg.FingerprintToBytes()
	return nil
}

// FingerprintToBytes converts SHA256 fingerprint from "AA:BB:CC:DD" format to bytes array
func (cfg *Config) FingerprintToBytes() []byte {
	var out []byte
	fingerprint := strings.TrimSpace(cfg.Account.FingerprintSHA256)
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
