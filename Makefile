.PHONY: test

help:
	@echo 'Help:'
	@echo '  build ........................ build the cpu entitlement binary'
	@echo '  install ...................... build and install the cpu entitlement binary'
	@echo '  test ......................... run tests (such as they are)'
	@echo '  help ......................... show help menu'

IGNORE_PROTOBUF_ERROR = "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore"
build:
	go build -ldflags $(IGNORE_PROTOBUF_ERROR) -mod vendor -o cpu-entitlement-plugin  ./cmd/cpu-entitlement
	go build -ldflags $(IGNORE_PROTOBUF_ERROR) -mod vendor -o cpu-overentitlement-instances-plugin  ./cmd/cpu-overentitlement-instances

test:
	ginkgo -ldflags $(IGNORE_PROTOBUF_ERROR) -r -p -mod vendor -skipPackage e2e,integration -keepGoing -randomizeAllSpecs -race

install: build
	cf uninstall-plugin CPUEntitlementPlugin || true
	cf install-plugin ./cpu-entitlement-plugin -f
	cf uninstall-plugin CPUEntitlementAdminPlugin || true
	cf install-plugin ./cpu-overentitlement-instances-plugin -f

e2e-test:
	ginkgo -ldflags $(IGNORE_PROTOBUF_ERROR) -mod vendor -randomizeAllSpecs -randomizeSuites -race -keepGoing e2e

integration-test:
	ginkgo -ldflags $(IGNORE_PROTOBUF_ERROR) -mod vendor -randomizeAllSpecs -randomizeSuites -race -keepGoing integration
