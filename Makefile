.PHONY: lint test install build build-dleq mock
all: build-dleq install

GOPATH ?= $(shell go env GOPATH)

lint: 
	./scripts/install-lint.sh
	${GOPATH}/bin/golangci-lint run

test:
	./scripts/run-unit-tests.sh

test-integration:
	./scripts/run-integration-tests.sh

install:
	cd cmd/ && go install && cd ..

build:
	./scripts/build.sh

build-all:
	ALL=true ./scripts/build.sh
	
build-dleq:
	./scripts/install-rust.sh && cd farcaster-dleq && cargo build --release && cd ..

mock:
	go generate -run mockgen ./...
