## what
Proof of concept and playground for a [daemontools](cr.yp.to/daemontools) written in go.
The goal is to have a stable service starter which automagically restarts a service if it has been shut down and log (with logrotation) all output.

You may start up with an empty (or not even existing) service-dir and create the services after starting `go-supervise`. As soon as a "run" executable is found, the service be started up.

You may delete a service sub-directory (e.g. ~/services/srv1) while `go-supervise` is running. It will detect the deletion and shut down the service.

For now, `go-supervise` uses [etcd](https://github.com/coreos/etcd) to save known services, because later on
>- It should be possible to add a service in etcd database and let `go-supervise` start it up on the correct server.
>- It should be possible to remove a service in etcd database and let `go-supervise` shut it down.

etcd is expected to run on **127.0.0.1:2379**, which is default. Later on `go-supervise` will have a configuration file.

## documentation
Not existent yet, Sorry :-)

## uses
Antigloss' [Logger](http://github.com/antigloss/go),
CoreOS [etcd client](http://github.com/coreos/etcd)

## usage
```
mkdir ~/services/srv1 && echo "exec sleep 1000" > ~/services/srv1/run
git clone https://github.com/Adar/go-supervise
cd go-supervise/svscan
./svscan -path ~/services/
```
