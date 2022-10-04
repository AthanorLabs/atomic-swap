# Trying the swap locally

### Requirements

- go 1.18+ (see [build instructions](./build.md) to download Go.)
- ganache (can be installed with `npm install --location=global ganache`)

These programs and scripts have only been tested on X86-64 Ubuntu 20.04 and 22.04.
Using nvm is [the suggested way](https://github.com/nvm-sh/nvm#installing-and-updating)
to install npm. If you install npm using a package manager like snap, ensure the install
prefix (`npm config get prefix`) is a directory that you have write access to without sudo.
You can change the directory with the command `npm config set prefix ~/.npm-packages`. See
[this document](https://github.com/sindresorhus/guides/blob/main/npm-global-without-sudo.md)
if you want a more sophisticated setup.

#### Set up development environment

Note: the `scripts/install-monero-linux.sh` script will download the monero binaries needed for you.
You can invoke it directly, but the next script below will run it if there is no symbolic link named
`monero-bin` to a monero installation in the project's root directory.

Execute the `scripts/setup-env.sh` script to launch ganache, an ethereum simulator, and monerod in regtest
mode. "regtest" mode is stand-alone (non-networked) mode of monerod for testing purposes.

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

Alice will print out a libp2p node address, for example
`/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAAxG7eTEHr2uBVw3BDMxYsxyqfKvj3qqqpRGtTfuzTuH`.
This will be used for Bob to connect. You can either grab this address from the
logs, our you can obtain it with this command:
```bash
./swapcli addresses
```
Pick the localhost address and assign it to a variable. For example (your value will be different):
```bash
BOOT_NODE=/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWHRi24PVZ6TBnQJHdVyewDRcKFZtYV3qmB4KQo8iMyqik
```
Now get the ethereum contract address that Alice deployed to. This can be pulled from the Alice's logs,
the file ..., or if you have `jq` installed (available via `sudo apt install jq`), you can set a
variable like this:
```bash
CONTRACT_ADDR=$(jq -r .ContractAddress /tmp/xmrtaker/contract-address.json)
```

Now start Bob's swapd instance in terminal 2:
```bash
./swapd --dev-xmrmaker --bootnodes "${BOOT_NODE}" --contract-address "${CONTRACT_ADDR}"
```

Note: when using the `--dev-xmrtaker` and `--dev-xmrmaker` flags, Alice's RPC server runs
on http://localhost:5001 and Bob's runs on http://localhost:5002 by default.

Now, in terminal 3, we will interact with the swap daemon using `swapcli`.

First we need mine some monero for Bob. Alice already has Ethereum, because she is using
a prefunded by ganache address. You can see the balances for Bob with the following
command:
```bash
./swapcli balances --swapd-port 5002
```
Note that Alice is on the default swapd port of 5001, so the `--swapd-port` flag is optional
when interacting with her daemon.

To mine some monero blocks for Bob, you can use our bash shell function shown below:
```bash
source scripts/testlib.sh
mine-monero-for-swapd 5002
```
Now you can use the second to last command to see Bob's updated monero balance.

Next we need Bob to make an offer and advertise it, so that Alice can take it:
```bash
./swapcli make --min-amount 0.1 --max-amount 1 --exchange-rate 0.05 --swapd-port 5002
# Published offer with ID cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9
```

Alternatively, you can make the offer via websockets and get notified when the swap is taken:
```bash
./swapcli make --min-amount 0.1 --max-amount 1 --exchange-rate 0.05 --swapd-port 5002 --subscribe
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
./swapcli take --multiaddr /ip4/127.0.0.1/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7 --offer-id cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9 --provides-amount 0.05 --subscribe --swapd-port 5001
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
