#!/bin/bash

PROJECT_ROOT="$(dirname "$(dirname "$(readlink -f "$0")")")"
cd "${PROJECT_ROOT}" || exit 1

./scripts/build.sh || exit 1

source "scripts/testlib.sh"
start-monerod-regtest
start-ganache
start-alice-wallet
start-bob-wallet
start-charlie-wallet

CHARLIE_ETH_KEY="${SWAP_TEST_DATA_DIR}/charlie-eth.key"
echo "87c546d6cb8ec705bea47e2ab40f42a768b1e5900686b0cecc68c0e8b74cd789" >"${CHARLIE_ETH_KEY}"

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
	--bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAAxG7eTEHr2uBVw3BDMxYsxyqfKvj3qqqpRGtTfuzTuH \
	--wallet-file test-wallet \
	--deploy

start-swapd charlie \
	--monero-endpoint "http://127.0.0.1:${CHARLIE_WALLET_PORT}/json_rpc" \
	--ethereum-privkey "${CHARLIE_ETH_KEY}" \
	--libp2p-port 9955 \
	--rpc-port 5003 \
	--bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAAxG7eTEHr2uBVw3BDMxYsxyqfKvj3qqqpRGtTfuzTuH \
	--deploy

sleep 3 # Time for Bob and Charlie's swapd to be fully up

# run tests
echo "running integration tests..."
TESTS=integration go test ./tests -v -count=1
OK="${?}"

# If we failed, make a copy of the log files that won't get deleted
if [[ "${OK}" -ne 0 ]]; then
	mkdir -p "${SWAP_TEST_DATA_DIR}/saved-logs"
	cp "${SWAP_TEST_DATA_DIR}/"*.log "${SWAP_TEST_DATA_DIR}/saved-logs/"
	echo "Logs saved to ${SWAP_TEST_DATA_DIR}/saved-logs/"
fi

stop-swapd alice
stop-swapd bob
stop-swapd charlie
stop-alice-wallet
stop-bob-wallet
stop-charlie-wallet
stop-monerod-regtest
stop-ganache
rm -f "${CHARLIE_ETH_KEY}"
if [[ "${OK}" -eq 0 ]]; then
	remove-test-data-dir
fi

exit "${OK}"
