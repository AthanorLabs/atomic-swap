#!/usr/bin/env bash

PROJECT_ROOT="$(dirname "$(dirname "$(realpath "$0")")")"
cd "${PROJECT_ROOT}" || exit 1
./scripts/cleanup-test-processes.sh

ALL=true ./scripts/build.sh || exit 1

source "scripts/testlib.sh"
check-set-swap-test-data-dir

# The first 5 ganache keys are reserved for use by integration tests and dev swapd
# instances. For now, we are only using 4: Alice, Bob, Charlie and the relayer.
GANACHE_KEYS=(
	"4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d" # Key 0
	"6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1" # Key 1
	"6370fd033278c143179d81c5526140625662b8daa446c22ee2d73db3707e620c" # Key 2
	"646f1ce2fdad0e6deeeb5c7e8e5543bdde65e86029e2fd9fc169899c440a7913" # Key 3
	"add53f9a7e588d003326d1cbf9e4a43c061aadd9bc938c843a79e7b4fd2ad743" # Key 4 (placeholder)
)

KEY_USERS=(
	"alice"
	"bob"
	"charlie"
	"relayer"
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

start-relayer() {
	local relayer_flags=("${@}")
	local log_file="${SWAP_TEST_DATA_DIR}/relayer.log"
	echo "Starting relayer with logs in ${log_file}"
	./bin/relayer "${relayer_flags[@]}" &>"${log_file}" &
	local relayer_pid="${!}"
	echo "${relayer_pid}" >"${SWAP_TEST_DATA_DIR}/relayer.pid"
	sleep 1
	if ! kill -0 "${relayer_pid}" 2>/dev/null; then
		echo "Failed to start relayer"
		echo "=============== Failed logs  ==============="
		cat "${log_file}"
		echo "============================================"
		exit 1
	fi
}

stop-relayer() {
	stop-program "relayer"
}

start-swapd() {
	local swapd_user="${1:?}"
	local swapd_flags=("${@:2}")
	local log_file="${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.log"
	echo "Starting ${swapd_user^}'s swapd, logs in ${SWAP_TEST_DATA_DIR}/${swapd_user}-swapd.log"
	./bin/swapd "${swapd_flags[@]}" &>"${log_file}" &
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
	# Wait up to 60 seconds for Alice's swapd instance to start and deploy the swap contract
	#
	CONTRACT_ADDR_FILE="${SWAP_TEST_DATA_DIR}/alice/contract-addresses.json"
	for _ in {1..60}; do
		if [[ -f "${CONTRACT_ADDR_FILE}" ]]; then
			break
		fi
		sleep 1
	done
	SWAP_FACTORY_ADDR="$(jq -r .swapFactory "${CONTRACT_ADDR_FILE}")"
	FORWARDER_ADDR="$(jq -r .forwarder "${CONTRACT_ADDR_FILE}")"
	if [[ -z "${SWAP_FACTORY_ADDR}" ]] || [[ -z "${FORWARDER_ADDR}" ]]; then
		echo "Failed to get Alice's deployed contract addresses"
		stop-daemons
		exit 1
	fi

	start-relayer \
		"--log-level=${LOG_LEVEL}" \
		--data-dir="${SWAP_TEST_DATA_DIR}" \
		--rpc \
		--bootnodes="${ALICE_MULTIADDR}" \
		--forwarder-address="${FORWARDER_ADDR}"

	start-swapd bob \
		--dev-xmrmaker \
		"--log-level=${LOG_LEVEL}" \
		"--data-dir=${SWAP_TEST_DATA_DIR}/bob" \
		--libp2p-port=9944 \
		"--bootnodes=${ALICE_MULTIADDR}" \
		"--contract-address=${SWAP_FACTORY_ADDR}"

	start-swapd charlie \
		"--log-level=${LOG_LEVEL}" \
		--data-dir "${SWAP_TEST_DATA_DIR}/charlie" \
		--libp2p-port=9955 \
		--rpc-port 5002 \
		"--bootnodes=${ALICE_MULTIADDR}" \
		"--contract-address=${SWAP_FACTORY_ADDR}"

	# Give time for Bob and Charlie's swapd instances to fully start
	# and do peer discovery.
	sleep 30
}

stop-daemons() {
	stop-swapd charlie
	stop-swapd bob
	stop-relayer
	stop-swapd alice
	stop-monerod-regtest
	stop-ganache
}

# run tests
echo "running integration tests..."
create-eth-keys
start-daemons
TESTS=integration CONTRACT_ADDR=${SWAP_FACTORY_ADDR} go test ./tests -v -count=1 -timeout=30m
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
