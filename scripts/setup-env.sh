#!/bin/bash

# install monero and run daemon and wallet RPC servers for alice and bob
./scripts/install-monero-linux.sh
echo "starting monerod..."
./monero-bin/monerod --detach --regtest --offline --fixed-difficulty=1 --rpc-bind-ip 127.0.0.1 --rpc-bind-port 18081 &
sleep 5

echo "starting monero-wallet-rpc on port 18083..."
mkdir bob-test-keys
./monero-bin/monero-wallet-rpc --rpc-bind-ip 127.0.0.1 --rpc-bind-port 18083 --disable-rpc-login --wallet-dir ./bob-test-keys &> monero-wallet-cli-bob.log &
MONERO_WALLET_CLI_BOB_PID=$!

sleep 5
curl http://localhost:18083/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"create_wallet","params":{"filename":"test-wallet","password":"","language":"English"}}' -H 'Content-Type: application/json'

echo "starting monero-wallet-rpc on port 18084..."
mkdir alice-test-keys
./monero-bin/monero-wallet-rpc --rpc-bind-ip 127.0.0.1 --rpc-bind-port 18084 --disable-rpc-login --wallet-dir ./alice-test-keys &> monero-wallet-cli-alice.log &
MONERO_WALLET_CLI_ALICE_PID=$!

# install ganache and run 
echo "installing and starting ganache-cli..."
if ! command -v golangci-lint &> /dev/null; then
	npm i -g ganache-cli
fi
export NODE_OPTIONS=--max_old_space_size=8192
ganache-cli --deterministic &> ganache-cli.log &
GANACHE_CLI_PID=$!

# wait for servers to start
sleep 10
