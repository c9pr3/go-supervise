#!/bin/zsh
go build -o svc  *.go && ./svc -t stop ~/services/srv1
