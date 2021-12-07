.PHONY: lint test install build 
all: install

lint: 
	./scripts/install_lint.sh
	${GOPATH}/bin/golangci-lint run

test:
	./scripts/run-unit-tests.sh

test-integration:
	./scripts/run-integration-tests.sh

install:
	cd cmd/ && go install && cd ..

build:
	./scripts/build.sh