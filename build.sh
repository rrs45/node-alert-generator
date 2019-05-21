#!/usr/bin/env bash

set -ex

name=alert-generator
godir=/tmp/go/src
scratchdir=/tmp/scratch

mkdir -p $godir
export GOPATH=/tmp/go

cd $godir/$name
go get ./...

CGO_ENABLED=1 GOOS=linux go build -o bin/alert-generator cmd/alert_generator.go

mkdir -p /git-root/build
mv bin/$name /git-root/build
