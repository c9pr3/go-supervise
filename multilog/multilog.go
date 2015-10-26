// Go Supervise
// multilogger.go - Logging for svscan
//
// (c) 2015, Christian Senkowski

package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/antigloss/go/logger"
	"github.com/vektra/tai64n"
	"io"
	"log/syslog"
	"os"
)

var LOGGER, LOGERR = syslog.New(syslog.LOG_WARNING, "svscan")

const (
	VERSION = "0.1"
)

func main() {

	servicePath := flag.String("path", "", "service path")
	flag.Parse()
	if LOGERR != nil {
		panic(LOGERR)
	}
	LOGGER.Warning("multilog starting up " + *servicePath)

	if len(flag.Args()) != 0 {
		LOGGER.Crit("not enough arguments")
		fmt.Printf("not enough arguments - %s\n", *servicePath)
		os.Exit(1)
	}

	servicePath = removeSlashes(servicePath)

	logger.Init("/"+*servicePath+"/log",
		50,    // maximum logfiles allowed under the specified log directory
		20,    // number of logfiles to delete when number of logfiles exceeds the configured limit
		1,     // maximum size of a logfile in MB
		false) // whether logs with Trace level are written down

	for {
		info, _ := os.Stdin.Stat()

		if (info.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
			// no input
		} else if info.Size() > 0 {
			reader := bufio.NewReader(os.Stdin)
			for {
				input, err := reader.ReadString('\n')
				if err != nil && err == io.EOF {
					LOGGER.Crit(fmt.Sprintf("input read error %s", err))
					break
				}
				// remove trailing newline
				input = input[0 : len(input)-1]

				timeStamp := tai64n.Now().Label()

				LOGGER.Crit(fmt.Sprintf("input received %s", input))
				logger.Info("%s %s", timeStamp, input)
			}
		}
	}

	LOGGER.Crit("multilog shutting down\n")
}

func removeSlashes(s *string) *string {
	p := *s
	for {
		if p[len(p)-1:len(p)] == "/" {
			p = p[0 : len(p)-1]
		} else if p[0:1] == "/" {
			p = p[1:len(p)]
		} else {
			break
		}
	}
	return &p
}
