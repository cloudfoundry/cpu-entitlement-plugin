.PHONY: test

help:
	@echo 'Help:'
	@echo '  build ........................ builds the cpu entitlement binary'
	@echo '  install ...................... builds and installs the cpu entitlement binary'
	@echo '  test ......................... run tests (such as they are)'
	@echo '  help ......................... show help menu'

build:
	go build

test:
	ginkgo -r --race

install: build
	cf uninstall-plugin CPUEntitlementPlugin || true
	cf install-plugin ./cpu-entitlement-plugin -f
