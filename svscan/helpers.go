// Go Supervise
// helpers.go - Helper methods
//
// (c) 2015, Christian Senkowski

package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

/**
 * Read service dir and return a list of valid services
 *
 * @param *string servicePath
 * @return []string
 */
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

/**
 * Remove a string's leading and trailing slashes
 *
 * @param *string
 */
func removeSlashes(s *string) {
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

/**
 * Prune old service from database before (re)starting it
 *
 * @param *[]string servicesInDirectory
 * @param string key the current service name
 * @return error
 */
func removeServiceBefore(servicesInDir *[]string, key string) error {
	found := false
	for _, dir := range *servicesInDir {
		if dir == key {
			found = true
		}
	}

	if !found {
		LOGGER.Debug(fmt.Sprintf("Before: Did not find %s, %s\n", key))
		new(DB).deleteService(key)
		err := fmt.Errorf("Before: Invalid service %s, %s", key)
		return err
	}

	return nil
}

/**
 * Remove and kill the service after it has been erased from directory
 *
 * @param *[]string services in directory
 * @param string current service name
 * @param *Service elem
 * @apram chan srvDone
 * @return error
 */
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
		new(DB).deleteService(key)
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
			new(DB).createService(dir, *servicePath)
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
