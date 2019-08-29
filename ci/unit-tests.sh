#!/usr/bin/env bash
set -euo pipefail

go version

cd src/code.cloudfoundry.org/cpu-entitlement-plugin

ginkgo -mod vendor -randomizeAllSpecs -randomizeSuites -race -skipPackage e2e,integration -r
