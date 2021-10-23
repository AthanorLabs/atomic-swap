# ETH-XMR Atomic Swaps

This is a prototype of ETH<->XMR atomic swaps, which was worked on during ETHLisbon.

### Instructions

Start ganache-cli with determinstic keys:
```
ganache-cli -d
```

Start monerod for regtest:
```
./monerod --regtest --fixed-difficulty=1 --rpc-bind-port 18081 --offline
```

Start monero-wallet-rpc for Bob with some wallet that has regtest monero:
```
./monero-wallet-rpc  --rpc-bind-port 18083 --password "" --disable-rpc-login --wallet-file test-wallet
```

Start monero-wallet-rpc for Alice:
```
./monero-wallet-rpc  --rpc-bind-port 18084 --password "" --disable-rpc-login --wallet-dir .
```

##### Compiling contract bindings

Download solc v0.8.9

Set `SOLC_BIN` to the downloaded binary
```
export SOLC_BIN=solc
```

Generate the bindings
```
./scripts/generate-bindings.sh
```

##### Testing
To run tests on the go bindings, execute
```
go test ./swap-contract
```
