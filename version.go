// +build linux !windows
// +build !freebsd

package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

// Version returns our version, os and Go runtime info.
func Version() (os string, rt string) {
	var osinfo syscall.Utsname
	if err := syscall.Uname(&osinfo); err == nil {
		os = fmt.Sprintf("%s %s %s", Atos(osinfo.Sysname),
			Atos(osinfo.Release), Atos(osinfo.Machine))
	}
	if lsb, err := exec.Command("lsb_release", "-ds").Output(); err == nil {
		os = fmt.Sprintf("%s %s %s", strings.Trim(string(lsb), "\"\n"), Atos(osinfo.Release),
			Atos(osinfo.Machine))
	}

	rt = fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return os, BOTVERSION + " (" + rt + ")"
}

// Atos converts syscall.Utsname field to string.
func Atos(ca [65]int8) string {
	s := make([]byte, len(ca))
	var c int
	for ; c < len(ca); c++ {
		if ca[c] == 0 {
			break
		}
		s[c] = byte(ca[c])
	}
	return string(s[0:c])
}
