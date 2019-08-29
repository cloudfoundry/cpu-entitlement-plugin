#!/usr/bin/env bash
set -e

pushd cpu-entitlement-plugin
ginkgo -mod vendor -randomizeAllSpecs -randomizeSuites -race -keepGoing integration
popd
