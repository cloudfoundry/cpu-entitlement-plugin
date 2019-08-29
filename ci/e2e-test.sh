#!/usr/bin/env bash
set -e

pushd cpu-entitlement-plugin
ginkgo -mod vendor e2e
popd
