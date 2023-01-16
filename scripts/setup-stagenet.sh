#!/usr/bin/env bash

# swapd will call monero-wallet-rpc, so make sure it is installed
bash ./scripts/install-monero-linux.sh

#
# Put your own ethereum endpoint below (or define ETHEREUM_ENDPOINT before
# invoking this script. There is no guarantee that the node below will be
# running or should be trusted when you test.
#
ETHEREUM_ENDPOINT="${ETHEREUM_ENDPOINT:-"https://ethereum-goerli-rpc.allthatnode.com"}"

#
# swapd has some preconfigured stagenet monerod nodes, but the best option
# is to run and sync your own local node. If you do this, add the flags
# --monerod-host=127.0.0.1 and --monero-port=38081 to swapd.
#
# ./monero-bin/monerod --detach --stagenet &

log_level=info # change to "debug" for more logs

./bin/swapd --env stagenet \
	"--log-level=${log_level}" \
	"--ethereum-endpoint=${ETHEREUM_ENDPOINT}" \
	&>swapd.log &

echo "swapd start with logs in swapd.log"
