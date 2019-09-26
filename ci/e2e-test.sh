#!/usr/bin/env bash
set -e

pushd cpu-entitlement-plugin
echo -e "$ROUTER_CA_CERT" > ca-cert.pem
ginkgo -mod vendor -randomizeAllSpecs -randomizeSuites -race -keepGoing e2e
popd
