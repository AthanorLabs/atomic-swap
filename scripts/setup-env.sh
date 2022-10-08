#!/bin/bash
# If you need to debug unit tests, this script will turn up the necessary
# daemons without running the tests. Warning: It will terminate any
# currently running test daemons.

if [[ "${BASH_SOURCE[0]}" != "$0" ]]; then
	echo "Execute ${BASH_SOURCE[0]} instead of souring it"
	return
fi

SCRIPTS_DIR="$(dirname "${BASH_SOURCE[0]}")"
"${SCRIPTS_DIR}/cleanup-test-processes.sh"
sleep 2 # give monerod time to fully shutdown
source "${SCRIPTS_DIR}/testlib.sh"

start-monerod-regtest
start-ganache

printf "\nTo stop the daemons and, optionally, cleanup:\n\n"
printf "\t%s/cleanup-test-processes.sh\n" "${SCRIPTS_DIR}"
printf "\trm -rf %s\n\n", "${SWAP_TEST_DATA_DIR}"
