#!/usr/bin/env bash
set -euo pipefail

export GOPATH=$PWD

PLUGIN_PATH=$GOPATH/src/code.cloudfoundry.org/cpu-entitlement-plugin
BINARY_PATH="${PWD}/plugin-binaries/$(basename $PLUGIN_PATH)-$(cat version/number)"

cd $PLUGIN_PATH
GOOS=linux   GOARCH=amd64 go build -o ${BINARY_PATH}.linux64
GOOS=linux   GOARCH=386   go build -o ${BINARY_PATH}.linux32
GOOS=windows GOARCH=amd64 go build -o ${BINARY_PATH}.win64
GOOS=windows GOARCH=386   go build -o ${BINARY_PATH}.win32
GOOS=darwin  GOARCH=amd64 go build -o ${BINARY_PATH}.osx
