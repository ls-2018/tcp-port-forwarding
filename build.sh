#!/bin/sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o tcpproxy.x64.linux
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o tcpproxy.x64.exe
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o tcpproxy.x64.mac
./upx.exe -9 tcpproxy.x64.exe
# CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o tcpproxy.x64.exe