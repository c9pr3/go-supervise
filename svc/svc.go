// Go Supervise
// svc.go - External Service controller
//
// (c) 2015, Christian Senkowski

package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
)

var LOGGER, LOGERR = syslog.New(syslog.LOG_WARNING, "svc")

const (
	VERSION = "0.1"
)

func main() {

	workType := flag.String("t", "", "type")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Printf("not enough arguments - %s\n", *workType)
		usage(1)
	}
}

func usage(code int) {
	fmt.Printf(
		`go- %s - (c) 2015 Christian Senkowski
			Usage: svc -t type /service-path/srv
			Where type can be one of the following
			stop
			start
			restart
			terminate
		`, VERSION)
	os.Exit(code)
}
