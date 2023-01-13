#!/bin/bash
# This script can be handy to kill test processes launched manually
# for debugging or still hanging around for other reasons.

pkill --echo --uid "${UID}" --full '/monerod .* --regtest '
pkill --echo --uid "${UID}" --full '/ganache.* --deterministic '
pkill --echo --uid "${UID}" --full '/relayer'

# If you have monero-wallet-rpc or swapd processes owned by the current user
# that you don't want to kill, don't use this script!
if pkill --echo --uid "${UID}" --full '/swapd '; then
	# If swapd instances were killed, give the monero-wallet-rpc instances
	# some time to shutdown.
	sleep 4 # Give time for wallet to shutdown
fi

# Take note of monero-wallet-rpc instances being killed here if the instances
# were started by swapd
pkill --echo --uid "${UID}" --full '/monero-wallet-rpc '

# Don't use the exit value of the last pkill, since it will exit with
# non-zero value if there were no processes to kill.
exit 0
