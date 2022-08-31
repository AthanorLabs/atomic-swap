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

.PHONY: lint
lint: lint-go lint-shell

.PHONY: format-go
format-go:
	go fmt ./...

.PHONY: format-shell
format-shell:
	shfmt -w scripts/*.sh

.PHONY: format
format: format-go format-shell

.PHONY: test
test: init
	./scripts/run-unit-tests.sh 2>&1 | tee test.log

.PHONY: test-integration
test-integration: init
	./scripts/run-integration-tests.sh 2>&1 | tee test-integration.log

.PHONY: install
install: init
	cd cmd/ && go install ./...

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

.PHONY: clean
clean:
	rm -f $(DLEQ_LIB)
