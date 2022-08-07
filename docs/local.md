# Trying the swap locally

### Requirements

- go 1.17+ (see [build instructions](./build.md) to download Go.)
- ganache (can be installed with `npm install --location=global ganache`)

These programs and scripts have only been tested on X86-64 Ubuntu 20.04 and 22.04.
Using nvm is [the suggested way](https://github.com/nvm-sh/nvm#installing-and-updating)
to install npm. If you install npm using a package manager like snap, ensure the install
prefix (`npm config get prefix`) is a directory that you have write access to without sudo.
You can change the directory with the command `npm config set prefix ~/.npm-packages`. See
[this document](https://github.com/sindresorhus/guides/blob/main/npm-global-without-sudo.md)
if you want a more sophisticated setup.

#### Set up development environment

Note: the `scripts/install-monero-linux.sh` script will download the monero binaries needed for you. You can also check out the `scripts/run-unit-tests.sh` script for the commands needed to setup the environment.

Start ganache with deterministic keys. We disable instamine, for a more realistic
simulation, by setting the miner.blockTime.
```bash
ganache --deterministic --accounts=50 --miner.blockTime=1
```

Start monerod for regtest, this binary is in the monero bin directory:
```bash
cd ./monero-bin
./monerod --regtest --fixed-difficulty=1 --rpc-bind-ip 127.0.0.1 --rpc-bind-port 18081 --offline
```

Create a wallet for "Bob", who will own XMR later on:
```bash
./monero-wallet-cli // you will be prompted to create a wallet. In the next steps, we will go with "Bob", without password. Remember the name and optionally the password for the upcoming steps
```

You do not need to mine blocks, and you can exit the the wallet-cli once Bob's account has been created by typing "exit".

Start monero-wallet-rpc for Bob on port 18083. Make sure `--wallet-dir` corresponds to the directory the wallet from the previous step is in:
```bash
./monero-wallet-rpc --rpc-bind-ip 127.0.0.1 --rpc-bind-port 18083 --password "" --disable-rpc-login --wallet-dir .
```

Open the wallet:
```bash
curl http://localhost:18083/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"open_wallet","params":{"filename":"Bob","password":""}}' -H 'Content-Type: application/json'

# {
#   "id": "0",
#   "jsonrpc": "2.0",
#   "result": {
#   }
# }
```

Determine the address of `Bob` by looking at `monero-wallet-rpc` logs, in our case 45GcPCB ... uLkV5bTrZRe
```bash
# 2022-01-20 21:40:06.460	W Loaded wallet keys file, with public address: 45GcPCBQgCG3tYcYqLdj4iQixpDZYw1MGew4PH1rthp9X2YrB2c2dty1r7SwhbCXw1RJMvfy8cW1UXyeESTAuLkV5bTrZRe
```

Then, mine some blocks on the monero test chain by running the following RPC command, replacing the address with the one from Bob's wallet:
```bash
curl -X POST http://127.0.0.1:18081/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"generateblocks","params":{"wallet_address":"45GcPCBQgCG3tYcYqLdj4iQixpDZYw1MGew4PH1rthp9X2YrB2c2dty1r7SwhbCXw1RJMvfy8cW1UXyeESTAuLkV5bTrZRe","amount_of_blocks":100}' -H 'Content-Type: application/json'
```

This will deposit some XMR in Bob's account.


Start monero-wallet-rpc for Alice on port 18084 (note that the directory provided to `--wallet-dir` is where Alice's XMR wallet will end up):
```bash
./monero-wallet-rpc --rpc-bind-ip 127.0.0.1 --rpc-bind-port 18084 --password "" --disable-rpc-login --wallet-dir .
```
#### Build and run

Build binary:
```bash
make build
```

This creates `swapd` and `swapcli` binaries in the root directory.

To run as Alice, execute in terminal 1:
```bash
./swapd --dev-xmrtaker
```

Alice will print out a libp2p node address, for example `/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWFUEQpGHQ3PtypLvgnWc5XjrqM2zyvdrZXin4vTpQ6QE5`. This will be used for Bob to connect.

To run as Bob and connect to Alice, replace the bootnode in the following line with what Alice logged, and execute in terminal 2:

```bash
./swapd --dev-xmrmaker --wallet-file Bob --bootnodes /ip4/127.0.0.1/tcp/9933/p2p/12D3KooWFUEQpGHQ3PtypLvgnWc5XjrqM2zyvdrZXin4vTpQ6QE5
```

Note: when using the `--dev-xmrtaker` and `--dev-xmrmaker` flags, Alice's RPC server runs on http://localhost:5001, Bob's runs on http://localhost:5002 by default.

In terminal 3, we will interact with the swap daemon using `swapcli`.

Firstly, we need Bob to make an offer and advertise it, so that Alice can take it:
```bash
./swapcli make --min-amount 0.1 --max-amount 1 --exchange-rate 0.05 --daemon-addr=http://localhost:5002
# Published offer with ID cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9
```

Alternatively, you can make the offer via websockets and get notified when the swap is taken:
```bash
./swapcli make --min-amount 0.1 --max-amount 1 --exchange-rate 0.05 --daemon-addr=ws://localhost:8082 --subscribe
```

Now, we can have Alice begin discovering peers who have offers advertised.
```bash
./swapcli discover --provides XMR --search-time 3
# [[/ip4/127.0.0.1/tcp/9934/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv /ip4/127.0.0.1/tcp/9934/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv]]
```

Query the returned peer as to how much XMR they can provide and their preferred exchange rate (replace `"--multiaddr"` field with one of the addresses returned in the above step):
```bash
./swapcli query --multiaddr /ip4/192.168.0.101/tcp/9934/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv
# Offer ID=cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9 Provides=XMR MinimumAmount=0.1 MaximumAmount=1 ExchangeRate=0.05
```

Now, we can tell Alice to initiate the protocol w/ the peer (Bob), the offer (copy the Offer id from above), and a desired amount to swap:
```bash
./swapcli take --multiaddr /ip4/192.168.0.101/tcp/9934/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv --offer-id cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9 --provides-amount 0.05
# Initiated swap with ID=0
```

Alternatively, you can take the offer via websockets and get notified when the swap status updates:
```bash
./swapcli take --multiaddr /ip4/127.0.0.1/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7 --offer-id cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9 --provides-amount 0.05 --subscribe --daemon-addr=ws://localhost:8081
```

If all goes well, you should see Alice and Bob successfully exchange messages and execute the swap protocol. The result is that Alice now owns the private key to a Monero account (and is the only owner of that key) and Bob has the ETH transferred to him. On Alice's side, a Monero wallet will be generated in the `--wallet-dir` provided in the `monero-wallet-rpc` step for Alice.

To query the information for an ongoing swap, you can run:
```bash
./swapcli get-ongoing-swap
```

To query information for a past swap using its ID, you can run:
```bash
./swapcli get-past-swap --id <id>
```
