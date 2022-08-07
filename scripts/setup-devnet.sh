#!/bin/bash

# useful dir relative to this script
MONERO_DIR="../monero-bin"
# either a TMPDIR is set, or use /tmp
LOG_DIR=${TMPDIR:-"/tmp"}
ALICE_P2P_ADDRESS="12D3KooWBD82zGTFqk6Qmu5zeS6dQfiaAcn8go2QWE29HPmRX3yB"

echo "cleanup"
pkill -e -f monero
pkill -e -f ganache
killall -v swapd
pkill -e -f swapcli

echo "start ganache"
"$(npm config get prefix)/bin/ganache" --deterministic --accounts=50 --miner.blockTime=1 &> "${LOG_DIR}/ganache.log" &

echo "move to $MONERO_DIR"
cd "${MONERO_DIR}"

echo "starting monerod..."
./monerod --regtest --detach --fixed-difficulty=1 --rpc-bind-port 18081 --offline &> "${LOG_DIR}/monerod.log" &

echo "Zzz... 10s"
sleep 10

echo "mine blocks for XMRMaker"
curl -X POST http://127.0.0.1:18081/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"generateblocks","params":{"wallet_address":"43wote1FPHrQQL35p3LMbNGi4J6zLcwUF9EZiw2xKfyzbQVhFXQ3VcmFuM4RDK7gxh8FGgN2C3ssXcSeJR2wY2Gx92b5gxn","amount_of_blocks":100}' -H 'Content-Type: application/json' &> $LOG_DIR/block-mining-bob.log &

echo "Zzz... 15s"
sleep 15

echo "start monero-wallet-rpc for XMRTaker on port 18084"
./monero-wallet-rpc  --rpc-bind-port 18084 --password "" --disable-rpc-login --wallet-dir . &> "${LOG_DIR}/alice-wallet-rpc.log" &

echo "start monero-wallet-rpc for XMRMaker on port 18083"
./monero-wallet-rpc --rpc-bind-port 18083 --password "" --disable-rpc-login --wallet-dir . &> "${LOG_DIR}/bob-wallet-rpc.log" &

echo "launch XMRTaker swapd"
../swapd --dev-xmrtaker --external-signer --contract-address 0xe78A0F7E598Cc8b0Bb87894B0F60dD2a88d6a8Ab &> "${LOG_DIR}/alice-swapd.log" &

echo "Zzz... 10s"
sleep 10

echo "launch XMRMaker swapd"
../swapd --dev-xmrmaker --wallet-file XMRMaker --bootnodes "/ip4/127.0.0.1/tcp/9933/p2p/${ALICE_P2P_ADDRESS}" &> "${LOG_DIR}/bob-swapd.log" &

echo "Zzz... 10s"
sleep 10

echo "let XMRMaker make an offer"
../swapcli make --min-amount 0.1 --max-amount 1 --exchange-rate 0.05 --daemon-addr=http://localhost:5002
