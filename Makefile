.PHONY: lint test install build 
all: install

lint: 
	./scripts/install_lint.sh
	${GOPATH}/bin/golangci-lint run

test:
	./scripts/run-unit-tests.sh

install:
	cd cmd/ && go install && cd ..

build:
	./scripts/build.sh