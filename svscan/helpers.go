package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func readServiceDir(servicePath *string) []string {
	files, _ := ioutil.ReadDir("/" + *servicePath + "/")
	rval := make([]string, len(files))
	for _, f := range files {
		if f.Name() == "" {
			continue
		}
		if _, err := os.Stat("/" + *servicePath + "/" + f.Name() + "/run"); err != nil {
			continue
		}

		rval = append(rval, f.Name())
	}
	return rval
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

func removeServiceBefore(servicesInDir *[]string, key string) error {
	found := false
	for _, dir := range *servicesInDir {
		if dir == key {
			found = true
		}
	}

	if !found {
		fmt.Printf("Before: Did not find %s, %s\n", key)
		deleteService(key)
		err := fmt.Errorf("Before: Invalid service %s, %s", key)
		return err
	}

	return nil
}

func removeServiceIfNeeded(servicesInDir *[]string, key string, elem *Service, srvDone chan error) error {
	found := false
	for _, dir := range *servicesInDir {
		if dir == key {
			found = true
		}
	}

	if !found {
		fmt.Printf("Did not find %s, %s\n", key, *elem)
		err := fmt.Errorf("Invalid service %s, %s", key, *elem)
		return err
	}

	return nil
	/*
		done := make(chan bool)
		count := 0
			for id, _ := range *knownServices {
				knownServices := knownServices
				found := false

				for _, dir := range *servicesInDir {
					if dir == id {
						found = true
					}
				}

				if !found {
					fmt.Printf("Did not find %s\n", id)
					fmt.Printf("srv %s %s\n", (*knownServices)[id])
					deleteService(id)
					cmd := (*knownServices)[id].Cmd
					if cmd != nil {
						process := cmd.Process
						fmt.Printf("process: %s\n", process)
						if process != nil {
							err := process.Kill
							if err != nil {
							} else {
							}
						} else {
							fmt.Printf("process of %s is nil\n", id)
						}
					}
					delete(*knownServices, id)
				}
			}

			if count > 0 {
				fmt.Printf("i is %d", count)
				for i := 0; i <= count; i++ {
					fmt.Printf("i is %d", i)
					<-done
				}
			}
	*/
}

func createNewServicesIfNeeded(servicesInDir *[]string, knownServices *map[string]*Service, servicePath *string) {
	done := make(chan bool)
	count := 0

	fmt.Printf("known services %s\n", *knownServices)

	for _, dir := range *servicesInDir {
		dir := dir
		if dir == "" {
			continue
		}

		_, ok := (*knownServices)[dir]

		if dir == "" || ok == true {
			continue
		}

		go func() {
			fmt.Printf("creating new service %s, %s\n", dir, ok)
			createService(dir, *servicePath)
			srv := new(Service)
			srv.Value = dir
			(*knownServices)[dir] = srv
			count++
			done <- true
		}()

	}

	if count > 0 {
		fmt.Printf("i is %d", count)
		for i := 0; i <= count; i++ {
			fmt.Printf("i is %d", i)
			<-done
		}
	}
}

func getHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return hostName
}
