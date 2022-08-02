.PHONY: init lint test install build build-dleq mock
all: install

GOPATH ?= $(shell go env GOPATH)

init:
	./scripts/install-rust.sh
	git submodule update --init --recursive
	cd dleq/cgo-dleq && make build

lint: 
	./scripts/install-lint.sh
	${GOPATH}/bin/golangci-lint run

test:
	./scripts/run-unit-tests.sh

test-integration:
	./scripts/run-integration-tests.sh

install: init 
	cd cmd/ && go install && cd ..

build: init
	./scripts/build.sh

build-all: init
	ALL=true ./scripts/build.sh

mock:
	go generate -run mockgen ./...
