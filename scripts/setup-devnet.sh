#!/bin/bash

# useful dir relative to this script
MONERO_DIR="../monero-x86_64-linux-gnu-v0.17.3.0"
LOG_DIR="../log"
ALICE_P2P_ADDRESS="12D3KooWF5dTdfrVv6oFwGGGyobfxtZBVhVR654wt5ED6PU1SBqd"

echo "cleanup"
pkill -e -f monero;
pkill -e -f ganache-cli;
killall -v swapd;
pkill -e -f swapcli;
rm -rf $LOG_DIR/*;

echo "start ganache-cli"
ganache-cli -d &> $LOG_DIR/ganache-cli.log &

echo "move to $MONERO_DIR"
cd $MONERO_DIR 

echo "starting monerod..."
./monerod --regtest --detach --fixed-difficulty=1 --rpc-bind-port 18081 --offline &> $LOG_DIR/monerod.log &

echo "Zzz... 10s"
sleep 10

echo "mine blocks for Bob"
curl -X POST http://127.0.0.1:18081/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"generateblocks","params":{"wallet_address":"45GcPCBQgCG3tYcYqLdj4iQixpDZYw1MGew4PH1rthp9X2YrB2c2dty1r7SwhbCXw1RJMvfy8cW1UXyeESTAuLkV5bTrZRe","amount_of_blocks":100}' -H 'Content-Type: application/json' &> $LOG_DIR/block-mining-bob.log &

echo "Zzz... 15s"
sleep 15

echo "start monero-wallet-rpc for Alice on port 18084"
./monero-wallet-rpc  --rpc-bind-port 18084 --password "" --disable-rpc-login --wallet-dir . &> $LOG_DIR/alice-wallet-rpc.log &

echo "start monero-wallet-rpc for Bob on port 18083"
./monero-wallet-rpc --rpc-bind-port 18083 --password "" --disable-rpc-login --wallet-dir . &> $LOG_DIR/bob-wallet-rpc.log &

echo "launch Alice swapd"
../swapd --dev-alice  &> $LOG_DIR/alice-swapd.log &

echo "Zzz... 10s"
sleep 10

echo "launch Bob swapd"
../swapd --dev-bob --wallet-file Bob --bootnodes /ip4/127.0.0.1/tcp/9933/p2p/$ALICE_P2P_ADDRESS &> $LOG_DIR/bob-swapd.log &

echo "Zzz... 10s"
sleep 10

echo "let Bob make an offer"
../swapcli make --min-amount 0.1 --max-amount 1 --exchange-rate 0.05 --daemon-addr=http://localhost:5002
