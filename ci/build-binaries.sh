#!/usr/bin/env bash
set -euo pipefail

BINARY_PATH="${PWD}/plugin-binaries/cpu-entitlement-plugin-$(cat version/number)"

cd cpu-entitlement-plugin
GOOS=linux   GOARCH=amd64 go build -mod vendor -o ${BINARY_PATH}.linux64
GOOS=linux   GOARCH=386   go build -mod vendor -o ${BINARY_PATH}.linux32
GOOS=windows GOARCH=amd64 go build -mod vendor -o ${BINARY_PATH}.win64
GOOS=windows GOARCH=386   go build -mod vendor -o ${BINARY_PATH}.win32
GOOS=darwin  GOARCH=amd64 go build -mod vendor -o ${BINARY_PATH}.osx
