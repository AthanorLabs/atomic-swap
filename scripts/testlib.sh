#!/bin/bash
# This script is designed to be sourced by other test scripts, or by the shell
# on the command line if you are using bash. Here is an example:
#
# $ source scripts/testlib.sh
# $ start-ganache
# $ start-monerod-regtest
#
# Later, from the same shell that started the services, you can run
# $ stop-ganache
# $ stop-monerod-regtest
#
# If the shell that started the services is no longer running, no
# worries, you can use this script to clean-up the processes:
# ./scripts/cleanup-test-processes.sh
#
# If you want to inspect log files:
# $ cd ${SWAP_TEST_DATA_DIR}
# Note: 'SWAP_TEST_DATA_DIR' is not defined until the first service
# is started.

MONEROD_PORT=18081
GANACHE_PORT=8545

BOB_WALLET_PORT=18083
ALICE_WALLET_PORT=18084
CHARLIE_WALLET_PORT=18085

PROJECT_ROOT="$(dirname "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")")"
MONERO_BIN_DIR="${PROJECT_ROOT}/monero-bin"

# return 0 (true) if the passed port is open, otherwise non-zero (false)
is-port-open() {
	local port="${1:?}"
	: &>/dev/null <"/dev/tcp/127.0.0.1/${port}"
}

monero-rpc-request() {
	local port="${1:?}"   # can be a monerod or monero-wallet-rpc port
	local method="${2:?}" # RPC method name
	local params="${3:?}" # JSON parameters to method
	curl "http://localhost:${port}/json_rpc" \
		--silent \
		--show-error \
		-d "{\"jsonrpc\":\"2.0\",\"id\":\"0\",\"method\":\"${method}\",\"params\":${params}" \
		-H 'Content-Type: application/json' \
		-w "\n"
}

check-set-swap-test-data-dir() {
	if [[ -z "${SWAP_TEST_DATA_DIR}" ]]; then
		SWAP_TEST_DATA_DIR="$(mktemp --tmpdir -d atomic-swap-test-data-XXXXXXXXXX)"
		echo "Swap test data dir is ${SWAP_TEST_DATA_DIR}"
	else
		mkdir -p "${SWAP_TEST_DATA_DIR}" # make sure it exists if the variable was already set
	fi
}

remove-test-data-dir() {
	# If this was run after all test processes have been cleaned up, the test data
	# directory should be empty. We don't force remove it, as it may have diagnostics
	# on what failed.
	if ! rmdir "${SWAP_TEST_DATA_DIR}"; then
		echo "ERROR: failed to remove ${SWAP_TEST_DATA_DIR} (probably not empty)"
		return 1
	fi
	return 0
}

stop-program() {
	local name="${1:?}" # Name of the PID file, usually the program name except when there are multiple instances
	check-set-swap-test-data-dir
	if [[ -z "${SWAP_TEST_DATA_DIR}" ]]; then
		echo "skipping stop of ${name}, SWAP_TEST_DATA_DIR variable is not set"
		return 0
	fi
	local pid_file="${SWAP_TEST_DATA_DIR}/${name}.pid"
	if [[ ! -e "${pid_file}" ]]; then
		echo "skipping stop of ${name}, ${pid_file} not found"
		return 0
	fi
	if ! kill "$(cat "${pid_file}")"; then
		echo "ERROR: failed to kill ${name}"
		return 1
	fi
	sleep 2 # let program flush data and exit so we delete all files below
	# Remove the PID file, log file and any data subdirectory
	rm -rf "${SWAP_TEST_DATA_DIR:?}/${name}"{.pid,.log,}
}

start-monerod-regtest() {
	./scripts/install-monero-linux.sh # install monero binaries if they are not already installed
	if is-port-open "${MONEROD_PORT}"; then
		echo "WARNING: Skipping launch of monerod, port ${MONEROD_PORT} is already in use"
		return 0 # Assume the user wanted to use the existing instance
	fi
	check-set-swap-test-data-dir
	echo "starting monerod..."
	"${MONERO_BIN_DIR}/monerod" \
		--detach \
		--regtest \
		--offline \
		--data-dir="${SWAP_TEST_DATA_DIR}/monerod" \
		--pidfile="${SWAP_TEST_DATA_DIR}/monerod.pid" \
		--fixed-difficulty=1 \
		--rpc-bind-ip=127.0.0.1 \
		--rpc-bind-port=18081 \
		--keep-fakechain
	sleep 5
}

