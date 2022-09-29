#!/bin/bash

PROJECT_ROOT="$(dirname "$(dirname "$(readlink -f "$0")")")"
cd "${PROJECT_ROOT}" || exit 1

# Integration tests still need Bob's wallet on a pre-established port, so XMR Maker
# funds can be mined.
BOB_WALLET_PORT=18083

./scripts/build.sh || exit 1

source "scripts/testlib.sh"
start-monerod-regtest
start-ganache

mkdir -p "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}

# Charlie uses deterministic ganache key #49
CHARLIE_ETH_KEY="${SWAP_TEST_DATA_DIR}/charlie/eth.key"
echo "87c546d6cb8ec705bea47e2ab40f42a768b1e5900686b0cecc68c0e8b74cd789" >"${CHARLIE_ETH_KEY}"

# This is the local multiaddr created when using ./tests/alice-libp2p.key on the default libp2p port
ALICE_MULTIADDR=/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2
ALICE_LIBP2PKEY=./tests/alice-libp2p.key

start-swapd() {
	local swapd_user="${1:?}"
	local swapd_flags=("${@:2}")
	local log_file="${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.log"
	echo "Starting ${swapd_user^}'s swapd, logs in ${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.log"
	./swapd "${swapd_flags[@]}" &>"${log_file}" &
	local swapd_pid="${!}"
	echo "${swapd_pid}" >"${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.pid"
	sleep 1
	if ! kill -0 "${swapd_pid}" 2>/dev/null; then
		echo "Failed to start ${swapd_user^}'s swapd"
		echo "=============== Failed logs  ==============="
		cat "${log_file}"
		echo "============================================"
		exit 1
	fi
}

stop-swapd() {
	local swapd_user="${1}"
	stop-program "${swapd_user}-swapd"
}

start-swapd alice \
	--dev-xmrtaker \
	"--data-dir=${SWAP_TEST_DATA_DIR}/alice" \
	"--libp2p-key=${ALICE_LIBP2PKEY}" \
	--deploy

#
# Wait up to 5 seconds for Alice's swapd instance to start and deploy the swap contract
#
CONTRACT_ADDR_FILE="${SWAP_TEST_DATA_DIR}/alice/contract-address.json"
for _ in {1..5}; do
	if [[ -f "${CONTRACT_ADDR_FILE}" ]]; then
		break
	fi
	sleep 1
done
if ! CONTRACT_ADDR="$(cat "${SWAP_TEST_DATA_DIR}/alice/contract-address.json")"; then
	echo "Failed to get Alice's deployed contract address"
	exit 1
fi

start-swapd bob \
	--dev-xmrmaker \
	"--wallet-port=${BOB_WALLET_PORT}" \
	"--data-dir=${SWAP_TEST_DATA_DIR}/bob" \
	--libp2p-port=9944 \
	"--bootnodes=${ALICE_MULTIADDR}" \
	"--contract-address=${CONTRACT_ADDR}"

start-swapd charlie \
	--data-dir "${SWAP_TEST_DATA_DIR}/charlie" \
	--ethereum-privkey "${CHARLIE_ETH_KEY}" \
	--libp2p-port=9955 \
	--rpc-port 5003 \
	"--bootnodes=${ALICE_MULTIADDR}" \
	"--contract-address=${CONTRACT_ADDR}"

# Give time for Bob and Charlie's swapd instances to fully start
sleep 5

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
stop-monerod-regtest
stop-ganache
rm -f "${CHARLIE_ETH_KEY}"
rm -f "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}/{contract-address.json,info-*.json,monero-wallet-rpc.log}
rm -rf "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}/{wallet,libp2p-datastore}
rmdir "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}

if [[ "${OK}" -eq 0 ]]; then
	remove-test-data-dir
fi

exit "${OK}"
