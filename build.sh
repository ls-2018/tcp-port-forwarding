#!/bin/sh
export GOPROXY=https://proxy.golang.com.cn,direct
rm -rf tcpproxy.x64.*
ver=`date +%Y%m%d_%H%m%S`
sed 's/var ver = ""/var ver = "'${ver}'"/g' main.go > build.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o tcpproxy.x64.${ver}.linux build.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o tcpproxy.x64.${ver}.exe build.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o tcpproxy.x64.${ver}.mac build.go
rm -f build.go
# ./upx.exe -9 tcpproxy.x64.${ver}.linux
# ./upx.exe -9 tcpproxy.x64.${ver}.exe
# ./upx.exe -9 tcpproxy.x64.${ver}.mac
# CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o tcpproxy.x64.exe