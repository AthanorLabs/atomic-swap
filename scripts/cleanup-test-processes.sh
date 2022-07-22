#!/bin/bash
# This script can be handy to kill test processes launched manually
# for debugging or still hanging around for other reasons.

pkill --uid "${UID}" --full '/monerod .* --regtest '
pkill --uid "${UID}" --full '/ganache.* --deterministic '

# If you have monero-wallet-rpc or swapd processes owned by the current user
# that you don't want to kill, don't use this script!
pkill --uid "${UID}" --full '/monero-wallet-rpc '
pkill --uid "${UID}" --full '/swapd '

# These directories MUST be removed every time you start a fresh monerod instance
rm -rf alice-test-keys bob-test-keys
