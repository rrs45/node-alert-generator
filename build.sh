#!/usr/bin/env bash

set -ex

godir=/tmp/go/src/github.com

mkdir -p $godir/node-alert-generator

export GOPATH=/tmp/go

mv cmd $godir/node-alert-generator/
mv pkg $godir/node-alert-generator/
mv go.mod $godir/node-alert-generator/

cd $godir/node-alert-generator
mkdir bin
export GO111MODULE=on
go get ./...
go test -v ./pkg/...
CGO_ENABLED=1 GOOS=linux go build -o bin/node-alert-generator -ldflags '-w' cmd/node_alert_generator.go

mkdir -p /git-root/build
mv bin/node-alert-generator /git-root/build
