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

type Service struct {
	Value   string
	Cmd     *exec.Cmd
	LogFile *os.File
}

func deleteService(serviceName string) {
	LOGGER.Debug(fmt.Sprintf("removing service %s\n", serviceName))
	getClient().Delete(context.Background(), "supervise/"+getHostName()+"/"+serviceName, nil)
}

func createService(serviceName string, servicePath string) {
	LOGGER.Debug(fmt.Sprintf("createServiceInDb - %s\n", serviceName))
	getClient().Create(context.Background(), "supervise/"+getHostName()+"/"+serviceName, servicePath+"/"+serviceName)
}

func getServices() map[string]*Service {
	services := make(map[string]*Service)

	client, err := getClient().Get(context.Background(), "supervise/"+getHostName(), nil)
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

func getClient() client.KeysAPI {
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
