package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
	"os/exec"
	//	"os/signal"
	//"syscall"
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

	servicePath = removeSlashes(servicePath)

	if LOGERR != nil {
		panic(LOGERR)
	}

	getClient()

	LOGGER.Warning(fmt.Sprintf("Scanning %s\n", *servicePath))

	runningServices := make(map[string]*Service)

	/*
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT) //, os.Interrupt)
		go func() {
			sig := <-sigs
			fmt.Printf(fmt.Sprintf("caught signal: %s\n", sig))
			/*
				for _, elem := range runningServices {
					if err := elem.Cmd.Process.Kill(); err != nil {
						panic(err)
					}
				}
		}()
	*/

	for {
		knownServices := getServices()
		servicesInDir := readServiceDir(servicePath)
		createNewServicesIfNeeded(&servicesInDir, &knownServices, servicePath)

		for key, elem := range knownServices {
			key := key
			elem := elem

			srvDone := make(chan error, 1)

			_, ok := runningServices[key]
			if ok != true {
				go func() {
					err := removeServiceBefore(&servicesInDir, key)
					if err == nil {
						LOGGER.Debug(fmt.Sprintf("%s not yet running\n", key))
						value := string(elem.Value)
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

	fmt.Println("exiting")
}

func startService(srvDone chan error, elem *Service, runningServices map[string]*Service, key string, value string) {
	//elem.Cmd = exec.Command("/Users/chris/private_git/go-supervise/supervise/supervise", "-path", "/"+value)
	knownServices := getServices()
	if _, ok := knownServices[key]; ok != true {
		return
	}
	LOGGER.Info(fmt.Sprintf("Starting %s\n", value))
	elem.Cmd = exec.Command("/" + value + "/run")
	//elem.Cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	go func(elem *Service) {
		elem.Cmd.Start()
		LOGGER.Debug(fmt.Sprintf("Starting %s, %s\n", elem.Cmd.Process, elem.Value))
		runningServices[key] = elem
		srvDone <- elem.Cmd.Wait()
	}(elem)
	select {
	case err := <-srvDone:
		if err != nil {
			LOGGER.Warning(fmt.Sprintf("process done with error = %v\n", err))
			startService(srvDone, elem, runningServices, key, value)
		} else {
			LOGGER.Warning(fmt.Sprintf("process %s interrupted\n", key))
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
