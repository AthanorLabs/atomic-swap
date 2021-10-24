# ETH-XMR Atomic Swaps

This is a prototype of ETH<->XMR atomic swaps, which was worked on during ETHLisbon.

### Protocol

1. Alice has ETH and wants XMR, Bob has XMR and wants ETH. They come to an agreement to do the swap and the amounts they will swap.
2. Alice and Bob each generate a private Monero spend and view key (s, v) and their corresponding public keys (S, V).
3.

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

Build binary:
```
./scripts/build.sh
```

This creates an `atomic-swap` binary in the root directory.

To run as Alice, execute:
```
./atomic-swap --amount 1 --alice
```

Alice will print out a libp2p node address, for example `/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWBW1cqB9t5fKP8yZPq3PcWcgbvuNai5ZpAeWFAbs5RNAA`. This will be used for Bob to connect.

To run as Bob and connect to Alice, replace the bootnode in the following line with what Alice logged, and execute:

```
./atomic-swap --amount 1 --bob --bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWBW1cqB9t5fKP8yZPq3PcWcgbvuNai5ZpAeWFAbs5RNAA
```

If all goes well, you should see Alice and Bob successfully exchange messages and execute the swap protocol. The result is that Alice now owns the private key to a Monero account (and is the only owner of that key) and Bob has the ETH transferred to him.


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
