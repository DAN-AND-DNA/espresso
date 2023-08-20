@echo off
cd %~dp0
cd ..
set GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,https://goproxy.io,direct
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -v -o code-generator ./cmd/code-generator/
go build -v -o config-generator ./cmd/config-generator/
go build -v -o e2etest ./cmd/e2etest/
go build -v -o v2 ./cmd/v2/
echo Build OK
