// Go Supervise
// multilogger.go - Logging for svscan
//
// (c) 2015, Christian Senkowski
//
// @TODO
// - Make configurable

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
	VERSION = "0.3"
)

/**
 * Receives lines and log(rotates) them away
 */
func main() {

	servicePath := flag.String("path", "", "service path")
	flag.Parse()
	if LOGERR != nil {
		panic(LOGERR)
	}
	LOGGER.Debug("multilog starting up " + *servicePath)

	if len(flag.Args()) != 0 {
		LOGGER.Crit("not enough arguments")
		fmt.Printf("not enough arguments - %s\n", *servicePath)
		os.Exit(1)
	}

	removeSlashes(servicePath)

	logger.Init("/"+*servicePath+"/log",
		50,    // maximum logfiles allowed under the specified log directory
		20,    // number of logfiles to delete when number of logfiles exceeds the configured limit
		1,     // maximum size of a logfile in MB
		false) // whether logs with Trace level are written down

	info, _ := os.Stdin.Stat()

	if (info.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
		// no input
	} else {
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

			LOGGER.Debug(fmt.Sprintf("input received %s", input))
			logger.Info("%s %s", timeStamp, input)
		}
	}

	LOGGER.Debug("multilog shutting down\n")
}

/**
 * Remove slashes
 *
 * @param *string str with possible leading and trailing slashes
 */
func removeSlashes(s *string) {
	if len(*s) <= 1 {
		return
	}
	for {
		if os.IsPathSeparator(((*s)[len(*s)-1 : len(*s)])[0]) {
			*s = (*s)[0 : len(*s)-1]
		} else if os.IsPathSeparator(((*s)[0:1])[0]) {
			*s = (*s)[1:len(*s)]
		} else {
			break
		}
	}
}
