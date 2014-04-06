// +build !linux windows

package main

import (
	"fmt"
	"runtime"
	"strings"
	"syscall"
)

// Version returns bot version, os and Go runtime info on Windows.
func Version() (os string, rt string) {
	os = runtime.GOOS
	var flavor string
	if ver, err := syscall.GetVersion(); err == nil {
		var major, minor, build = byte(ver), uint8(ver >> 8), uint16(ver >> 16)
		switch {
		case major == 4:
			switch minor {
			case 0:
				flavor = "NT"
			case 10:
				flavor = "98"
			case 90:
				flavor = "Me"
			}
		case major == 5:
			switch {
			case minor == 2:
				flavor = "2003"
			case minor == 1 && build == 2600:
				flavor = "XP"
			case minor == 0:
				flavor = "2000"
			}
		case major == 6:
			switch minor {
			case 3:
				flavor = "8.1"
			case 2:
				flavor = "8"
			case 1:
				flavor = "7"
			case 0:
				flavor = "Vista"
			}
		}
		os = fmt.Sprintf("%s %s: [Version %d.%d.%d]",
			strings.Title(runtime.GOOS), flavor,
			major, minor, build)
	}
	rt = fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return os, BOTVERSION + " (" + rt + ")"
}
