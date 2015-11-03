#!/bin/zsh
cd ../multilog && go build -o multilog *.go
cd ../svc && go build -o svc *.go
cd ../svscan && go build -o svscan *.go
./svscan
