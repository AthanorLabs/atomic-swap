#!/usr/bin/env bash

PROJECT_ROOT="$(dirname "$(dirname "$(realpath "$0")")")"
cd "${PROJECT_ROOT}" || exit 1
./scripts/cleanup-test-processes.sh

source "scripts/testlib.sh"
start-monerod-regtest
start-ganache

# run unit tests
echo "running unit tests..."
rm -f coverage.txt
go test -coverpkg=./... -v -short -timeout=30m -count=1 -covermode=atomic -coverprofile=coverage.txt ./...
OK=$?

if [[ -e coverage.txt ]]; then
	go tool cover -html=coverage.txt -o coverage.html
fi

stop-monerod-regtest
stop-ganache
remove-test-data-dir

exit "${OK}"
