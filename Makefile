.PHONY: test

help:
	@echo 'Help:'
	@echo '  build ........................ build the cpu entitlement binary'
	@echo '  install ...................... build and install the cpu entitlement binary'
	@echo '  test ......................... run tests (such as they are)'
	@echo '  help ......................... show help menu'

build:
	go build -mod vendor

test:
	ginkgo -r --race

install: build
	cf uninstall-plugin CPUEntitlementPlugin || true
	cf install-plugin ./cpu-entitlement-plugin -f
