# ETH-XMR Atomic Swaps

This is a WIP prototype of ETH<->XMR atomic swaps, currently in the early development phase. It currently consists of a single `atomic-swap` binary which allows for peers to discover each other over the network based on what you want to swap for, querying peers for additional info such as their desired exchange rate, and the ability to initiate and perform the entire protocol. The `atomic-swap` program has a JSON-RPC endpoint which the user can use to interact with the process. 

## Disclaimer

**This code is unaudited and under active development and should not be used on mainnet!** Running this on mainnet may result in loss of funds.

## Protocol

Please see the [protocol documentation](docs/protocol.md) for how it works.

## Instructions

### Requirements

- go 1.17
- ganache-cli (can be installed with `npm i -g ganache-cli`) I suggest using nvm to install npm: https://github.com/nvm-sh/nvm#installing-and-updating

Note: this program has only been tested on Ubuntu 20.04.

#### Set up development environment

Note: the `scripts/install-monero-linux.sh` script will download the monero binaries needed for you. You can also check out the `scripts/run-unit-tests.sh` script for the commands needed to setup the environment.

Start ganache-cli with determinstic keys:
```
ganache-cli -d
```

Start monerod for regtest:
```
./monerod --regtest --fixed-difficulty=1 --rpc-bind-port 18081 --offline
```

Create a wallet (will represent Bob, who owns XMR):
```
./monero-wallet-cli // you will be prompted to create a wallet. remember the name and optionally the password for the upcoming steps
```

Start monero-wallet-rpc for Bob creating a wallet (make sure `--wallet-dir` corresponds to the directory the wallet from the previous step is in:
```
./monero-wallet-rpc  --rpc-bind-port 18083 --password "" --disable-rpc-login --wallet-dir .
```

Open the wallet:
```
$ curl http://localhost:18082/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"open_wallet","params":{"filename":"mytestwallet","password":"mytestpassword"}}' -H 'Content-Type: application/json'
{
  "id": "0",
  "jsonrpc": "2.0",
  "result": {
  }
}
```

Determine the address of `test-wallet` by running `monero-wallet-cli` and `address all`

Then, mine some blocks on the monero test chain by running the following RPC command, replacing the address with the one from the previous step:
```
curl -X POST http://127.0.0.1:18081/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"generateblocks","params":{"wallet_address":"49oFJna6jrkJYvmupQktXKXmhnktf1aCvUmwp8HJGvY7fdXpLMTVeqmZLWQLkyHXuU9Z8mZ78LordCmp3Nqx5T9GFdEGueB","amount_of_blocks":100}' -H 'Content-Type: application/json'
```

This will deposit some XMR in your account.

Start monero-wallet-rpc for Alice (note that the directory provided to `--wallet-dir` is where Alice's XMR wallet will end up):
```
./monero-wallet-rpc  --rpc-bind-port 18084 --password "" --disable-rpc-login --wallet-dir .
```

#### Build and run

Build binary:
```
make build
```

This creates `swapd` and `swapcli` binaries in the root directory.

To run as Alice, execute in terminal 1:
```
./swapd --amount 1 --alice
```

Alice will print out a libp2p node address, for example `/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWBW1cqB9t5fKP8yZPq3PcWcgbvuNai5ZpAeWFAbs5RNAA`. This will be used for Bob to connect.

To run as Bob and connect to Alice, replace the bootnode in the following line with what Alice logged, and execute in terminal 2:

```
./swapd --amount 1 --bob --bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWBW1cqB9t5fKP8yZPq3PcWcgbvuNai5ZpAeWFAbs5RNAA
```

Note: amount doesn't matter at this point, it's only used in the `QueryResponse` message (ie. what's returned by `swapcli query`)

Note: Alice's RPC server runs on http://localhost:5001, Bob's runs on http://localhost:5002 by default.

In terminal 3, we will interact with the swap daemon using `swapcli`.

This posts a call to Alice's daemon to begin discovering peers who provide XMR.
```
$ ./swapcli discover --provides XMR --search-time 3
[[/ip4/192.168.0.101/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7 /ip4/127.0.0.1/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7 /ip4/38.88.101.233/tcp/48161/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7]]
```

Query the returned peer as to how much they XMR they can provide and their preferred exchange rate (replace `"--multiaddr"` field with one of the addresses returned in the above step):
```
$ ./swapcli query --multiaddr /ip4/192.168.0.101/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7
Provides: [XMR]
MaximumAmount: [1]
ExchangeRate (ETH/XMR): 0.0578261
```

Now, we can tell Alice to initiate the protocol w/ the peer it found (which is Bob):
```
$ ./swapcli initiate --multiaddr /ip4/192.168.0.101/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7 --provides ETH --provides-amount 100000000000000000 --desired-amount 1000000000
```

If all goes well, you should see Alice and Bob successfully exchange messages and execute the swap protocol. The result is that Alice now owns the private key to a Monero account (and is the only owner of that key) and Bob has the ETH transferred to him. On Alice's side, a Monero wallet will be generated in the `--wallet-dir` provided in the `monero-wallet-rpc` step for Alice.

### Developer instructions

##### Compiling contract bindings

If you update the `Swap.sol` contract for some reason, you will need to re-generate the Go bindings for the contract. **Note:** you do *not* need to do this to try out the swap; only if you want to edit the contract for development purposes.

Download solc v0.8.9: https://github.com/ethereum/solidity/releases/tag/v0.8.9

Set `SOLC_BIN` to the downloaded binary
```
export SOLC_BIN=solc
```

Install `abigen`
```
git clone https://github.com/ethereum/go-ethereum.git && cd go-ethereum/cmd/abigen
go install
```

Generate the bindings
```
./scripts/generate-bindings.sh
```
Note: you may need to add `$GOPATH` and `$GOPATH/bin` to your path.

#### Testing
To setup the test environment and run all unit tests, execute:
```
make test
```

This will test the main protocol functionality on the ethereum side:
1. Success case, where both parties obey the protocol
2. Case where Bob never locks monero on his side. Alice can Refund
3. Case where Bob locks monero, but never claims his ether from the contract

Upon Refund/Claim by either side, they reveal the secret to the counterparty, which *always* guarantees that the counteryparty can claim the locked funds on ethereum.

## Contributions

If you'd like to contribute, feel free to fork the repo and make a pull request. Please make sure the CI is passing - you can run `make build`, `make lint`, and `make test` to make sure the checks pass locally.

## Donations

The work on this project is currently not funded. If you'd like to donate, you can send XMR to the following address: `48WX8KhD8ECgnRBonmDdUecGt8LtQjjesDRzxAjj7tie4CdhtqeBjSLWHhNKMc52kWayq365czkN3MV62abQobTcT1Xe6xC`