#!/bin/zsh
go build -o svscan  *.go && ./svscan -path ~/services
