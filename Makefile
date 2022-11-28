GOPATH ?= $(shell go env GOPATH)

.PHONY: all
all: install

.PHONY: lint-go
lint-go: 
	./scripts/install-lint.sh
	${GOPATH}/bin/golangci-lint run

.PHONY: lint-shell
lint-shell:
	shellcheck --source-path=.:scripts scripts/*.sh

.PHONY: lint-solidity
lint-solidity: 
	"$$(npm config get prefix)/bin/solhint" $$(find ethereum -name '*.sol')

.PHONY: lint
lint: lint-go lint-shell lint-solidity

.PHONY: format-go
format-go:
	go fmt ./...

.PHONY: format-shell
format-shell:
	shfmt -w scripts/*.sh

.PHONY: format-solidity
format-solidity:
	"$$(npm config get prefix)/bin/prettier" --print-width 100 --write $$(find ethereum -name '*.sol')

.PHONY: format
format: format-go format-shell format-solidity

.PHONY: test
test: 
	./scripts/run-unit-tests.sh 2>&1 | tee test.log

.PHONY: test-integration
test-integration: 
	./scripts/run-integration-tests.sh 2>&1 | tee test-integration.log

# Install all the binaries into $HOME/go/bin (or alternative GOPATH bin directory)
.PHONY: install
install: 
	go install ./cmd/...

# Install swapd and swapcli into this directory (top of the project)
.PHONY: build
build: 
	./scripts/build.sh

# Test macos/arm build from linux. Use "make build" if compiling on macos.
.PHONY: build-darwin
build-darwin:
	mkdir -p bin/
	GOOS=darwin GOARCH=arm64 go build  -o ./bin ./cmd/...

# Same as build, but also includes some lesser used binaries
.PHONY: build-all
build-all: 
	ALL=true $(MAKE) build

# Go bindings for solidity contracts
.PHONY: bindings
bindings:
	./scripts/install-abigen.sh
	./scripts/generate-bindings.sh
	./ethereum/block/testdata/generate-bindings.sh

.PHONY: mock
mock:
	go generate -run mockgen ./...

# Deletes all executables matching the directory names in cmd/
.PHONY: clean
clean:
	rm -r bin/