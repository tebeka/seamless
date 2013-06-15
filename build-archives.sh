#!/bin/bash

set -e

version=$(./seamless --version | awk '{print $2}')

filename=seamless-${version}-amd64.bz2
echo "building ${filename}"
go build
bzip2 -c seamless > ${filename}

filename=seamless-${version}-386.bz2
echo "building ${filename}"
GOARCH=386 go build
bzip2 -c seamless > ${filename}
