#!/bin/bash

make
version=$(./seamless --version | awk '{print $2}')
bzip2 -c seamless > seamless-${version}-amd64.bz2
GOARCH=386 make
bzip2 -c seamless > seamless-${version}-386.bz2
