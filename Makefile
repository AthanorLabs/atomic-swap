GOPATH ?= $(shell go env GOPATH)

.PHONY: all
all: install

.PHONY: init
init:
	./scripts/install-rust.sh
	git submodule update --init --recursive
	cd dleq/cgo-dleq && make build

.PHONY: lint
lint: init
	./scripts/install-lint.sh
	${GOPATH}/bin/golangci-lint run

.PHONY: test
test:
	./scripts/run-unit-tests.sh 2>&1 | tee test.log

.PHONY: test-integration
test-integration:
	./scripts/run-integration-tests.sh 2>&1 | tee test-integration.log

.PHONY: install
install:
	cd cmd/ && go install && cd ..

.PHONY: build
build:
	./scripts/build.sh

.PHONY: build-all
build-all:
	ALL=true ./scripts/build.sh

.PHONY: mock
mock:
	go generate -run mockgen ./...
