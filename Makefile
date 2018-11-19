.PHONY: test

build:
	go build

test:
	ginkgo -r --race

install:
	rm ~/.cf/plugins/CPUEntitlementPlugin || true
	cf uninstall-plugin CPUEntitlementPlugin || true
	cf install-plugin ./cpu-entitlement-plugin -f
