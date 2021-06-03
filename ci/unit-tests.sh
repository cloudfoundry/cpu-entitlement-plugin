#!/usr/bin/env bash
set -euo pipefail

go version

cd src/code.cloudfoundry.org/cpu-entitlement-plugin

make test
