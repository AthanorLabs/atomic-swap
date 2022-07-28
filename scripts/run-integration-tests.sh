#!/bin/bash

# install monero and run daemon and wallet RPC servers for alice and bob
./scripts/install-monero-linux.sh
echo "starting monerod..."
./monero-bin/monerod --detach --regtest --offline --fixed-difficulty=1 --rpc-bind-ip 127.0.0.1 --rpc-bind-port 18081
sleep 5

echo "starting monero-wallet-rpc on port 18083..."
mkdir -p bob-test-keys
./monero-bin/monero-wallet-rpc --rpc-bind-ip 127.0.0.1 --rpc-bind-port 18083 --disable-rpc-login --wallet-dir ./bob-test-keys &> monero-wallet-cli-bob.log &
MONERO_WALLET_CLI_BOB_PID=$!

sleep 5
curl http://localhost:18083/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"create_wallet","params":{"filename":"test-wallet","password":"","language":"English"}}' -H 'Content-Type: application/json'
echo

echo "starting monero-wallet-rpc on port 18084..."
mkdir -p alice-test-keys
./monero-bin/monero-wallet-rpc --rpc-bind-ip 127.0.0.1 --rpc-bind-port 18084 --disable-rpc-login --wallet-dir ./alice-test-keys &> monero-wallet-cli-alice.log &
MONERO_WALLET_CLI_ALICE_PID=$!

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

# start alice and bob swapd instances
echo "starting alice, logs in ./tests/alice.log"
bash scripts/build.sh
./swapd --dev-xmrtaker --libp2p-key=./tests/alice.key &> ./tests/alice.log &
ALICE_SWAPD_PID=$!
sleep 3
echo "starting bob, logs in ./tests/bob.log"
./swapd --dev-xmrmaker --bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2 --wallet-file test-wallet --deploy &> ./tests/bob.log &
BOB_SWAPD_PID=$!
sleep 3 
echo "starting charlie, logs in ./tests/charlie.log"
./swapd --libp2p-port 9955 --rpc-port 5003 --ws-port 8083 --bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2 --deploy &> ./tests/charlie.log &
CHARLIE_SWAPD_PID=$!
sleep 3 

# run tests
echo "running integration tests..."
TESTS=integration go test ./tests -v
OK=$?

# kill processes
kill "${MONERO_WALLET_CLI_BOB_PID}" || echo "Bob's wallet CLI was not running at end of test"
kill "${MONERO_WALLET_CLI_ALICE_PID}" || echo "Alice's wallet CLI was not running at end of test"
kill "${GANACHE_PID}" || echo "ganache was not running at end of test"
kill "${ALICE_SWAPD_PID}" || echo "Alice's swapd was not running at end of test"
kill "${BOB_SWAPD_PID}" || echo "Bob's swapd was not running at end of test"
kill "${CHARLIE_SWAPD_PID}" || echo "Charlie's swapd was not running at end of test"
# rm -rf ./alice-test-keys
# rm -rf ./bob-test-keys
exit $OK
