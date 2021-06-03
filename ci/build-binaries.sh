#!/usr/bin/env bash
set -euo pipefail

IGNORE_PROTOBUF_ERROR="-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore"

cd cpu-entitlement-plugin

go-build() {
  local file_extension
  file_extension="$1"

  binary_path="${PWD}/../plugin-binaries/cpu-${plug}-plugin-$(cat ../version/number)"
  go build -ldflags "${IGNORE_PROTOBUF_ERROR}" -mod vendor -o "${binary_path}.${file_extension}" ./cmd/cpu-${plug}
}

for plug in entitlement overentitlement-instances; do
  GOOS=linux   GOARCH=amd64 go-build linux64
  GOOS=linux   GOARCH=386   go-build linux32
  GOOS=windows GOARCH=amd64 go-build win64
  GOOS=windows GOARCH=386   go-build win32
  GOOS=darwin  GOARCH=amd64 go-build osx
done
