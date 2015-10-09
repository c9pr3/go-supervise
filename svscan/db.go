package main

import (
	"fmt"
	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
	"os/exec"
	"strings"
	"time"
)

type Service struct {
	Value string
	Cmd   *exec.Cmd
}

func deleteService(id string) {
	fmt.Printf("removing service %s\n", id)
	getClient().Delete(context.Background(), id, nil)
}

func createService(serviceName string, servicePath string) {
	fmt.Printf("createServiceInDb - %s\n", serviceName)
	getClient().Create(context.Background(), "supervise/"+serviceName, servicePath+"/"+serviceName)
}

func getServices() map[string]*Service {
	services := make(map[string]*Service)

	client, err := getClient().Get(context.Background(), "supervise", nil)
	if err != nil {
		panic(err)
	}
	values := client.Node.Nodes
	if values != nil {
		for _, key := range values {
			service := new(Service)
			service.Value = key.Value
			services[strings.Replace(key.Key, "/supervise/", "", 1)] = service
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
	cc.Set(context.Background(), "supervise", "", &client.SetOptions{Dir: true})

	return cc
}
