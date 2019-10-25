#!/usr/bin/env bash
set -euo pipefail

cd cpu-entitlement-plugin

for plug in entitlement overentitlement-instances; do
  BINARY_PATH="${PWD}/plugin-binaries/cpu-${plug}-plugin-$(cat ../version/number)"

  GOOS=linux   GOARCH=amd64 go build -mod vendor -o ${BINARY_PATH}.linux64 ./cmd/cpu-${plug}
  GOOS=linux   GOARCH=386   go build -mod vendor -o ${BINARY_PATH}.linux32 ./cmd/cpu-${plug}
  GOOS=windows GOARCH=amd64 go build -mod vendor -o ${BINARY_PATH}.win64   ./cmd/cpu-${plug}
  GOOS=windows GOARCH=386   go build -mod vendor -o ${BINARY_PATH}.win32   ./cmd/cpu-${plug}
  GOOS=darwin  GOARCH=amd64 go build -mod vendor -o ${BINARY_PATH}.osx     ./cmd/cpu-${plug}
done
