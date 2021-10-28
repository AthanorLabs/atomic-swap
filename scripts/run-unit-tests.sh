#!/bin/bash

# install monero and run daemon and wallet RPC servers for alice and bob
bash ./scripts/install-monero-linux.sh
echo "starting monerod..."
./monero-x86_64-linux-gnu-v0.17.2.3/monerod --detach --regtest --offline --fixed-difficulty=1 --rpc-bind-port 18081 &
MONEROD_PID=$!
sleep 5

echo "starting monero-wallet-rpc on port 18083..."
./monero-x86_64-linux-gnu-v0.17.2.3/monero-wallet-rpc --rpc-bind-port 18083 --disable-rpc-login --wallet-file test-wallet --password "" &> monero-wallet-cli-bob.log &
MONERO_WALLET_CLI_BOB_PID=$!

echo "starting monero-wallet-rpc on port 18084..."
mkdir test-keys
./monero-x86_64-linux-gnu-v0.17.2.3/monero-wallet-rpc --rpc-bind-port 18084 --disable-rpc-login --wallet-dir ./test-keys &> monero-wallet-cli-alice.log &
MONERO_WALLET_CLI_ALICE_PID=$!

# install ganache and run 
echo "installing and starting ganache-cli..."
if ! command -v golangci-lint &> /dev/null; then
	npm i -g ganache-cli
fi
ganache-cli -d &> ganache-cli.log &
GANACHE_CLI_PID=$!

# wait for servers to start
sleep 10

# run unit tests
echo "running unit tests..."
go test ./... -v -short
OK=$!

# kill processes
kill $MONEROD_PID
kill $MONERO_WALLET_CLI_BOB_PID
kill $MONERO_WALLET_CLI_ALICE_PID
kill $GANACHE_CLI_PID
rm -rf ./test-keys
exit $OK