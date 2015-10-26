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
		if os.IsPathSeparator((p[len(p)-1 : len(p)])[0]) {
			p = p[0 : len(p)-1]
		} else if os.IsPathSeparator((p[0:1])[0]) {
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
		LOGGER.Debug(fmt.Sprintf("Before: Did not find %s, %s\n", key))
		deleteService(key)
		err := fmt.Errorf("Before: Invalid service %s, %s", key)
		return err
	}

	return nil
}

func removeServiceAfter(servicesInDir *[]string, key string, elem *Service, srvDone chan error) error {
	found := false
	for _, dir := range *servicesInDir {
		if dir == key {
			found = true
		}
	}

	if !found {
		LOGGER.Warning(fmt.Sprintf("service %s gone, killing %s\n", key, (*elem).Cmd.Process))
		err := fmt.Errorf("service %s gone, %s", key, (*elem).Cmd.Process)
		deleteService(key)
		srvDone <- elem.Cmd.Process.Kill()
		return err
	}

	return nil
}

func createNewServicesIfNeeded(servicesInDir *[]string, knownServices *map[string]*Service, servicePath *string) {
	done := make(chan bool)
	count := 0

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
			LOGGER.Info(fmt.Sprintf("creating new service %s, %s\n", dir, ok))
			createService(dir, *servicePath)
			srv := new(Service)
			srv.Value = dir
			(*knownServices)[dir] = srv
			count++
			done <- true
		}()
	}

	if count > 0 {
		for i := 0; i <= count; i++ {
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
