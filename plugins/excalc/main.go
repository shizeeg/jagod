package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	lang     string
	filename = "ivy -e=true"
)

func init() {
	flag.StringVar(&lang, "lang", "en", "Language to output in.")
	flag.Parse()
}

func main() {
	if len(os.Getenv("EXCALC")) > 1 {
		filename = os.Getenv("EXCALC")
	}
	params := strings.Split(filename, " ")
	filename = params[0]
	plugin := exec.Command(filename, params[1:]...)
	plugin.Args = append(plugin.Args, flag.Args()...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	plugin.Stdout = &out
	plugin.Stderr = &stderr
	if err := plugin.Run(); err != nil {
		errmsg := fmt.Sprintf("%v - %s\n", err, strings.Trim(stderr.String(), " \t\r\n"))
		log.Println(errmsg)
		fmt.Print(errmsg)
		return
	}
	fmt.Println(strings.Trim(out.String(), " \t\r\n"))

}
