#!/bin/bash

PROJECT_ROOT="$(dirname "$(dirname "$(readlink -f "$0")")")"
cd "${PROJECT_ROOT}" || exit 1

./scripts/build.sh || exit 1

source "scripts/testlib.sh"
start-monerod-regtest
start-ganache
start-alice-wallet
start-bob-wallet

# wait for wallets to start
sleep 5

start-swapd() {
	local swapd_user="${1:?}"
	local swapd_flags=("${@:2}")
	echo "Starting ${swapd_user^}'s swapd, logs in ${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.log"
	./swapd "${swapd_flags[@]}" &>"${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.log" &
	echo "${!}" >"${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.pid"
}

stop-swapd() {
	local swapd_user="${1}"
	stop-program "${swapd_user}-swapd"
}

start-swapd alice \
	--dev-xmrtaker \
	--libp2p-key=./tests/alice.key

sleep 3 # Alice's swapd is a bootnode for Bob and Charlie's swapd

start-swapd bob \
	--dev-xmrmaker \
	--bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2 \
	--wallet-file test-wallet \
	--deploy

start-swapd charlie \
	--libp2p-port 9955 \
	--rpc-port 5003 \
	--ws-port 8083 \
	--bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2 \
	--deploy

sleep 3 # Time for Bob and Charlie's swapd to be fully up

# run tests
echo "running integration tests..."
TESTS=integration go test ./tests -v -count=1
OK=$?

stop-swapd alice
stop-swapd bob
stop-swapd charlie
stop-alice-wallet
stop-bob-wallet
stop-monerod-regtest
stop-ganache
remove-test-data-dir

exit $OK
