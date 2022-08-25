#!/bin/bash

PROJECT_ROOT="$(dirname "$(dirname "$(readlink -f "$0")")")"
cd "${PROJECT_ROOT}" || exit 1

source "scripts/testlib.sh"
start-monerod-regtest
start-ganache
start-alice-wallet
start-bob-wallet

# run unit tests
echo "running unit tests..."
rm -f coverage.out
go test ./... -v -short -timeout=30m -count=1 -covermode=atomic -coverprofile=coverage.out
OK=$?

if [[ -e coverage.out ]]; then
	go tool cover -html=coverage.out -o coverage.html
fi

stop-alice-wallet
stop-bob-wallet
stop-monerod-regtest
stop-ganache
remove-test-data-dir

exit $OK
