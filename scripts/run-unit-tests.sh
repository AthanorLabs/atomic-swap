#!/bin/bash

./scripts/setup-env.sh

# run unit tests
echo "running unit tests..."
rm -f coverage.out
go test ./... -v -short -timeout=30m -count=1 -covermode=atomic -coverprofile=coverage.out
OK=$?

if [[ -e coverage.out ]]; then
	go tool cover -html=coverage.out -o coverage.html
fi

# kill processes
kill "${GANACHE_PID}" || echo "ganache was not running at end of test"
exit $OK
