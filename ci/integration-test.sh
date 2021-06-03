#!/usr/bin/env bash
set -e

pushd cpu-entitlement-plugin
make integration-test
popd
