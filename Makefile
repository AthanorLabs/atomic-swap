GOPATH ?= $(shell go env GOPATH)
DLEQ_LIB=dleq/cgo-dleq/lib/libdleq.so

.PHONY: all
all: install

$(DLEQ_LIB):
	./scripts/install-rust.sh
	git submodule update --init --recursive
	cd dleq/cgo-dleq && make build

.PHONY: init
init: $(DLEQ_LIB)

.PHONY: lint-go
lint-go: init
	./scripts/install-lint.sh
	${GOPATH}/bin/golangci-lint run

.PHONY: lint-shell
lint-shell: init
	shellcheck --source-path=.:scripts scripts/*.sh

.PHONY: lint-solidity
lint-solidity: init
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
test: init
	./scripts/run-unit-tests.sh 2>&1 | tee test.log

.PHONY: test-integration
test-integration: init
	./scripts/run-integration-tests.sh 2>&1 | tee test-integration.log

# Install all the binaries into $HOME/go/bin (or alternative GOPATH bin directory)
.PHONY: install
install: init
	go install ./cmd/...

# Install swapd and swapcli into this directory (top of the project)
.PHONY: build
build: init
	./scripts/build.sh

# WARNING: this should not be used in production, as the DLEq prover has been stubbed out and now proves nothing.
.PHONY: build-go
build-go:
	GOBIN="$(CURDIR)/bin" go build -tags=fakedleq ./cmd/swapd
	GOBIN="$(CURDIR)/bin" go build -tags=fakedleq ./cmd/swapcli

# WARNING: this should not be used in production, as the DLEq prover has been stubbed out and now proves nothing.
.PHONY: build-go-darwin
build-go-darwin:
	GOOS=darwin GOARCH=arm64 $(MAKE) build-go

# Same as build, but also includes some lesser used binaries
.PHONY: build-all
build-all: init
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
.PHONY: clean-go
clean-go:
	rm -f bin/

.PHONY: clean
clean: clean-go
	rm -f $(DLEQ_LIB)
