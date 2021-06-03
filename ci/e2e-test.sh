#!/usr/bin/env bash
set -e

pushd cpu-entitlement-plugin
make e2e-test
popd
