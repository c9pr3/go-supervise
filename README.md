## what
Proof of concept and playground for a daemontools / supervise written in go

## documentation
Not existent yet, Sorry :-)

## usage
```
mkdir ~/services/srv1 && echo "exec sleep 1000" > ~/services/srv1/run
git clone https://github.com/Adar/go-supervise
cd go-supervise/svscan
./svscan -path ~/services/
```
