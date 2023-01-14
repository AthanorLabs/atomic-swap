# Trying the swap locally

### Requirements

#### Operating System
The code base is only tested regularly with Ubuntu 22.04 on X86-64, but we would like to
support most 64-bit Linux distributions, macOS, and WSL on Windows both with X86-64
and ARM processors.

#### Installed Dependencies for Building/Testing
- go 1.19+ (see [build instructions](./build.md) to download Go.)
- node/npm (to install ganache, see suggestions after list)
- ganache (can be installed with `npm install --location=global ganache`)
- jq, curl, bzip2, realpath

[The suggested way](https://github.com/nvm-sh/nvm#installing-and-updating) to install
node/npm is using nvm. If you install npm using a package manager like snap, ensure
the install prefix (`npm config get prefix`) is a directory that you have write
access to without sudo. You can change the directory with this command:
```
npm config set prefix ~/.npm-packages
```
See [this document](https://github.com/sindresorhus/guides/blob/main/npm-global-without-sudo.md)
if you want a more sophisticated setup.

If you are testing, you'll want `jq` installed. `curl` and `bzip2` are normally
preinstalled, but if you are running in a docker container, you might have to
install them.
```bash
sudo apt install curl bzip2 jq
```

#### Macos Notes
On macOS, you'll need to install `realpath` (from the `coreutils` package). If you
are using Homebrew, you can use these commands to install all the needed tools:
```bash
brew install coreutils go jq nvm
nvm install node
npm install --location=global ganache
```

### Set up development environment

Note: the `scripts/install-monero-linux.sh` script will download the monero binaries needed for you.
You can invoke it directly, but the next script below will run it if there is no symbolic link named
`monero-bin` to a monero installation in the project's root directory.

Use this command to launch ganache, an ethereum simulator, and monerod in regtest mode.
"regtest" mode is stand-alone (non-networked) mode of monerod for testing purposes.
Warning: the command below will kill running instances of `ganache`, `monerod`,
`monero-wallet-rpc` or `swapd`.
```bash
./scripts/setup-env.sh
```

To avoid confusion, delete any data directories from old runs of `swapd` that used
the flags `--dev-xmrtaker` or `--dev-xmrmaker`:
```bash
rm -rf "${TMPDIR:-/tmp}"/xmr[mt]aker-*
```

### Build the Executables

Build binary:
```bash
make build
```

This creates `swapd` and `swapcli` binaries in the `bin` directory at the top of the project.

### Launch Alice and Bob's swapd Instances

To launch Alice's swapd instance, use this command:
```bash
./bin/swapd --dev-xmrtaker --deploy &> alice.log &
```

We are going to use Alice's instance as a bootnode for Bob's instance. To configure this,
we need one of the multi-addresses that Alice is listening on. Alice will be listing both
on a TCP port and a UDP/QUIC port, it doesn't matter which one you pick. Use this command
to list Alice's listening libp2p multi-addresses:
```bash
./bin/swapcli addresses
```
Assign the value you picked to a variable. Your value will be different, this is just an
example below:
```bash
BOOT_NODE=/ip4/127.0.0.1/udp/9933/quic-v1/p2p/12D3KooWHRi24PVZ6TBnQJHdVyewDRcKFZtYV3qmB4KQo8iMyqik
```
Now get the ethereum contract address that Alice deployed to. This can be pulled from the Alice's logs,
the file ..., or if you have `jq` installed (available via `sudo apt install jq`), you can set a
variable like this:
```bash
CONTRACT_ADDR=$(jq -r .ContractAddress "${TMPDIR-/tmp}"/xmrtaker-*/contract-address.json)
```

Now start Bob's swapd instance:
```bash
./bin/swapd --dev-xmrmaker --bootnodes "${BOOT_NODE}" --contract-address "${CONTRACT_ADDR}" &> bob.log &
```

### Using swapcli To Check Balances

`swapcli` is an executable to interact with a `swapd` instance via it's RPC port on the
local host (`127.0.0.1`). The current security model of swapd assumes connections
originating from the local host are authorized, so you should not run `swapd` for
production swaps on multi-user hosts or hosts with malicious software running on them.

Note: when using the `--dev-xmrtaker` and `--dev-xmrmaker` flags, Alice's RPC server runs
on http://localhost:5000 (the default port) and Bob's runs on http://localhost:5001. Since
Bob's `swapd` RPC port is not the default, you will need to pass `--swapd-port 5001` to
`swapcli` when interacting with his daemon.

Alice and Bob are both using Ethereum wallet keys that are prefunded by Ganache.
Background Monero mining was started for Bob, because his swapd instance used the
`--dev-xmrmaker` flag.

You can check Alice's balance with this command:
```bash
./bin/swapcli balances
```
And Bob's balances with this command:
```bash
./bin/swapcli balances --swapd-port 5001
```

### Make a Swap Offer

Next we need Bob to make an offer and advertise it, so that Alice can take it:
```bash
./bin/swapcli make --min-amount 0.1 --max-amount 1 --exchange-rate 0.05 --swapd-port 5001
```
Example output:
```
Published:
	Offer ID:  0x09dd41c7b8620cdc3716463dc947a11edf3af45ff07c8b0ff89dd23592e732ca
	Peer ID:   12D3KooWK7989g6xmAaEsKFPuZTj2CVknRxQuk7dFL55CC1rpEWW
	Taker Min: 0.005 ETH
	Taker Max: 0.05 ETH
```

Alternatively, you can make the offer via websockets and get notified when the swap is
taken. This option will block waiting for update messages, so you will need to dedicate a
separate terminal for it:
```bash
./bin/swapcli make --min-amount 0.1 --max-amount 1 --exchange-rate 0.05 --swapd-port 5001 --subscribe
```

### Discover Swap Offers

Now, Alice can discover peers who have advertised offers.
```bash
./bin/swapcli discover --provides XMR --search-time 3
```
```
Peer 0: 12D3KooWAE3zH374qcxyFCA8B5g1uMqhgeiHoXT5KKD6A54SGGsp
```

Query the returned peer to find the range of XMR they are willing to swap and at what exchange rate.
Note: You need to update the peer ID below with the one from your output in the previous step.
```bash
./bin/swapcli query --peer-id 12D3KooWAE3zH374qcxyFCA8B5g1uMqhgeiHoXT5KKD6A54SGGsp
```
```
Offer ID: 0xcc57d3d1b9d8186118f1f1581a8dc4dca0e5aa6c39a5255bd0c2ebb824cfe2eb
Provides: XMR
Min Amount: 0.1
Max Amount: 1
Exchange Rate: 0.05
ETH Asset: ETH
```

### Take a Swap Offers

Alice now has the information needed to start a swap with Bob. You'll need Bob's peer ID and his offer ID
from the previous step to update the command below:
```bash
./bin/swapcli take \
  --peer-id 12D3KooWAE3zH374qcxyFCA8B5g1uMqhgeiHoXT5KKD6A54SGGsp \
  --offer-id 0xcc57d3d1b9d8186118f1f1581a8dc4dca0e5aa6c39a5255bd0c2ebb824cfe2eb \
  --provides-amount 0.05
```

Alternatively, you can take the offer via websockets and get notified when the swap status updates.
```bash
./bin/swapcli take \
  --peer-id 12D3KooWAE3zH374qcxyFCA8B5g1uMqhgeiHoXT5KKD6A54SGGsp \
  --offer-id 0xcc57d3d1b9d8186118f1f1581a8dc4dca0e5aa6c39a5255bd0c2ebb824cfe2eb \
  --provides-amount 0.05 --subscribe
```
```
Initiated swap with offer ID 0xcc57d3d1b9d8186118f1f1581a8dc4dca0e5aa6c39a5255bd0c2ebb824cfe2eb
> Stage updated: ExpectingKeys
> Stage updated: ETHLocked
> Stage updated: ContractReady
> Stage updated: Success
```

If all goes well, you should see Alice and Bob successfully exchange messages and execute
the swap protocol.

### Other swapcli Commands

To query the information for an ongoing swap, you can run:
```bash
./bin/swapcli get-ongoing-swap --offer-id <id>
```

To query information for a past swap using its ID, you can run:
```bash
./bin/swapcli get-past-swap --offer-id <id>
```
