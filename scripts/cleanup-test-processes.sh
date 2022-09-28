#!/bin/bash
# This script can be handy to kill test processes launched manually
# for debugging or still hanging around for other reasons.

pkill --echo --uid "${UID}" --full '/monerod .* --regtest '
pkill --echo --uid "${UID}" --full '/ganache.* --deterministic '

# If you have monero-wallet-rpc or swapd processes owned by the current user
# that you don't want to kill, don't use this script!
pkill --echo --uid "${UID}" --full '/swapd '

# In theory, killing the swapd processes should kill the monero-wallet-rpc
# processes. Sleep for a second so we know.
sleep 1
pkill --echo --uid "${UID}" --full '/monero-wallet-rpc '

# Don't use the exit value of the last pkill, since it will exit with
# non-zero value if there were no processes to kill.
exit 0
