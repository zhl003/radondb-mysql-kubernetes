#! /bin/sh
export GOPROXY=https://goproxy.cn
dlv --headless --log --listen :9009 --api-version 2 --check-go-version=false --accept-multiclient debug cmd/manager/main.go