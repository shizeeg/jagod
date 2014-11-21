// +build !linux freebsd !windows

package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Uname is a wrapper around `uname` command
func Uname(param string) (os string, err error) {
	uname := exec.Command("uname", param)
	var out bytes.Buffer
	uname.Stdout = &out
	err = uname.Run()
	if err != nil {
		return
	}
	return out.String(), nil
}

// Version returns our version, os and Go runtime info.
func Version() (os string, rt string) {
	if osver, err := Uname("-rms"); err == nil {
		os = strings.Trim(osver, "\n\r")
	}

	rt = fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return os, BOTVERSION + " (" + rt + ")"
}
