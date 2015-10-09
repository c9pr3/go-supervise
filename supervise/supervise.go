package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

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

	logger, err := syslog.New(syslog.LOG_WARNING, "supervise")
	if err != nil {
		panic(err)
	}

	sigs := make(chan os.Signal, 1)
	srvDone := make(chan error, 1)

	signal.Notify(sigs, syscall.SIGTERM)

	cmdName := *servicePath + "/run"
	logger.Warning(fmt.Sprintf("Starting up SUPER VISOR for %s", cmdName))
	cmd := exec.Command(cmdName)

	go func() {
		sig := <-sigs
		if sig == syscall.SIGTERM {
			logger.Warning(fmt.Sprintf("caught signal: ", sig))
			err := cmd.Process.Kill()
			if err != nil {
				logger.Warning(fmt.Sprintf("Caught error: %v", err))
			}
		}
	}()
	logger.Warning(fmt.Sprintf("Running %s\n", cmdName))
	go func() {
		cmd.Start()
		srvDone <- cmd.Wait()
	}()
	select {
	case err := <-srvDone:
		if err != nil {
			logger.Warning(fmt.Sprintf("process done with error = %v", err))
		}
	}

	logger.Warning(fmt.Sprintf("Shutting down %s", cmdName))
}

func usage(code int) {
	fmt.Printf(
		`go- %s - (c) 2015 Christian Senkowski
			Usage: supervise -path /service-path/
		`, VERSION)
	os.Exit(code)
}
