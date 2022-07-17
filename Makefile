GOPATH ?= $(shell go env GOPATH)

.PHONY: all
all: build-dleq install

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
install:
	cd cmd/ && go install && cd ..

.PHONY: build
build:
	./scripts/build.sh

.PHONY: build-all
build-all:
	ALL=true ./scripts/build.sh

.PHONY: build-dleq
build-dleq:
	./scripts/install-rust.sh && cd farcaster-dleq && cargo build --release && cd ..

.PHONY: mock
mock:
	go generate -run mockgen ./...
