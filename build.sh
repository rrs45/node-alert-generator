#!/usr/bin/env bash

set -ex

sudo yum install -y systemd-libs systemd-devel

name=alert-generator
tarname=$(find . -type f -iname "*.tar.gz" -printf "%f\n")


godir=/tmp/go/src
scratchdir=/tmp/scratch

mkdir -p $godir
export GOPATH=/tmp/go

cd $godir/$name
go get ./...

CGO_ENABLED=1 GOOS=linux go build -o bin/alert-generator cmd/alert_generator.go

mkdir -p /git-root/build
mv bin/$name /git-root/build