@echo off
cd %~dp0
cd ..
set GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,https://goproxy.io,direct
SET GOOS=windows
SET CGO_ENABLED=0
SET GOARCH=amd64
go build -v -o code-generator.exe ./cmd/code-generator/
go build -v -o v2.exe ./cmd/v2/
echo Build OK
