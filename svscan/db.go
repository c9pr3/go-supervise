// Go Supervise
// db.go - etcd DB connection code
//
// (c) 2015, Christian Senkowski

package main

import (
	"fmt"
	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
	"os"
	"os/exec"
	"strings"
	"time"
)

type DB struct {
}

/**
 * One Service
 */
type Service struct {
	Value     string
	Cmd       *exec.Cmd
	LogCmd    *exec.Cmd
	LogBuffer []string
	LogFile   *os.File
	Startups  int
}

/**
 * Delete a service from DB
 *
 * @param string serviceName
 */
func (d *DB) deleteService(serviceName string) {
	LOGGER.Debug(fmt.Sprintf("removing service %s\n", serviceName))
	d.getClient().Delete(context.Background(), "supervise/"+getHostName()+"/"+serviceName, nil)
}

/**
 * Create a service in DB
 *
 * @param string serviceName
 * @param string servicePath
 */
func (d *DB) createService(serviceName string, servicePath string) {
	LOGGER.Debug(fmt.Sprintf("createServiceInDb - %s\n", serviceName))
	d.getClient().Create(context.Background(), "supervise/"+getHostName()+"/"+serviceName, servicePath+"/"+serviceName)
}

/**
 * Get all services from DB
 *
 * @return map[string ]*Service
 */
func (d *DB) getServices() map[string]*Service {
	services := make(map[string]*Service)

	client, err := d.getClient().Get(context.Background(), "supervise/"+getHostName(), nil)
	if err != nil {
		panic(err)
	}
	values := client.Node.Nodes
	if values != nil {
		for _, key := range values {
			service := new(Service)
			service.Value = key.Value
			services[strings.Replace(key.Key, "/supervise/"+getHostName()+"/", "", 1)] = service
		}
	}

	return services
}

/**
 * Get client
 *
 * @return client.KeysAPI
 */
func (d *DB) getClient() client.KeysAPI {
	cfg := client.Config{
		Endpoints:               []string{"http://127.0.0.1:2379"},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		panic(err)
	}
	cc := client.NewKeysAPI(c)
	cc.Set(context.Background(), "supervise/"+getHostName(), "", &client.SetOptions{Dir: true})

	return cc
}
