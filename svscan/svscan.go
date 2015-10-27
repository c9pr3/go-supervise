// Go Supervise
// svscan.go - Service controller code
//
// (c) 2015, Christian Senkowski

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log/syslog"
	"os"
	"os/exec"
	"time"
)

var LOGGER, LOGERR = syslog.New(syslog.LOG_WARNING, "svscan")

const (
	VERSION = "0.1"
)

func main() {

	servicePath := flag.String("path", "", "service path")
	flag.Parse()

	if len(flag.Args()) != 0 {
		fmt.Printf("not enough arguments - %s\n", *servicePath)
		usage(1)
	}

	removeSlashes(servicePath)

	if LOGERR != nil {
		panic(LOGERR)
	}

	getClient()

	LOGGER.Warning(fmt.Sprintf("Scanning %s\n", *servicePath))

	if _, err := os.Stat("/" + *servicePath); err != nil {
		if crErr := os.Mkdir("/"+*servicePath, 0755); crErr != nil {
			LOGGER.Crit(fmt.Sprintf("Scanning %s failed - directory does not exist and is not creatable\n", *servicePath))
			fmt.Printf("Scanning %s failed - directory does not exist and is not creatable\n", *servicePath)
			usage(1)
		}
	}

	runningServices := make(map[string]*Service)

	for {
		knownServices := getServices()
		servicesInDir := readServiceDir(servicePath)
		createNewServicesIfNeeded(&servicesInDir, &knownServices, servicePath)

		for key, elem := range knownServices {
			key := key
			elem := elem
			value := string(elem.Value)

			srvDone := make(chan error, 1)

			_, ok := runningServices[key]
			if ok != true {
				go func() {
					err := removeServiceBefore(&servicesInDir, key)
					if err == nil {
						LOGGER.Debug(fmt.Sprintf("%s not yet running\n", key))
						startService(srvDone, elem, runningServices, key, value)
					}
				}()
			} else {
				err := removeServiceAfter(&servicesInDir, key, runningServices[key], srvDone)
				if err == nil {
					LOGGER.Debug(fmt.Sprintf("%s already running\n", key))
				} else {
					delete(runningServices, key)
				}
			}
		}

		time.Sleep(5 * time.Second)
	}

	LOGGER.Warning("exiting")
}

func writeLine(elem *Service, stdin io.WriteCloser, line string, value string) error {
	LOGGER.Debug(fmt.Sprintf("writing \"%s\" to stdin, %s\n", line, stdin))
	_, err := io.WriteString(stdin, line+"\n")
	return err
}

func startLogger(elem *Service, loggerDone chan error, value string, stdout io.ReadCloser) {

	elem.LogCmd = exec.Command("./../multilog/multilog", "-path", "/"+value)

	//@TODO rewrite multilog so that it can take stderr and stdout separately
	stdOutBuff := bufio.NewScanner(stdout)
	stdin, err := elem.LogCmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	defer stdin.Close()
	err = elem.LogCmd.Start()
	if err != nil {
		panic(err)
	}

	if len(elem.LogBuffer) > 0 {
		LOGGER.Crit(fmt.Sprintf("found unhandled log lines %s, writing those first", elem.LogBuffer))
		for _, line := range elem.LogBuffer {
			err := writeLine(elem, stdin, line, value)
			if err != nil {
				LOGGER.Crit(fmt.Sprintf("Could not write buffered log for %s. error: %s", value, err))
				LOGGER.Crit(fmt.Sprintf("%s - %s", value, line))
				break
			}
		}
		elem.LogBuffer = nil
	}

	for stdOutBuff.Scan() {
		err := writeLine(elem, stdin, stdOutBuff.Text(), value)
		if err != nil {
			LOGGER.Crit(fmt.Sprintf("IO gone away for %s, %s", value, elem.LogCmd.Process))
			elem.LogBuffer = append(elem.LogBuffer, stdOutBuff.Text()+"\n")
			break
		}
	}
	loggerDone <- elem.LogCmd.Wait()
	select {
	case err := <-loggerDone:
		if err != nil {
			LOGGER.Warning(fmt.Sprintf("logger %s done with error = %v\n", value, err))
			startLogger(elem, loggerDone, value, stdout)
		} else {
			LOGGER.Warning(fmt.Sprintf("logger %s interrupted\n", value))
			startLogger(elem, loggerDone, value, stdout)
		}
	}
}

func startService(srvDone chan error, elem *Service, runningServices map[string]*Service, key string, value string) {
	loggerDone := make(chan error, 1)
	knownServices := getServices()
	if _, ok := knownServices[key]; ok != true {
		return
	}
	LOGGER.Warning(fmt.Sprintf("Starting %s\n", value))

	elem.Cmd = exec.Command("/" + value + "/run")

	stdout, _ := elem.Cmd.StdoutPipe()

	if err := elem.Cmd.Start(); err != nil {
		LOGGER.Crit(fmt.Sprintf("service %s not startable: %s", key, err))
		return
	}
	LOGGER.Debug(fmt.Sprintf("Starting %s, %s\n", elem.Cmd.Process, elem.Value))

	go startLogger(elem, loggerDone, value, stdout)

	runningServices[key] = elem
	srvDone <- elem.Cmd.Wait()
	select {
	case err := <-srvDone:
		if err != nil {
			LOGGER.Warning(fmt.Sprintf("process %s done with error = %v", key, err))
			LOGGER.Warning(fmt.Sprintf("restarting service %s", key))
			if elem.LogCmd != nil && elem.LogCmd.Process != nil {
				LOGGER.Warning(fmt.Sprintf("restarting service-logger 2 %s", key))
				loggerDone <- elem.LogCmd.Process.Kill()
			}
			LOGGER.Warning(fmt.Sprintf("restarting service now %s", key))
			startService(srvDone, elem, runningServices, key, value)
		} else {
			LOGGER.Warning(fmt.Sprintf("restarting service %s", key))
			if elem.LogCmd != nil && elem.LogCmd.Process != nil {
				loggerDone <- elem.LogCmd.Process.Kill()
			}
			startService(srvDone, elem, runningServices, key, value)
		}
	}
}

func usage(code int) {
	fmt.Printf(
		`go- %s - (c) 2015 Christian Senkowski
			Usage: go-supervise -path /service-path/
		`, VERSION)
	os.Exit(code)
}
