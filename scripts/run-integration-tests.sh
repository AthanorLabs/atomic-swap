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

LOG_LEVEL=debug

BOOTNODE_RPC_PORT=4999
ALICE_RPC_PORT=5000
BOB_RPC_PORT=5001
CHARLIE_RPC_PORT=5002

start-bootnode() {
	local flags=(
		"--rpc-port=${BOOTNODE_RPC_PORT}"
		"--libp2p-ip=127.0.0.1"
		"--log-level=${LOG_LEVEL}"
	)
	local log_file="${SWAP_TEST_DATA_DIR}/bootnode.log"

	echo "Starting bootnode, logs in ${log_file}"
	./bin/bootnode "${flags[@]}" &>"${log_file}" &
	local pid="${!}"
	echo "${pid}" >"${SWAP_TEST_DATA_DIR}/bootnode.pid"

	if wait-rpc-started "bootnode" "${pid}" "${BOOTNODE_RPC_PORT}"; then
		return 0 # success
	fi

	echo "Failed to start bootnode"
	echo "=============== Failed logs  ==============="
	cat "${log_file}"
	echo "============================================"
	stop-daemons
	exit 1
}

stop-bootnode() {
	stop-program "bootnode"
}

start-swapd() {
	local swapd_user="${1:?}"
	local rpc_port="${2:?}"
	local swapd_flags=(
		"${@:3}"
		--env=dev
		"--rpc-port=${rpc_port}"
		"--log-level=${LOG_LEVEL}"
	)
	local log_file="${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.log"

	echo "Starting ${swapd_user^}'s swapd, logs in ${log_file}"
	./bin/swapd "${swapd_flags[@]}" &>"${log_file}" &
	local swapd_pid="${!}"
	echo "${swapd_pid}" >"${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.pid"

	if wait-rpc-started "${swapd_user}" "${swapd_pid}" "${rpc_port}"; then
		return 0 # success, bypass failure code below
	fi

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

# Waits for up to 60 seconds for the RPC port to be listening.
# If the passed PID exits, we stop waiting. This method only
# returns success(0)/failure(1) and will not exit the script.
wait-rpc-started() {
	local user="${1}"
	local pid="${2}"
	local port="${3}"

	# Wait up to 60 seconds for the daemon's port to be listening
	for i in {1..60}; do
		sleep 1

		# Test if pid is still alive, fail if it isn't
		if ! kill -0 "${pid}" 2>/dev/null; then
			return 1 # fail
		fi

		# Test if RPC port is listening, return success if it is
		if is-port-open "${port}"; then
			echo "${user^}'s instance is listening after ${i} seconds"
			return 0 # success
		fi
	done

	return 1 # fail
}

start-daemons() {
	start-monerod-regtest
	start-ganache
	start-bootnode

	local bootnode_addr
	bootnode_addr="$(./bin/swapcli addresses --swapd-port ${BOOTNODE_RPC_PORT} | grep '^1:' | sed 's/.* //')"
	start-swapd alice "${ALICE_RPC_PORT}" \
		--dev-xmrtaker \
		"--bootnodes=${bootnode_addr}" \
		"--data-dir=${SWAP_TEST_DATA_DIR}/alice" \
		--deploy

	local contract_addr
	contract_addr="$(./bin/swapcli version | grep '^swap creator address' | sed 's/.*: //')"
	if [[ -z "${contract_addr}" ]]; then
		echo "Failed to get Alice's deployed contract addresses"
		stop-daemons
		exit 1
	fi

	start-swapd bob "${BOB_RPC_PORT}" \
		--dev-xmrmaker \
		"--bootnodes=${bootnode_addr}" \
		"--data-dir=${SWAP_TEST_DATA_DIR}/bob" \
		--libp2p-port=9944 \
		"--contract-address=${contract_addr}"

	start-swapd charlie "${CHARLIE_RPC_PORT}" \
		"--bootnodes=${bootnode_addr}" \
		--data-dir "${SWAP_TEST_DATA_DIR}/charlie" \
		--libp2p-port=9955 \
		"--contract-address=${contract_addr}" \
		"--relayer"
}

stop-daemons() {
	stop-swapd charlie
	stop-swapd bob
	stop-swapd alice
	stop-bootnode
	stop-monerod-regtest
	stop-ganache
}

# run tests
echo "running integration tests..."
create-eth-keys
start-daemons
TESTS=integration go test ./tests -v -count=1 -timeout=30m
OK="${?}"
KEEP_TEST_DATA="${OK}" stop-daemons

# Cleanup test files if we succeeded
if [[ "${OK}" -eq 0 ]]; then
	rm -f "${CHARLIE_ETH_KEY}"
	rm -f "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}/monero-wallet-rpc.log
	rm -f "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}/{net,eth}.key
	rm -rf "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}/{wallet,libp2p-datastore,db}
	rmdir "${SWAP_TEST_DATA_DIR}/"{alice,bob,charlie}
	remove-test-data-dir
fi

exit "${OK}"
