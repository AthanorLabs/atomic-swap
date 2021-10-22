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

Download solc v0.6.12

```
./solc-static-linux --bin contracts/contracts/Swap.sol -o contracts/bin/ --overwrite
./solc-static-linux --abi contracts/contracts/Swap.sol -o contracts/abi/ --overwrite
```

```
abigen --abi contracts/abi/Swap.abi --pkg swap --type Swap --out swap.go --bin contracts/bin/Swap.bin 
```