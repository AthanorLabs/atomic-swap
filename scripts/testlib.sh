#!/usr/bin/env bash
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

PROJECT_ROOT="$(dirname "$(dirname "$(realpath "${BASH_SOURCE[0]}")")")"
MONERO_BIN_DIR="${PROJECT_ROOT}/monero-bin"

# return 0 (true) if the passed TCP port is open, otherwise non-zero (false)
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

mine-monero() {
	local monero_addr="${1:?}"  # primary monero address (required)
	local num_blocks="${2:-64}" # defaults to 32 if not passed
	monero-rpc-request "${MONEROD_PORT}" "generateblocks" \
		"{\"amount_of_blocks\":${num_blocks},\"wallet_address\":\"${monero_addr}\"}"
}

mine-monero-for-swapd() {
	local swapd_port="${1:-5001}" # defaults to 5001 if not passed
	local wallet_addr
	wallet_addr="$(
		"${PROJECT_ROOT}/bin/swapcli" balances --swapd-port "${swapd_port}" | grep 'Monero address:' | sed 's/.*: //'
	)"
	echo "mining to address ${wallet_addr}"
	mine-monero "${wallet_addr}"
}

check-set-swap-test-data-dir() {
	if [[ -z "${SWAP_TEST_DATA_DIR}" ]]; then
		SWAP_TEST_DATA_DIR="$(mktemp -d "${TMPDIR:-/tmp}/atomic-swap-test-data-XXXXXXXXXX")"
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
	if [[ "${KEEP_TEST_DATA}" -ne 1 ]]; then
		sleep 2 # let program flush data and exit so we delete all files below
		# Remove the PID file, log file and any data subdirectory
		rm -rf "${SWAP_TEST_DATA_DIR:?}/${name}"{.pid,.log,}
	fi
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
		--rpc-bind-port=18081
	sleep 5
	# Make sure the blockchain has some initial decoy outputs. Arbitrarily sending the
	# rewards to the Mastering Monero address.
	local rewardsAddr=4BKjy1uVRTPiz4pHyaXXawb82XpzLiowSDd8rEQJGqvN6AD6kWosLQ6VJXW9sghopxXgQSh1RTd54JdvvCRsXiF41xvfeW5
	mine-monero "${rewardsAddr}" >/dev/null
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
	echo "installing ganache in ${npm_install_dir}/bin"
	if [[ -d "${npm_install_dir}/bin" ]] && [[ ! -w "${npm_install_dir}/bin" ]]; then
		echo "${npm_install_dir}[/bin] is not writable"
		echo "You can use 'npm config set prefix DIRNAME' to pick a different install directory"
		return 1
	fi
	npm install --global ganache
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
	sleep 2
}

stop-ganache() {
	stop-program ganache
}
