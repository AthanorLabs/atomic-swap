#!/usr/bin/env bash

PROJECT_ROOT="$(dirname "$(dirname "$(realpath "$0")")")"
cd "${PROJECT_ROOT}" || exit 1
./scripts/cleanup-test-processes.sh

ALL=true ./scripts/build.sh || exit 1

source "scripts/testlib.sh"
check-set-swap-test-data-dir

# The first 5 ganache keys are reserved for use by integration tests and dev swapd
# instances. For now, we are only using 3: Alice, Bob, Charlie.
GANACHE_KEYS=(
	"4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d" # Key 0
	"6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1" # Key 1
	"6370fd033278c143179d81c5526140625662b8daa446c22ee2d73db3707e620c" # Key 2
	"646f1ce2fdad0e6deeeb5c7e8e5543bdde65e86029e2fd9fc169899c440a7913" # Key 3 (placeholder)
)

KEY_USERS=(
	"alice"
	"bob"
	"charlie"
)

create-eth-keys() {
	local i
	for i in "${!KEY_USERS[@]}"; do
		local key_file="${SWAP_TEST_DATA_DIR}/${KEY_USERS[${i}]}/eth.key"
		mkdir -p "$(dirname "${key_file}")"
		echo "${GANACHE_KEYS[${i}]}" >"${key_file}"
	done
}
# This is the local multiaddr created when using ./tests/alice-libp2p.key on the default libp2p port
ALICE_MULTIADDR=/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAAxG7eTEHr2uBVw3BDMxYsxyqfKvj3qqqpRGtTfuzTuH
ALICE_LIBP2PKEY=./tests/alice-libp2p.key
LOG_LEVEL=debug

ALICE_RPC_PORT=5000
BOB_RPC_PORT=5001
CHARLIE_RPC_PORT=5002

start-swapd() {
	local swapd_user="${1:?}"
	local rpc_port="${2:?}"
	local swapd_flags=("${@:3}" "--rpc-port=${rpc_port}")
	local log_file="${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.log"

	echo "Starting ${swapd_user^}'s swapd, logs in ${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.log"
	./bin/swapd "${swapd_flags[@]}" &>"${log_file}" &
	local swapd_pid="${!}"
	echo "${swapd_pid}" >"${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.pid"

	# Wait up to 60 seconds for the daemon's port to be listening
	for i in {1..60}; do
		sleep 1

		# Test if pid is still alive, leave loop if it is not
		if ! kill -0 "${swapd_pid}" 2>/dev/null; then
			break
		fi

		# Test if RPC port is listening, exit success if it is
		if is-port-open "${rpc_port}"; then
			echo "${swapd_user^}'s swapd instance is listening after ${i} seconds"
			return
		fi
	done

	echo "Failed to start ${swapd_user^}'s swapd"
	echo "=============== Failed logs  ==============="
	cat "${log_file}"
	echo "============================================"
	stop-daemons
	exit 1
}

stop-swapd() {
	local swapd_user="${1}"
	stop-program "${swapd_user}-swapd"
}

wait-rpc-started() {
	local swapd_user="${1}"
	local rpc_port="${1}"

}

start-daemons() {
	start-monerod-regtest
	start-ganache

	start-swapd alice "${ALICE_RPC_PORT}" \
		--dev-xmrtaker \
		"--log-level=${LOG_LEVEL}" \
		"--data-dir=${SWAP_TEST_DATA_DIR}/alice" \
		"--libp2p-key=${ALICE_LIBP2PKEY}" \
		--deploy

	CONTRACT_ADDR_FILE="${SWAP_TEST_DATA_DIR}/alice/contract-addresses.json"
	if [[ ! -f "${CONTRACT_ADDR_FILE}" ]]; then
		echo "Failed to get Alice's deployed contract address file"
		stop-daemons
		exit 1
	fi

	SWAP_CREATOR_ADDR="$(jq -r .swapCreatorAddr "${CONTRACT_ADDR_FILE}")"
	if [[ -z "${SWAP_CREATOR_ADDR}" ]] ; then
		echo "Failed to get Alice's deployed contract addresses"
		stop-daemons
		exit 1
	fi

	start-swapd bob "${BOB_RPC_PORT}" \
		--dev-xmrmaker \
		"--log-level=${LOG_LEVEL}" \
		"--data-dir=${SWAP_TEST_DATA_DIR}/bob" \
		--libp2p-port=9944 \
		"--bootnodes=${ALICE_MULTIADDR}" \
		"--contract-address=${SWAP_CREATOR_ADDR}"

	start-swapd charlie "${CHARLIE_RPC_PORT}" \
		"--env=dev" \
		"--log-level=${LOG_LEVEL}" \
		--data-dir "${SWAP_TEST_DATA_DIR}/charlie" \
		--libp2p-port=9955 \
		"--bootnodes=${ALICE_MULTIADDR}" \
		"--contract-address=${SWAP_CREATOR_ADDR}" \
		"--relayer"
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
create-eth-keys
start-daemons
TESTS=integration CONTRACT_ADDR=${SWAP_CREATOR_ADDR} go test ./tests -v -count=1 -timeout=30m
OK="${?}"
KEEP_TEST_DATA="${OK}" stop-daemons

# Cleanup test files if we succeeded
if [[ "${OK}" -eq 0 ]]; then
	rm -f "${CHARLIE_ETH_KEY}"
	rm -f "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}/{contract-addresses.json,monero-wallet-rpc.log}
	rm -f "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}/{net,eth}.key
	rm -rf "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}/{wallet,libp2p-datastore,db}
	rmdir "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}
	remove-test-data-dir
fi

exit "${OK}"
