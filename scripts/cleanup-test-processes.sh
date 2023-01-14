#!/usr/bin/env bash
# This script can be handy to kill test processes launched manually
# for debugging or still hanging around for other reasons.

pkill_cmd=(pkill --echo --uid "${UID}" --full)
if [[ "$(uname)" == 'Darwin' ]]; then
	pkill_cmd=(pkill -l -U "${UID}" -f)
fi

echo "Stoping any monerod regest instances"
"${pkill_cmd[@]}" '/monerod .* --regtest '

echo "Stoping any deterministic ganache instances"
"${pkill_cmd[@]}" '/ganache.* --deterministic '

echo "Stoping any relayer instances"
"${pkill_cmd[@]}" '/relayer'

# If you have monero-wallet-rpc or swapd processes owned by the current user
# that you don't want to kill, don't use this script!
echo "Stoping any swapd instances"
if "${pkill_cmd[@]}" '/swapd '; then
	# If swapd instances were killed, give the monero-wallet-rpc instances
	# some time to shutdown.
	sleep 4 # Give time for wallet to shutdown
fi

# Take note of monero-wallet-rpc instances being killed here if the instances
# were started by swapd
echo "Stoping any monero-wallet-rpc instances"
"${pkill_cmd[@]}" '/monero-wallet-rpc '

# Don't use the exit value of the last pkill, since it will exit with
# non-zero value if there were no processes to kill.
exit 0
