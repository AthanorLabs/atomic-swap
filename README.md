# ETH-XMR Atomic Swaps

This is a prototype of ETH<->XMR atomic swaps, which was worked on during ETHLisbon.

### Instructions

Start ganache-cli with determinstic keys:
```
ganache-cli -d
```

Start monerod for stagenet:
```
monerod --stagenet
```

Start monero-wallet-rpc:
```
./monero-wallet-rpc  --stagenet --rpc-bind-port 18082 --password "" --disable-rpc-login --wallet-dir .
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
