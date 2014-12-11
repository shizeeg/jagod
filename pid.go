package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// PID is a process id struct
type PID struct {
	ID       int
	FileName string
}

// Write PID struct (you can set filename p.FileName = "/var/run/pid")
// it uses "/run/executable_name/pid" by default
func (p *PID) Write() error {
	p.ID = os.Getpid()
	if len(p.FileName) < 2 {
		binary := filepath.Base(os.Args[0])
		p.FileName = fmt.Sprintf("/run/%s/pid", binary)
	}
	f, err := os.Create(p.FileName)
	defer f.Close()
	if err != nil {
		return err
	}
	fmt.Fprintf(f, "%d\n", p.ID)
	return err
}

// Read PID struct (p.FileName has to be set by the caller)
func (p *PID) Read() error {
	f, err := os.Open(p.FileName)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = fmt.Fscanf(f, "%d", &p.ID)
	return err
}
