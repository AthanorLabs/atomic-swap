# Mainnet swap usage

The swap operates on a maker/taker paradigm, where one party is a "market maker" and publishes offers for swaps they wish to execute, and the other party is a "taker" which takes a published offer, fulfilling the swap counterparty role.

Currently, due to protocol limitations, XMR holders who want ETH can only act as makers, and ETH holders who want XMR can only act as takers.

> **Note:** a swap on mainnet currently takes around 20-25 minutes due to block time.

> **Note:** the `swapd` process directly interacts with an unlocked Monero wallet and Ethereum private key. This is to allow for a smoother swap process that doesn't require any interaction from you once initiated. However, this effectively gives `swapd` access to all funds in those accounts. It's recommended to only keep funds that will be used for a swap in those accounts. In the future, there will be a mode that does not access your keys/wallet, but will require user interaction during a swap.

## Contents

* [Build](#Build)
* [Swap daemon setup](#Swap-daemon-setup)
* [Relayer](#Relayer)
* [swapcli commands](#swapcli-commands)
* [Monero taker](#Monero-taker)
* [Monero maker](#Monero-maker)
* [Troubleshooting](#Troubleshooting)

## Build

See the [build instructions](./build.md) to build the swapd and swapcli programs.

Alternatively, you can install the pre-built binaries provided with the [releases](https://github.com/AthanorLabs/atomic-swap/releases). After downloading the `swapd` and `swapcli` binaries for your machine, you can install them in your PATH. For example, on Linux:
```bash
sudo mv ~/Downloads/swapd-linux-x64 /usr/local/bin/swapd && sudo chmod +x /usr/local/bin/swapd
```

## Swap daemon setup 

The atomic swap daemon requires access to a fully synced Monero client and Ethereum client. It's recommended to run your own node, but you can also use a remote node.

1. Install the Monero CLI if you haven't already. You can get it [here](https://www.getmonero.org/downloads/#cli):

For Linux 64-bit, you can do:
```bash
./scripts/install-monero-linux.sh
```

2. Run `monerod` and wait for it to sync. This may take a few days. Alternatively, you can use an existing Monero endpoint if you know of one. **You can find remote Monero nodes here: https://monero.fail/?nettype=mainnet**

```bash
./monero-bin/monerod --detach 
```

3. Optional: If you have a Monero wallet file you wish to use, place it in `{DATA_DIR}/wallet/` and call it
   `swap-wallet`. By default `{DATA_DIR}` is `${HOME}/.atomicswap/mainnet`, but if you
   are creating multiple `swapd` instances on the same host, you should pass
   `swapd` the `--data-dir` flag so each instance has a separate directory to store its
   data. If you skip this step, a new wallet will be created that you can later fund for
   swaps.

4. Obtain an Ethereum JSON-RPC endpoint.

5. Start the `swapd` daemon. Change `--eth-endpoint` to point to your endpoint.
```bash
./bin/swapd --eth-endpoint MAINNET_ENDPOINT
```

If you did not provide a Monero wallet file with `--wallet-file` above, a Monero wallet file is generated for you at `${HOME}/.atomicswap/mainnet/wallet/swap-wallet`. You can pass this to `monero-wallet-cli --wallet-file FILE` to interact with it. **The wallet password is empty by default.**

Note: You may need additional flags above:
* `--eth-privkey`: Path to a file containing an Ethereum private key (hex string). If you want to act as an XMR-taker (ETH provider), `swapd` needs access to a funded account. If you do not provide a key with this flag, you should transfer funds to the address logged when the node starts up.
* `--data-dir PATH`: Needed if you are launching more than one `swapd` instance
  on the same host, otherwise accepting the default of `${HOME}/.atomicswap/mainnet`
  is fine.
* `--monerod-host HOSTNAME_OR_IP` and `--monerod-port PORT_NUM`: Ideally, you have your
  own node on the local network and will use these values. If that is not an
  option, our default uses `node.sethforprivacy.com`.
* `--libp2p-port PORT`. The default is `9900`. Use this flag when creating multiple
  swapd instances on the same host.
* `--rpc-port PORT`. The default is `5000`. Use this flag when creating multiple
  swapd instances on the same host.
* `--log-level LEVEL`. If you want to see debug logs, you can set `LEVEL` to `debug`. If you want less logs, you can set it to `warn` or `error`.

> Note: please also see the [RPC documentation](./rpc.md) for complete documentation on available RPC calls and their parameters.

## Relayer
 
The Ethereum network requires that users have ether in an account to be able to execute any transactions from that account. For ETH-takers, this means that they would need to have an already-funded account to claim their swap funds. However, this is not ideal for privacy. A workaround is to have users relay transactions on behalf of others, meaning that the relayer would pay the gas fee for the swap claim transaction and receive a small portion of the funds in return.

To run a node as a relayer, pass the `--relayer` flag to `swapd`:
```bash
./bin/swapd --eth-endpoint MAINNET_ENDPOINT --relayer
```

**Note:** the current fee sent to relayers is 0.01 ETH per swap. Subtract the gas cost from this to determine how much profit will be made. The gas required to do a relayer-claim transaction is `85040` gas. Multiply this by the transaction gas price for the gas cost. The gas price is set via oracle unless you manually set it with the `personal_setGasPrice` RPC call.

## swapcli commands

`swapcli` is used to interact with `swapd`, ie. for finding peers and offers on the network and making/taking swaps.

Some useful commands are:
* `swapcli balances`: check your ETH and XMR addresses and balances.
* `swapcli ongoing`: check the status of all ongoing swaps.
* `swapcli past`: see all your past swaps.
* `swapcli get-offers`: see all your currently advertised offers.

You can see all available commands with `swapcli -h`.

## Monero Taker 

1. Check your Ethereum address and balance and ensure your address is funded:
```bash
./bin/swapcli balances
```

2. Search for existing XMR offers using `swapcli`:
```bash
./bin/swapcli discover --provides XMR --search-time 3 --swapd-port 5001
# [[/ip4/127.0.0.1/udp/9934/quic-v1/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv /ip4/127.0.0.1/udp/9934/quic-v1/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv]]
```

3. Query a returned peer as to how much XMR they can provide and their preferred exchange rate (replace `"--peer-id"` field with one of the addresses returned in the above step):
```bash
./bin/swapcli query --peer-id 12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv
# Offer ID=cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9 Provides=XMR MinimumAmount=0.1 MaximumAmount=1 ExchangeRate=0.05
```

> **Note:** the exchange rate is the ratio of XMR:ETH price. So for example, a ratio of 0.05 would mean 20 XMR to 1 ETH. 

> **Note:** the XMR-maker's offer may have an `EthAsset` set, meaning they wish to swap for an ERC20 token, not ETH. In this case, your account must be funded with that token to be able to take the offer.

4. a. Then, finding an offer you like, take the offer by copying the peer's multiaddress and offer ID into the command below. As well, specify how much ETH you would like to provide, taking into account the offer's exchange rate and min/max XMR amounts.
```bash
./bin/swapcli take --peer-id 12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv \
  --offer-id cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9 --provides-amount 0.05
# Initiated swap with ID=0
```

This will automatically provide you with pushed status updates. `CTRL+C` will stop the status updates, but does not stop the swap, so feel free to exit. 

5. b. Alternatively, you can take the offer without getting notified of swap status updates:
```bash
./bin/swapcli take --peer-id 12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7 \
  --offer-id 0xcf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9 \
  --provides-amount 0.05 --detached --swapd-port 5001
```

Either way, you can always check the swap's status via `./bin/swapcli ongoing`. 

If all goes well, you should see the node execute the swap protocol. If the swap ends successfully, the XMR received will automatically be transferred back your original wallet.

## Monero Maker

1. Check your Monero address and balance and ensure your address is funded:
```bash
./bin/swapcli balances
```

2. a. Make an offer with `swapcli`:
```bash
./bin/swapcli make --min-amount MIN-XMR-AMOUNT --max-amount MAX-XMR-AMOUNT --exchange-rate EXCHANGE-RATE
# Published offer with ID cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9
```

This will automatically provide you with pushed status updates when the swap is taken. `CTRL+C` will stop the status updates, but does not stop the swap or remove the offer, so feel free to exit. 

> **Note:** the exchange rate is the ratio of XMR:ETH price. So for example, a ratio of 0.05 would mean 20 XMR to 1 ETH. You can see a suggested exchange rate from the Chainlink oracle using `swapcli suggested-exchange-rate`; however, you should always double check this against your own sources.

> **Note:** if you wish to swap for an ERC20 instead of ETH, you can set the asset with `--eth-asset TOKEN-CONTRACT-ADDRESS`. However, you must have a funded ETH account to perform a swap for an ERC20, as relayers are not supported for token swaps.

3. b. Alternatively, make an offer with `swapcli` without subscribing to updates:
```bash
./bin/swapcli make --min-amount MIN-XMR-AMOUNT --max-amount MAX-XMR-AMOUNT --exchange-rate EXCHANGE-RATE --detached
# Published offer with ID cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9
```

When a peer takes your offer, you will see logs in `swapd` notifying you that a swap has been initiated. If all goes well, you'll receive the ETH in the account used by `swapd` and success logs in `swapd`. You can always check the swap's status via `swapcli ongoing` or `swapcli past`.

## Troubleshooting

Ideally, the exit case of the swap should be `Success`. If this is not the case, it will either be one of `Refunded` or `Aborted`.
- `Refunded` means that the swap refunded after your funds were already locked. In this case, you would lose transaction fees only.
- `Aborted` means that the swap exited before any funds were locked, so nothing was lost except time.

Neither of these should happen, so if they happen, it indicates an issue either on your side or the remote peer's side.

A few common errors are:
- `Failed to get height`: double check that your `monerod` process is running.
- `unlocked balance is less than maximum offer amount`: you will see this if you're a maker and try to make an offer but don't have enough balance. Either fund your account with more XMR or wait for your balance to unlock. This can also happen if you make a transfer out of your swap wallet and a majority of your funds end up in a change output waiting for confirmations.
- A bad Ethereum endpoint. If you're using a remote endpoint and it goes down, or you run out of requests, the swap daemon will not be able to make progress. You will probably see some Ethereum-related error logs in this case. Get a new endpoint and restart the swap daemon with it. Your swap progress will not be lost.

## Bug reports

If you find any bugs or unexpected swap occurrences, please [open an issue](https://github.com/athanorlabs/atomic-swap/issues/new) on the repo, detailing exact steps you took to setup `swapd` and what caused the bug to occur. Please mention the commit used. Your OS and environment would be helpful as well. Any bug reports or general improvement suggestions are much appreciated.
