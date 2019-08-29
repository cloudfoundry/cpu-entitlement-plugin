.PHONY: test

help:
	@echo 'Help:'
	@echo '  build ........................ build the cpu entitlement binary'
	@echo '  install ...................... build and install the cpu entitlement binary'
	@echo '  test ......................... run tests (such as they are)'
	@echo '  help ......................... show help menu'

build:
	go build -mod vendor -o cpu-entitlement-plugin  ./cmd/cpu-entitlement
	go build -mod vendor -o cpu-overentitlement-instances-plugin  ./cmd/cpu-overentitlement-instances

test:
	ginkgo -r -p -mod vendor -skipPackage e2e,integration -keepGoing

install: build
	cf uninstall-plugin CPUEntitlementPlugin || true
	cf install-plugin ./cpu-entitlement-plugin -f
	cf uninstall-plugin CPUEntitlementAdminPlugin || true
	cf install-plugin ./cpu-overentitlement-instances-plugin -f
