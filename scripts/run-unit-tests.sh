#!/bin/bash

# install monero and run daemon and wallet RPC servers for alice and bob
./scripts/install-monero-linux.sh
echo "starting monerod..."
./monero-bin/monerod --detach --regtest --offline --fixed-difficulty=1 --rpc-bind-ip 127.0.0.1 --rpc-bind-port 18081
sleep 5

# install ganache and run
GANACHE_EXEC="$(npm config get prefix)/bin/ganache"
if [[ ! -x "${GANACHE_EXEC}" ]]; then
	echo "installing ganache"
	npm install --location=global ganache
fi
echo "starting ganache"
export NODE_OPTIONS=--max_old_space_size=8192
"${GANACHE_EXEC}" --deterministic --accounts=50 --miner.blockTime=1 &> ganache.log &
GANACHE_PID=$!

# wait for servers to start
sleep 10

# run unit tests
echo "running unit tests..."
rm -f coverage.out
go test ./... -v -short -timeout=30m -covermode=atomic -coverprofile=coverage.out
OK=$?

if [[ -e coverage.out ]]; then
	go tool cover -html=coverage.out -o coverage.html
fi

# kill processes
kill "${GANACHE_PID}" || echo "ganache was not running at end of test"
exit $OK
