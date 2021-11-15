# ETH-XMR Atomic Swaps

This is a WIP prototype of ETH<->XMR atomic swaps, currently in the early development phase. It currently consists of a single `atomic-swap` binary which allows for peers to discover each other over the network based on what you want to swap for, querying peers for additional info such as their desired exchange rate, and the ability to initiate and perform the entire protocol. The `atomic-swap` program has a JSON-RPC endpoint which the user can use to interact with the process. 

### Protocol

Alice has ETH and wants XMR, Bob has XMR and wants ETH. They come to an agreement to do the swap and the amounts they will swap.

##### Initial (offchain) phase
- Alice and Bob each generate Monero secret keys (which consist of secret spend and view keys): (`s_a`, `v_a`) and (`s_b`, `v_b`), which are used to construct valid points on the ed25519 curve (ie. their public keys): `P_ed_a` and `P_ed_b` accordingly. Alice sends Bob her public keys and Bob sends Alice his public spend key and private view key. This is so Alice can check that Bob actually locked the amount of XMR he claims he will.

##### Step 1.
Alice deploys a smart contract on Ethereum and locks her ETH in it. The contract has the following properties:
- it is non-destructible

- it contains two timestamps, `t_0` and `t_1`, before and after which different actions are authorized.

- it is constructed containing `P_ed_a` & `P_ed_b`, so that if Alice or Bob reveals their secret by calling the contract, the contract will verify that the secret corresponds to the expected public key that it was initalized with.

- it has a `Ready()` function which can only be called by Alice. Once `Ready()` is invoked, Bob can proceed with redeeming his ether. Alice has until the `t_0` timestamp to call `Ready()` - once `t_0` passes, then the contract automatically allows Bob to claim his ether, up until some second timestamp `t_1`.

- it has a `Claim()` function which can only be called by Bob after `Ready()` is called or `t_0` passes, up until the timestamp `t_1`. After `t_1`, Bob can no longer claim the ETH.

- `Claim()` takes one parameter from Bob: `s_b`. Once `Claim()` is called, the ETH is transferred to Bob, and simultaneously Bob reveals his secret and thus Alice can claim her XMR by combining her and Bob's secrets.

- it has a `Refund()` function that can only be called by Alice and only before `Ready()` is called *or* `t_0` is reached. Once `Ready()` is invoked, Alice can no longer call `Refund()` until the next timestamp `t_1`.  If Bob doesn't claim his ether by `t_1`, then `Refund()` can be called by Alice once again.

- `Refund()` takes one parameter from Alice: `s_a`. This allows Alice to get her ETH back in case Bob goes offline, but it simulteneously reveals her secret, allowing Bob to regain access to the XMR he locked.

##### Step 2. 
Bob sees the smart contract has been deployed with the correct parameters. He sends his XMR to an account address constructed from `P_ed_a + P_ed_b`. Thus, the funds can only be accessed by an entity having both `s_a` & `s_b`, as the secret spend key to that account is `s_a + s_b`. The funds are viewable by someone having `v_a + v_b`.

Note: `Refund()` and `Claim()` cannot be called at the same time. This is to prevent the case of front-running where, for example, Bob tries to claim, so his secret `s_b` is in the mempool, and then Alice tries to call `Refund()` with a higher priority while also transferring the XMR in the account controlled by `s_a + s_b`. If her call goes through before Bob's and Bob doesn't notice this happening in time, then Alice will now have *both* the ETH and the XMR. Due to this case, Alice and Bob should not call `Refund()` or `Claim()` when they are approaching `t_0` or `t_1` respectively, as their transaction may not go through in time.

##### Step 3.
Alice sees that the XMR has been locked, and the amount is correct (as she knows `v_a` and Bob send her `v_b` in the first key exchange step). She calls `Ready()` on the smart contract if the XMR has been locked. If the amount of XMR locked is incorrect, Alice calls `Refund()` to abort the swap and reclaim her ETH.

From this point on, Bob can redeem his ether by calling `Claim(s_b)`, which transfers the ETH to him.

By redeeming, Bob reveals his secret. Now Alice is the only one that has both `s_a` & `s_b` and she can access the monero in the account created from `P_ed_a + P_ed_b`.

#### What could go wrong

- **Alice locked her ETH, but Bob doesn't lock his XMR**. Alice has until time `t_0` to call `Refund()` to reclaim her ETH, which she should do if `t_0` is soon.

- **Alice called `Ready()`, but Bob never redeems.** Deadlocks are prevented thanks to a second timelock `t_1`, which re-enables Alice to call refund after it, while disabling Bob's ability to claim.

