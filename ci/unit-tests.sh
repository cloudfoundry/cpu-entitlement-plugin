#!/usr/bin/env bash
set -euo pipefail

export GOPATH=$PWD

go version

ginkgo -r --race
