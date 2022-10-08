#!/bin/bash

PROJECT_ROOT="$(dirname "$(dirname "$(readlink -f "$0")")")"
cd "${PROJECT_ROOT}" || exit 1

./scripts/build.sh || exit 1

source "scripts/testlib.sh"
check-set-swap-test-data-dir
mkdir -p "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}

# Charlie uses deterministic ganache key #49
CHARLIE_ETH_KEY="${SWAP_TEST_DATA_DIR}/charlie/eth.key"
echo "87c546d6cb8ec705bea47e2ab40f42a768b1e5900686b0cecc68c0e8b74cd789" >"${CHARLIE_ETH_KEY}"

# This is the local multiaddr created when using ./tests/alice-libp2p.key on the default libp2p port
ALICE_MULTIADDR=/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAAxG7eTEHr2uBVw3BDMxYsxyqfKvj3qqqpRGtTfuzTuH
ALICE_LIBP2PKEY=./tests/alice-libp2p.key
LOG_LEVEL=debug

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

start-daemons() {
	start-monerod-regtest
	start-ganache

	start-swapd alice \
		--dev-xmrtaker \
		"--log-level=${LOG_LEVEL}" \
		"--data-dir=${SWAP_TEST_DATA_DIR}/alice" \
		"--libp2p-key=${ALICE_LIBP2PKEY}" \
		--deploy

	#
	# Wait up to 10 seconds for Alice's swapd instance to start and deploy the swap contract
	#
	CONTRACT_ADDR_FILE="${SWAP_TEST_DATA_DIR}/alice/contract-address.json"
	for _ in {1..10}; do
		if [[ -f "${CONTRACT_ADDR_FILE}" ]]; then
			break
		fi
		sleep 1
	done
	if ! CONTRACT_ADDR="$(jq -r .ContractAddress "${SWAP_TEST_DATA_DIR}/alice/contract-address.json")"; then
		echo "Failed to get Alice's deployed contract address"
		stop-daemons
		exit 1
	fi

	start-swapd bob \
		--dev-xmrmaker \
		"--log-level=${LOG_LEVEL}" \
		"--data-dir=${SWAP_TEST_DATA_DIR}/bob" \
		--libp2p-port=9944 \
		"--bootnodes=${ALICE_MULTIADDR}" \
		"--contract-address=${CONTRACT_ADDR}"

	start-swapd charlie \
		"--log-level=${LOG_LEVEL}" \
		--data-dir "${SWAP_TEST_DATA_DIR}/charlie" \
		--libp2p-port=9955 \
		--rpc-port 5003 \
		"--bootnodes=${ALICE_MULTIADDR}" \
		"--contract-address=${CONTRACT_ADDR}"

	# Give time for Bob and Charlie's swapd instances to fully start
	sleep 10
}

stop-daemons() {
	stop-swapd charlie
	stop-swapd bob
	stop-swapd alice
	stop-monerod-regtest
	stop-ganache
}

# run tests
echo "running integration tests..."
start-daemons
TESTS=integration go test ./tests -v -count=1 -timeout=30m
OK="${?}"
KEEP_TEST_DATA="${OK}" stop-daemons

# Cleanup test files if we succeeded
if [[ "${OK}" -eq 0 ]]; then
	rm -f "${CHARLIE_ETH_KEY}"
	rm -f "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}/{contract-address.json,info-*.json,monero-wallet-rpc.log}
	rm -f "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}/{net,eth}.key
	rm -rf "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}/{wallet,libp2p-datastore}
	rmdir "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}
	remove-test-data-dir
fi

exit "${OK}"
