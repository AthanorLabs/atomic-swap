#!/bin/bash

# install monero and run daemon and wallet RPC servers for alice and bob
bash ./scripts/install-monero-linux.sh
echo "starting monerod..."
./monero-x86_64-linux-gnu-v0.17.3.0/monerod --detach --regtest --offline --fixed-difficulty=1 --rpc-bind-port 18081 &
sleep 5

echo "starting monero-wallet-rpc on port 18083..."
mkdir bob-test-keys
./monero-x86_64-linux-gnu-v0.17.3.0/monero-wallet-rpc --rpc-bind-port 18083 --disable-rpc-login --wallet-dir ./bob-test-keys &> monero-wallet-cli-bob.log &
MONERO_WALLET_CLI_BOB_PID=$!

sleep 5
curl http://localhost:18083/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"create_wallet","params":{"filename":"test-wallet","password":"","language":"English"}}' -H 'Content-Type: application/json'

echo "starting monero-wallet-rpc on port 18084..."
mkdir alice-test-keys
./monero-x86_64-linux-gnu-v0.17.3.0/monero-wallet-rpc --rpc-bind-port 18084 --disable-rpc-login --wallet-dir ./alice-test-keys &> monero-wallet-cli-alice.log &
MONERO_WALLET_CLI_ALICE_PID=$!

# install ganache and run 
echo "installing and starting ganache-cli..."
if ! command -v golangci-lint &> /dev/null; then
	npm i -g ganache-cli
fi
export NODE_OPTIONS=--max_old_space_size=8192
ganache-cli -d &> ganache-cli.log &
GANACHE_CLI_PID=$!

# wait for servers to start
sleep 10

# start alice and bob swapd instances
echo "starting alice, logs in ./tests/alice.log"
bash scripts/build.sh
TESTS=integration ./swapd --dev-alice --libp2p-key=./tests/alice.key &> ./tests/alice.log &
ALICE_PID=$!
sleep 3
echo "starting bob, logs in ./tests/bob.log"
./swapd --dev-bob --bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2 --wallet-file test-wallet &> ./tests/bob.log &
BOB_PID=$!
sleep 3 
echo "starting charlie, logs in ./tests/charlie.log"
./swapd --libp2p-port 9955 --rpc-port 5003 --ws-port 8083 --bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2 &> ./tests/charlie.log &
CHARLIE_PID=$!
sleep 3 

# run tests
echo "running integration tests..."
TESTS=integration go test ./tests -v
OK=$?

# kill processes
kill $MONERO_WALLET_CLI_BOB_PID
kill $MONERO_WALLET_CLI_ALICE_PID
kill $GANACHE_CLI_PID
kill $ALICE_PID
kill $BOB_PID
kill $CHARLIE_PID
# rm -rf ./alice-test-keys
# rm -rf ./bob-test-keys
exit $OK