#!/usr/bin/env bash
# This script can be handy to kill test processes launched manually
# for debugging or that are still hanging around for other reasons. If
# you have monero-wallet-rpc or swapd processes owned by the current
# user that you don't want killed, don't use this script!

pkill_cmd=(pkill --echo --uid "${UID}" --full)
if [[ "$(uname)" == 'Darwin' ]]; then
	pkill_cmd=(pkill -l -U "${UID}" -f)
fi

echo "Stopping any swapd instances"
if "${pkill_cmd[@]}" '/swapd '; then
	# If swapd instances were killed, give dependent monero-wallet-rpc instances
	# some time to shutdown.
	sleep 4 # Give time for wallet to shutdown
fi

# Take note of monero-wallet-rpc instances being killed here if the instances
# were started by swapd, as killing swapd should have already shut them down.
echo "Stopping any monero-wallet-rpc instances"
"${pkill_cmd[@]}" '/monero-wallet-rpc '

echo "Stopping any bootnode instances"
"${pkill_cmd[@]}" '/bootnode '

echo "Stopping any monerod regest instances"
if "${pkill_cmd[@]}" '/monerod .* --regtest '; then
	sleep 2 # we don't want to exit the script while it is still running
fi

echo "Stopping any ganache instances"
"${pkill_cmd[@]}" '/ganache '

# Don't use the exit value of the last pkill, since it will exit with
# non-zero value if there were no processes to kill.
exit 0
