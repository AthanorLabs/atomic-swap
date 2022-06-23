#!/bin/bash

bash ./scripts/install-monero-linux.sh
echo "starting monerod..."
nohup ./monero-x86_64-linux-gnu-v0.17.3.2/monerod --detach --stagenet --rpc-bind-port 18081 &
sleep 5

echo "starting monero-wallet-rpc on port 18083..."
nohup ./monero-x86_64-linux-gnu-v0.17.3.2/monero-wallet-rpc --rpc-bind-port 18083 --disable-rpc-login --wallet-dir ./bob-test-keys --stagenet --trusted-daemon 

echo "starting monero-wallet-rpc on port 18084..."
nohup ./monero-x86_64-linux-gnu-v0.17.3.2/monero-wallet-rpc --rpc-bind-port 18084 --disable-rpc-login --wallet-dir ./alice-test-keys --stagenet --trusted-daemon

# open Bob's wallet (must have funds)
sleep 5
curl http://localhost:18083/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"open_wallet","params":{"filename":"stagenet-wallet","password":""}}' -H 'Content-Type: application/json'

# check balance
curl http://localhost:18083/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":get_accounts","params":{}}' -H 'Content-Type: application/json'

# start mining (if synced)
# curl http://localhost:18083/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":start_mining","params":{"threads_count":1,"do_background_mining":true,"ignore_battery":false}}' -H 'Content-Type: application/json'
