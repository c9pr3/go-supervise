#!/bin/zsh
go build -o svc  *.go && ./svc -t stop /Users/chris/test/services/srv1
