// Go Supervise
// svscan.go - Service controller code
//
// (c) 2015, Christian Senkowski

package main

import (
	"bufio"
	"fmt"
	"github.com/adar/go-supervise/config"
	"io"
	"log/syslog"
	"os"
	"os/exec"
	"sync"
	"time"
)

var LOGGER, LOGERR = syslog.New(syslog.LOG_WARNING, "svscan")
var CONFIG, CONFERR = config.ReadConfig()

const (
	VERSION              = "0.3"
	MAX_SERVICE_STARTUPS = 5
)

type ServiceHandler struct {
	mutex   *sync.Mutex
	service *Service
}

func main() {
	if CONFERR != nil {
		fmt.Printf("Error while reading config ", CONFERR)
	} else {
		start()
	}
}

/**
* Start
* Loops through list of services in etcd database
* and on defined path, resorts those and starts/restarts the services
* which are still active.
 */
func start() {

	if LOGERR != nil {
		panic(LOGERR)
	}

	servicePath := &CONFIG.ServiceConfig.Path
	removeSlashes(servicePath)

	db := new(DB)
	db.getClient()

	LOGGER.Warning(fmt.Sprintf("Scanning %s\n", *servicePath))

	if _, err := os.Stat("/" + *servicePath); err != nil {
		if crErr := os.Mkdir("/"+*servicePath, 0755); crErr != nil {
			LOGGER.Crit(fmt.Sprintf("Scanning %s failed - directory does not exist and is not creatable\n", *servicePath))
			fmt.Printf("Scanning %s failed - directory does not exist and is not creatable\n", *servicePath)
			usage(1)
		}
	}

	runningServices := make(map[string]*Service)

	// Loop knownServices and services in directory
	// If differ, decide which to remove or add
	for {
		servicesInDir := readServiceDir(servicePath)
		db.createNewServicesIfNeeded(&servicesInDir, servicePath)
		knownServices := db.getServices()

		for serviceName, service := range knownServices {
			serviceName := serviceName
			service := service

			srvDone := make(chan error, 1)

			_, ok := runningServices[serviceName]
			if ok != true {
				go func() {
					err := removeServiceBefore(&servicesInDir, serviceName)
					if err == nil {
						LOGGER.Debug(fmt.Sprintf("%s not yet running\n", serviceName))
						time.Sleep(1 * time.Second)
						sv := new(ServiceHandler)
						sv.mutex = &sync.Mutex{}
						sv.service = service
						sv.startService(srvDone, runningServices, serviceName)
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

func (s *ServiceHandler) writeLine(stdin io.WriteCloser, line string) error {
	LOGGER.Debug(fmt.Sprintf("writing \"%s\" to stdin, %s\n", line, stdin))
	_, err := io.WriteString(stdin, line+"\n")
	return err
}

func (s *ServiceHandler) startLogger(loggerDone chan error, stdout io.ReadCloser) {
	if s.service.Startups >= MAX_SERVICE_STARTUPS {
		LOGGER.Crit(fmt.Sprintf("service %s has had too many startups in this session", s.service.Value))
		if len(s.service.LogBuffer) > 0 {
			LOGGER.Crit(fmt.Sprintf("%s - %s", s.service.Value, s.service.LogBuffer))
			s.service.LogBuffer = nil
		}
		return
	}

	s.mutex.Lock()
	s.service.LogCmd = exec.Command("./../multilog/multilog", "-path", "/"+s.service.Value)

	//@TODO rewrite multilog so that it can take stderr and stdout separately
	stdOutBuff := bufio.NewScanner(stdout)
	stdin, err := s.service.LogCmd.StdinPipe()
	if err != nil {
		LOGGER.Crit(fmt.Sprintf("Error while catching stdinpipe of %s, %s", s.service.Value, err))
		s.mutex.Unlock()
		return
	}
	defer stdin.Close()
	err = s.service.LogCmd.Start()
	if err != nil {
		LOGGER.Crit(fmt.Sprintf("Error while starting logger %s", err))
		s.mutex.Unlock()
		return
	}
	s.mutex.Unlock()

	if len(s.service.LogBuffer) > 0 {
		LOGGER.Crit(fmt.Sprintf("found unhandled log lines for %s, writing those first", s.service.Value))
		for _, line := range s.service.LogBuffer {
			err := s.writeLine(stdin, line)
			if err != nil {
				LOGGER.Crit(fmt.Sprintf("Could not write buffered log for %s. error: %s", s.service.Value, err))
				LOGGER.Crit(fmt.Sprintf("%s - %s", s.service.Value, line))
				break
			}
		}
		s.service.LogBuffer = nil
	}

	for stdOutBuff.Scan() {
		err := s.writeLine(stdin, stdOutBuff.Text())
		if err != nil {
			LOGGER.Crit(fmt.Sprintf("IO gone away for %s, %s", s.service.Value, s.service.LogCmd.Process))
			s.service.LogBuffer = append(s.service.LogBuffer, stdOutBuff.Text()+"\n")
			break
		}
	}
	select {
	case <-loggerDone:
		LOGGER.Warning(fmt.Sprintf("logger %s done without errors", s.service.Value))
		if len(s.service.LogBuffer) > 0 {
			s.startLogger(loggerDone, stdout)
		}
		break
	case err := <-loggerDone:
		LOGGER.Warning(fmt.Sprintf("logger %s done with error = %v\n", s.service.Value, err))
		s.startLogger(loggerDone, stdout)
	}
}

func (s *ServiceHandler) startService(srvDone chan error, runningServices map[string]*Service, serviceName string) {
	if s.service.Startups >= CONFIG.ServiceConfig.MaxFailedStartups {
		LOGGER.Crit(fmt.Sprintf("service %s has had too many startups in this session", serviceName))
		return
	}
	loggerDone := make(chan error, 1)
	s.mutex.Lock()
	knownServices := new(DB).getServices()
	s.mutex.Unlock()
	if _, ok := knownServices[serviceName]; ok != true {
		return
	}
	LOGGER.Warning(fmt.Sprintf("Starting %s\n", s.service.Value))

	s.service.Cmd = exec.Command("/" + s.service.Value + "/run")

	stdout, _ := s.service.Cmd.StdoutPipe()

	s.service.Startups += 1
	if err := s.service.Cmd.Start(); err != nil {
		LOGGER.Crit(fmt.Sprintf("service %s not startable: %s", serviceName, err))
		return
	}
	LOGGER.Debug(fmt.Sprintf("Starting %s, %s\n", s.service.Cmd.Process, s.service.Value))

	go s.startLogger(loggerDone, stdout)

	runningServices[serviceName] = s.service
	srvDone <- s.service.Cmd.Wait()
	select {
	case err := <-srvDone:
		LOGGER.Warning(fmt.Sprintf("restarting service %s, %s", serviceName, err))
		if s.service.LogCmd != nil && s.service.LogCmd.Process != nil {
			loggerDone <- s.service.LogCmd.Process.Kill()
		}
		time.Sleep(time.Duration(CONFIG.ServiceConfig.TimeWaitBetweenStartups) * time.Second)
		s.startService(srvDone, runningServices, serviceName)
	}
}

func usage(code int) {
	fmt.Printf(
		`go- %s - (c) 2015 Christian Senkowski
			Usage: go-supervise -path /service-path/
		`, VERSION)
	os.Exit(code)
}
