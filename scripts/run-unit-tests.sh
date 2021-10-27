#!/bin/bash

# install monero and run daemon and wallet RPC server
bash ./scripts/install-monero-linux.sh
echo "starting monerod..."
# nohup ./monero-x86_64-linux-gnu-v0.17.2.3/monerod --regtest --offline --fixed-difficulty=1 --rpc-bind-port 18081 &> monerod.log &
# MONEROD_PID=$!
# sleep 15
echo "starting monero-wallet-rpc on port 18083..."
nohup ./monero-x86_64-linux-gnu-v0.17.2.3/monero-wallet-rpc --rpc-bind-port 18083 --disable-rpc-login --wallet-file test-wallet --password "" &> monero-wallet-cli-bob.log &
MONERO_WALLET_CLI_BOB_PID=$!

echo "starting monero-wallet-rpc on port 18084..."
nohup ./monero-x86_64-linux-gnu-v0.17.2.3/monero-wallet-rpc --rpc-bind-port 18084 --disable-rpc-login --wallet-dir . &> monero-wallet-cli-alice.log &
MONERO_WALLET_CLI_ALICE_PID=$!

# install ganache and run 
echo "installing and starting ganache-cli..."
if ! command -v golangci-lint &> /dev/null; then
	npm i -g ganache-cli
fi
nohup ganache-cli -d &> ganache-cli.log &
GANACHE_CLI_PID=$!

# wait for servers to start
sleep 10

# run unit tests
echo "running unit tests..."
go test ./...

# kill processes
kill $MONEROD_PID
kill $MONERO_WALLET_CLI_BOB_PID
kill $MONERO_WALLET_CLI_ALICE_PID
kill $GANACHE_CLI_PID