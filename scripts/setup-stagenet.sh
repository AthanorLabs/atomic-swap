#!/bin/bash

bash ./scripts/install-monero-linux.sh
echo "starting monerod..."

./monero-bin/monerod --detach --stagenet --rpc-bind-port 18081 &
sleep 5

echo "starting monero-wallet-rpc on port 18083..."
nohup ./monero-bin/monero-wallet-rpc --rpc-bind-port 18083 --disable-rpc-login --wallet-dir ./node-keys --stagenet --trusted-daemon &> monero-wallet-cli.log &

# open wallet (must have funds)
sleep 5
curl http://localhost:18083/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"open_wallet","params":{"filename":"stagenet-wallet","password":""}}' -H 'Content-Type: application/json'

# check balance
curl http://localhost:18083/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":get_accounts","params":{}}' -H 'Content-Type: application/json'

# start mining (if synced)
# curl http://localhost:18083/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":start_mining","params":{"threads_count":1,"do_background_mining":true,"ignore_battery":false}}' -H 'Content-Type: application/json'
