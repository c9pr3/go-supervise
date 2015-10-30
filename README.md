## what
Proof of concept and playground for a [daemontools](http://cr.yp.to/daemontools.html) written in go.
The goal is to have a stable service starter which automagically restarts a service if it has been shut down and log (with logrotation) all output.

You may start up with an empty (or not even existing) service-dir and create the services after starting `go-supervise`. As soon as a "run" executable is found, the service is being started.

You may delete a service sub-directory (e.g. ~/services/srv1) while `go-supervise` is running. It will detect the deletion and shut down the service.

For now, `go-supervise` uses [etcd](https://github.com/coreos/etcd) to save known services, because later on
>- It should be possible to add a service in etcd database and let `go-supervise` starts it up on the correct server.
>- It should be possible to remove a service in etcd database and let `go-supervise` shut it down.

etcd is expected on **127.0.0.1:2379**, which is default. Later on `go-supervise` will have a configuration file.

## documentation
Not existent yet, Sorry :-)

## uses
Antigloss' [Logger](http://github.com/antigloss/go)
CoreOS [etcd client](http://github.com/coreos/etcd)
Vektra's [TAI64n](http://github.com/vektra/tai64n)

## usage
(Please make sure, GOPATH is set and etcd is running)
```
mkdir -p ~/services/srv1 && echo "echo \"starting up\" ; sleep 1000" > ~/services/srv1/run && chmod 755 ~/services/srv1/run
go get github.com/adar/go-supervise
# it will claim that there are no buildable go-files - ignore

# change service-path (edit config.json - ServiceConfig/path
cd $GOPATH/src/github.com/adar/go-supervise/config
vi config.json

cd ../svscan && ./run.sh & tail -f /var/log/syslog 
# where syslog might be /var/log/messages in CentOS
```
(If you see "Error while getting DB client client: etcd cluster is unavailable or
misconfigured" in your syslog, etcd is not running correctly.)

## known bugs
See [issues page](https://github.com/Adar/go-supervise/issues)

## License
MIT license, see LICENSE.txt for details.
