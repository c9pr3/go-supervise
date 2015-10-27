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
	VERSION              = "0.1"
	MAX_SERVICE_STARTUPS = 5
)

/**
 * Main
 * Loops through list of services in etcd database
 * and on defined path, resorts those and starts/restarts the services
 * which are still active.
 */
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

	/**
	 * Loop knownServices and services in directory
	 * If differ, decide which to remove or add
	 */
	for {
		knownServices := getServices()
		servicesInDir := readServiceDir(servicePath)
		createNewServicesIfNeeded(&servicesInDir, &knownServices, servicePath)

		for serviceName, elem := range knownServices {
			serviceName := serviceName
			elem := elem
			serviceDir := string(elem.Value)

			srvDone := make(chan error, 1)

			_, ok := runningServices[serviceName]
			if ok != true {
				go func() {
					err := removeServiceBefore(&servicesInDir, serviceName)
					if err == nil {
						LOGGER.Debug(fmt.Sprintf("%s not yet running\n", serviceName))
						time.Sleep(1 * time.Second)
						startService(srvDone, elem, runningServices, serviceName, serviceDir)
					}
				}()
			} else {
				err := removeServiceAfter(&servicesInDir, serviceName, runningServices[serviceName], srvDone)
				if err == nil {
					LOGGER.Debug(fmt.Sprintf("%s already running\n", serviceName))
				} else {
					delete(runningServices, serviceName)
				}
			}
		}

		time.Sleep(5 * time.Second)
	}

	LOGGER.Warning("exiting")
}

func writeLine(elem *Service, stdin io.WriteCloser, line string, serviceDir string) error {
	LOGGER.Debug(fmt.Sprintf("writing \"%s\" to stdin, %s\n", line, stdin))
	_, err := io.WriteString(stdin, line+"\n")
	return err
}

func startLogger(elem *Service, loggerDone chan error, serviceDir string, stdout io.ReadCloser) {
	if elem.Startups >= MAX_SERVICE_STARTUPS {
		if len(elem.LogBuffer) > 0 {
		}
		LOGGER.Crit(fmt.Sprintf("service %s has had too many startups in this session", serviceDir))
		return
	}

	elem.LogCmd = exec.Command("./../multilog/multilog", "-path", "/"+serviceDir)

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
		LOGGER.Crit(fmt.Sprintf("found unhandled log lines for %s, writing those first", serviceDir))
		for _, line := range elem.LogBuffer {
			err := writeLine(elem, stdin, line, serviceDir)
			if err != nil {
				LOGGER.Crit(fmt.Sprintf("Could not write buffered log for %s. error: %s", serviceDir, err))
				LOGGER.Crit(fmt.Sprintf("%s - %s", serviceDir, line))
				break
			}
		}
		elem.LogBuffer = nil
	}

	for stdOutBuff.Scan() {
		err := writeLine(elem, stdin, stdOutBuff.Text(), serviceDir)
		if err != nil {
			LOGGER.Crit(fmt.Sprintf("IO gone away for %s, %s", serviceDir, elem.LogCmd.Process))
			elem.LogBuffer = append(elem.LogBuffer, stdOutBuff.Text()+"\n")
			break
		}
	}
	loggerDone <- elem.LogCmd.Wait()
	select {
	case <-loggerDone:
		LOGGER.Warning(fmt.Sprintf("logger %s done without errors", serviceDir))
		if len(elem.LogBuffer) > 0 {
			startLogger(elem, loggerDone, serviceDir, stdout)
		}
		break
	case err := <-loggerDone:
		LOGGER.Warning(fmt.Sprintf("logger %s done with error = %v\n", serviceDir, err))
		startLogger(elem, loggerDone, serviceDir, stdout)
	}
}

func startService(srvDone chan error, elem *Service, runningServices map[string]*Service, serviceName string, serviceDir string) {
	if elem.Startups >= MAX_SERVICE_STARTUPS {
		LOGGER.Crit(fmt.Sprintf("service %s has had too many startups in this session", serviceName))
		return
	}
	loggerDone := make(chan error, 1)
	knownServices := getServices()
	if _, ok := knownServices[serviceName]; ok != true {
		return
	}
	LOGGER.Warning(fmt.Sprintf("Starting %s\n", serviceDir))

	elem.Cmd = exec.Command("/" + serviceDir + "/run")

	stdout, _ := elem.Cmd.StdoutPipe()

	elem.Startups += 1
	if err := elem.Cmd.Start(); err != nil {
		LOGGER.Crit(fmt.Sprintf("service %s not startable: %s", serviceName, err))
		return
	}
	LOGGER.Debug(fmt.Sprintf("Starting %s, %s\n", elem.Cmd.Process, elem.Value))

	go startLogger(elem, loggerDone, serviceDir, stdout)

	for elem.LogCmd == nil || elem.LogCmd.Process == nil {
		LOGGER.Warning(fmt.Sprintf("service %s, waiting for logger to come up", serviceName))
		time.Sleep(1 * time.Second)
	}

	runningServices[serviceName] = elem
	srvDone <- elem.Cmd.Wait()
	select {
	case err := <-srvDone:
		LOGGER.Warning(fmt.Sprintf("restarting service %s, %s", serviceName, err))
		if elem.LogCmd != nil && elem.LogCmd.Process != nil {
			loggerDone <- elem.LogCmd.Process.Kill()
		}
		time.Sleep(1 * time.Second)
		startService(srvDone, elem, runningServices, serviceName, serviceDir)
	}
}

func usage(code int) {
	fmt.Printf(
		`go- %s - (c) 2015 Christian Senkowski
			Usage: go-supervise -path /service-path/
		`, VERSION)
	os.Exit(code)
}