stop-monerod-regtest() {
	stop-program monerod
}

# Installs ganache if it is not already installed.
install-ganache() {
	if ! command -v npm &>/dev/null; then
		echo "npm executable not found!"
		return 1
	fi
	# shellcheck disable=SC2155
	local npm_install_dir="$(npm config get prefix)"
	local ganache_exec="${npm_install_dir}/bin/ganache"
	if [[ -x "${ganache_exec}" ]]; then
		return 0 # ganache already installed
	fi
	echo "installing ganache"
	if [[ -d "${npm_install_dir}/bin" ]] && [[ ! -w "${npm_install_dir}/bin" ]]; then
		echo "${npm_install_dir}[/bin] is not writable"
		echo "You can use 'npm config set prefix DIRNAME' to pick a different install directory"
		return 1
	fi
	npm install --location=global ganache
}

start-ganache() {
	if is-port-open "${GANACHE_PORT}"; then
		echo "WARNING: Skipping launch of ganache, port ${GANACHE_PORT} is already open"
		return 0 # Assume the user wanted to use the existing instance
	fi
	if ! install-ganache; then
		echo "ERROR: Skipping launch of ganache due to errors"
		return 1
	fi
	echo "starting ganache ..."
	# shellcheck disable=SC2155
	local ganache_exec="$(npm config get prefix)/bin/ganache"
	check-set-swap-test-data-dir
	NODE_OPTIONS="--max_old_space_size=8192" nohup \
		"${ganache_exec}" --deterministic \
		--accounts=50 \
		--miner.blockTime=1 \
		&>"${SWAP_TEST_DATA_DIR}/ganache.log" &
	echo "${!}" >"${SWAP_TEST_DATA_DIR}/ganache.pid"
}

stop-ganache() {
	stop-program ganache
}

start-monero-wallet-rpc() {
	local wallet_user=$1 # alice, bob, charlie
	local wallet_port=$2

	check-set-swap-test-data-dir
	if is-port-open "${wallet_port}"; then
		echo "WARNING: Skipping launch of monero-wallet-rpc for ${wallet_user^}, port ${wallet_port} is already in use"
		return 0 # Assume the user wanted to use the existing instance
	fi
	if ! is-port-open "${MONEROD_PORT}"; then
		echo "ERROR: Aborting launch monero-wallet-rpc for ${wallet_user^}, monerod not detected on port ${MONEROD_PORT}"
		return 0 # Assume the user wanted to use the existing instance
	fi
	check-set-swap-test-data-dir
	local name="${wallet_user}-monero-wallet-rpc"
	local wallet_dir="${SWAP_TEST_DATA_DIR}/${name}"
	mkdir -p "${wallet_dir}"
	echo "Starting ${wallet_user^}'s monero-wallet-rpc on port ${wallet_port} ..."
	"${MONERO_BIN_DIR}/monero-wallet-rpc" \
		--detach \
		--rpc-bind-ip 127.0.0.1 \
		--rpc-bind-port "${wallet_port}" \
		--pidfile="${SWAP_TEST_DATA_DIR}/${name}.pid" \
		--log-file="${SWAP_TEST_DATA_DIR}/${name}.log" \
		--disable-rpc-login \
		--wallet-dir "${wallet_dir}"
}

start-alice-wallet() {
	start-monero-wallet-rpc alice "${ALICE_WALLET_PORT}"
}

stop-alice-wallet() {
	stop-program alice-monero-wallet-rpc
}

start-bob-wallet() {
	start-monero-wallet-rpc bob "${BOB_WALLET_PORT}"
	sleep 5
	# Send the json output to /dev/null, any serious errors will be to stderr
	monero-rpc-request "${BOB_WALLET_PORT}" create_wallet \
		'{"filename":"test-wallet","password":"","language":"English"}' >/dev/null
}

stop-bob-wallet() {
	stop-program bob-monero-wallet-rpc
}

start-charlie-wallet() {
	start-monero-wallet-rpc charlie "${CHARLIE_WALLET_PORT}"
}

stop-charlie-wallet() {
	stop-program charlie-monero-wallet-rpc
}
