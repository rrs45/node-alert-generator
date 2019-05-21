#!/usr/bin/env bash

set -ex

godir=/tmp/go/src/github.com

mkdir -p $godir/box-autoremediation
export GOPATH=/tmp/go

mv cmd $godir/box-autoremediation/
mv pkg $godir/box-autoremediation/
cd $godir/box-autoremediation
mkdir bin
go get ./...

CGO_ENABLED=1 GOOS=linux go build -o bin/alert-generator cmd/alert_generator.go

mkdir -p /git-root/build
mv bin/alert-generator /git-root/build
