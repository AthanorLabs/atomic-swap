#!/bin/bash

# install monero and run daemon and wallet RPC servers for alice and bob
./scripts/install-monero-linux.sh
echo "starting monerod..."
./monero-bin/monerod --detach --regtest --offline --fixed-difficulty=1 --rpc-bind-ip 127.0.0.1 --rpc-bind-port 18081
sleep 5

# install ganache-cli and run
GANACHE_EXEC="$(npm config get prefix)/bin/ganache-cli"
if [[ ! -x "${GANACHE_EXEC}" ]]; then
	echo "installing ganache-cli"
	npm install --location=global ganache-cli
fi
echo "starting ganache-cli"
export NODE_OPTIONS=--max_old_space_size=8192
"${GANACHE_EXEC}" --deterministic --accounts=20 &> ganache-cli.log &
GANACHE_CLI_PID=$!

# wait for servers to start
sleep 10

# run unit tests
echo "running unit tests..."
go test ./... -v -short -timeout=30m -covermode=atomic -coverprofile=coverage.out
OK=$?

# kill processes
kill "${GANACHE_CLI_PID}" || echo "ganache-cli was not running at end of test"
exit $OK
