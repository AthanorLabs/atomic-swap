#!/bin/bash

MONERO_DIR="../monero-x86_64-linux-gnu-v0.17.3.0"

# bash ./scripts/install-monero-linux.sh
echo "========== start ganache-cli"
ganache-cli -d &

echo "========== move to $MONERO_DIR"
cd $MONERO_DIR 

echo "========== starting monerod..."
./monerod --regtest --fixed-difficulty=1 --rpc-bind-port 18081 --offline &
sleep 15

echo "========== mine blocks for Bob"
curl -X POST http://127.0.0.1:18081/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"generateblocks","params":{"wallet_address":"45GcPCBQgCG3tYcYqLdj4iQixpDZYw1MGew4PH1rthp9X2YrB2c2dty1r7SwhbCXw1RJMvfy8cW1UXyeESTAuLkV5bTrZRe","amount_of_blocks":100}' -H 'Content-Type: application/json'

echo "========== start monero-wallet-rpc for Alice on port 18084"
./monero-wallet-rpc  --rpc-bind-port 18084 --password "" --disable-rpc-login --wallet-dir . &

echo "========== start monero-wallet-rpc for Bob on port 18083"
./monero-wallet-rpc --rpc-bind-port 18083 --password "" --disable-rpc-login --wallet-dir . &

echo "========== go back to root"
cd ..

echo "launch Alice swapd"
./swapd --dev-alice &

echo "launch Bob swapd"
./swapd --dev-bob --wallet-file Bob --bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWFUEQpGHQ3PtypLvgnWc5XjrqM2zyvdrZXin4vTpQ6QE5 &

echo "let Bob make an offer"
./swapcli make --min-amount 0.1 --max-amount 1 --exchange-rate 0.05 --daemon-addr=http://localhost:5002
