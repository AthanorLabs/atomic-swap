GOPATH ?= $(shell go env GOPATH)

.PHONY: all
all: install

.PHONY: init
init:
	./scripts/install-rust.sh
	git submodule update --init --recursive
	cd dleq/cgo-dleq && make build

.PHONY: lint
lint:
	./scripts/install-lint.sh
	${GOPATH}/bin/golangci-lint run

.PHONY: test
test:
	./scripts/run-unit-tests.sh 2>&1 | tee test.log

.PHONY: test-integration
test-integration:
	./scripts/run-integration-tests.sh 2>&1 | tee test-integration.log

.PHONY: install
install: init
	cd cmd/ && go install && cd ..

.PHONY: build
build: init
	./scripts/build.sh

.PHONY: build-all
build-all: init
	ALL=true ./scripts/build.sh

# Go bindings for solidity contracts
.PHONY: bindings
bindings:
	./scripts/install-abigen.sh
	./scripts/generate-bindings.sh
	./ethereum/block/testdata/generate-bindings.sh

.PHONY: mock
mock:
	go generate -run mockgen ./...