- **Alice never calls `ready` within `t_0`**. Bob can still claim his ETH by waiting until after `t_0` has passed, as the contract automatically allows him to call `Claim()`.

### Requirements

- go 1.17

Note: this program has only been tested on ubuntu 20.04.

### Instructions

Start ganache-cli with determinstic keys:
```
ganache-cli -d
```

Note: the `scripts/run-unit-tests.sh` script will do the following setup for you including downloading the needed monero binaries and running the processes (up until the `make build` step)

Start monerod for regtest:
```
./monerod --regtest --fixed-difficulty=1 --rpc-bind-port 18081 --offline
```

Start monero-wallet-rpc for Bob with some wallet that has regtest monero:
```
./monero-wallet-rpc  --rpc-bind-port 18083 --password "" --disable-rpc-login --wallet-file test-wallet
```

Determine the address of `test-wallet` by running `monero-wallet-cli` and `address all`

Then, mine some blocks on the monero test chain by running the following RPC command, replacing the address with the one from the previous step:
```
curl -X POST http://127.0.0.1:18081/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"generateblocks","params":{ "wallet_address":"49oFJna6jrkJYvmupQktXKXmhnktf1aCvUmwp8HJGvY7fdXpLMTVeqmZLWQLkyHXuU9Z8mZ78LordCmp3Nqx5T9GFdEGueB","amount_of_blocks":100}' -H 'Content-Type: application/json'
```

This will deposit some XMR in your account.

Start monero-wallet-rpc for Alice:
```
./monero-wallet-rpc  --rpc-bind-port 18084 --password "" --disable-rpc-login --wallet-dir .
```

Build binary:
```
make build
```

This creates an `atomic-swap` binary in the root directory.

To run as Alice, execute in terminal 1:
```
./atomic-swap --amount 1 --alice
```

Alice will print out a libp2p node address, for example `/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWBW1cqB9t5fKP8yZPq3PcWcgbvuNai5ZpAeWFAbs5RNAA`. This will be used for Bob to connect.

To run as Bob and connect to Alice, replace the bootnode in the following line with what Alice logged, and execute in terminal 2:

```
./atomic-swap --amount 1 --bob --bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWBW1cqB9t5fKP8yZPq3PcWcgbvuNai5ZpAeWFAbs5RNAA
```

Note: amount doesn't matter at this point, it's only used in the `QueryResponse` message (ie. what's returned by `net_queryPeer`)

Note: Alice's RPC server runs on http://locahost:5001, Bob's runs on http://localhost:5002 by default.

In terminal 3, we will make RPC calls to the swap daemon.

This posts a call to Alice's daemon to begin discovering peers who provide XMR.
```
$ curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_discover","params":{"provides":"XMR"}}' -H 'Content-Type: application/json'
{"jsonrpc":"2.0","result":{"peers":[["/ip4/192.168.0.101/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7","/ip4/127.0.0.1/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7"]]},"id":"0"}
```

Get Alice to query the returned peer as to how much they XMR they can provide and their preferred exchange rate (replace `"multiaddr"` field with one of the addresses returned in the above step):
```
$ curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_queryPeer","params":{"multiaddr":"/ip4/38.88.101.233/tcp/41044/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7"}}' -H 'Content-Type: application/json'
{"jsonrpc":"2.0","result":{"provides":["XMR"],"maximumAmount":[33300],"exchangeRate":0.0578261},"id":"0"}
```

Now, we can tell Alice to initiate the protocol w/ the peer it found (which is Bob):
```
$ curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_initiate","params":{"multiaddr":"/ip4/38.88.101.233/tcp/41044/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7", "provides":"ETH", "providesAmount":333, "desiredAmount":33000 }}' -H 'Content-Type: application/json'
{"jsonrpc":"2.0","result":{"success":true},"id":"0"}
```

If all goes well, you should see Alice and Bob successfully exchange messages and execute the swap protocol. The result is that Alice now owns the private key to a Monero account (and is the only owner of that key) and Bob has the ETH transferred to him. On Alice's side, a Monero wallet will be generated in the `--wallet-dir` provided in the `monero-wallet-rpc` step for Alice.


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
To setup the test environment and run all unit tests, execute:
```
make test
```

This will test the main protocol functionality on the ethereum side:
1. Success case, where both parties obey the protocol
2. Case where Bob never locks monero on his side. Alice can Refund
3. Case where Bob locks monero, but never claims his ether from the contract

Upon Refund/Claim by either side, they reveal the secret to the counterparty, which *always* guarantees that the counteryparty can claim the locked funds on ethereum.
