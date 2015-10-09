package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	//"os/signal"
	//"syscall"
	"time"
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

	servicePath = removeSlashes(servicePath)
	getClient()

	fmt.Printf("Scanning %s\n", *servicePath)

	runningServices := make(map[string]*Service)

	for {
		knownServices := getServices()
		servicesInDir := readServiceDir(servicePath)
		createNewServicesIfNeeded(&servicesInDir, knownServices, servicePath)

		for key, elem := range knownServices {
			key := key
			elem := elem
			srvDone := make(chan error, 1)

			_, ok := runningServices[key]
			if ok != true {
				fmt.Printf("%s not yet running\n", key)
				value := string(elem.Value)
				elem.Cmd = exec.Command("/Users/chris/private_git/go-supervise/supervise/supervise", "-path", "/"+value)
				go func(elem *Service) {
					elem.Cmd.Start()
					fmt.Printf("Starting %s\n", elem.Cmd.Process)
					runningServices[key] = elem
					srvDone <- elem.Cmd.Wait()
				}(elem)
			} else {
				// todo
			}
		}

		time.Sleep(5 * time.Second)
	}

	fmt.Println("exiting")
}

func usage(code int) {
	fmt.Printf(
		`go- %s - (c) 2015 Christian Senkowski
			Usage: go-supervise /service-path/
		`, VERSION)
	os.Exit(code)
}
