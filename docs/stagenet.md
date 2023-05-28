# Joining the Stagenet/Sepolia network

Swaps can be performed on the Sepolia (Ethereum testnet) together with the
Monero Stagenet network. This document describes how to do stagenet swaps.

> Note: stagenet and mainnet swaps currently takes around 20 minutes due to the 2-minute block time and required 10 confirmations before funds can be spent.

> Note: the `swapd` process directly interacts with an unlocked Monero wallet and Ethereum private key. This is to allow for a smoother swap process that doesn't require any interaction from you once initiated. However, this effectively gives `swapd` access to all your (testnet) funds. In the future, there will be a mode that does not access your keys/wallet, but will require user interaction during a swap.

## Setup 

The atomic swap daemon requires access to a fully synced, stagenet monerod daemon,
a Sepolia network endpoint, and a Sepolia network private key funded with some SepETH.

1. Install the Monero CLI if you haven't already. You can get it [here](https://www.getmonero.org/downloads/#cli):

For Linux 64-bit, you can do:
```bash
./scripts/install-monero-linux.sh
```

2. Begin the stagenet daemon and wait for it to sync. This may take a day or so. Alternatively, you can use an existing stagenet endpoint if you know of one. **You can find remote Monero nodes here: https://monero.fail/?nettype=stagenet**

```bash
./monero-bin/monerod --detach --stagenet
```

3. Optional: Place your stagenet monero wallet file in `{DATA_DIR}/wallet/` and call it
   `swap-wallet`. By default `{DATA_DIR}` is `${HOME}/.atomicswap/stagenet`, but if you
   are creating multiple stagenet `swapd` instances on the same host, you should pass
   `swapd` the `--data-dir` flag so each instance has a separate directory to store its
   data. If you skip this step, a new wallet will be created that you can later fund for
   swaps.

4. Create a Sepolia wallet. You can do this using Metamask by selecting "Sepolia Test Network" from the networks, then creating a new account with "Create account". I'd recommend naming this new account something explicit like `sepolia-swap-account`.

5. Optional: Export the private key for this account by navigating to: three dots in upper right of
   Metamask -> account details -> export private key. Paste this private key into a file
   named `{DATA_DIR}/eth.key`. If you skip this step, a new wallet will be created for you
   that you can transfer Sepolia ether to or fund directly in the next step.

6a. Fund your Sepolia account using a faucet:
- https://sepolia-faucet.pk910.de/
- https://sepoliafaucet.com/
- https://sepolia.dev/

6b. Optional: Obtain some Sepolia ERC20 tokens

This [unaffiliated project](https://github.com/bokkypoobah/WeenusTokenFaucet/blob/master/README.md)
has deployed some Sepolia tokens of different decimal sizes that can be useful
for testing. You can use MetaMask to send the contract address zero Sepolia ETH
and the contract will grant you 1000 of its ERC20 tokens. You will pay gas fees
that you should validate are sane before sending.

7. Obtain a Sepolia JSON-RPC endpoint. If you don't want to sync your own node, you can find public ones here: https://sepolia.dev/

8. Install go 1.20+. See [build instructions](./build.md) for more details.

9. Clone and build the atomic-swap binaries:
```bash
git clone https://github.com/athanorlabs/atomic-swap.git
cd atomic-swap
make build
```

10. Start the `swapd` daemon. Change `--eth-endpoint` to point to your endpoint.
```bash
./bin/swapd --env stagenet --eth-endpoint SEPOLIA_ENDPOINT
```
Note: You probably need additional flags above:
* `--data-dir PATH`: Needed if you are launching more than one `swapd` instance
  on the same host, otherwise accepting the default of `${HOME}/.atomicswap/stagenet`
  is fine.
* `--monerod-host HOSTNAME_OR_IP` and `--monerod-port PORT_NUM`: Ideally, you have your
  own stagenet node on the local network and will use these values. If that is not an
  option, our stagenet default uses `node.sethforprivacy.com:38089`.
* `--libp2p-port PORT`. The default is `9900`. Use this flag when creating multiple
  swapd instances on the same host.
* `--rpc-port PORT`. The default is `5000`. Use this flag when creating multiple
  swapd instances on the same host.

> Note: please also see the [RPC documentation](./rpc.md) for complete documentation on available RPC calls and their parameters.

## Taker 

As a taker, you can use either the UI or `swapcli` to discover and take offers.

### UI

**WARNING: The UI is currently unmaintained and probably does not work; please use the CLI to take offers.**

1. From the `atomic-swap` directory, build and start the UI. Note: you need to have node.js installed.
```bash
cd ui/
yarn install
yarn build
yarn start
```

2. Navigate to http://localhost:8080 to see the UI running. It will automatically connect to your `swapd` process and try to find offers. You can also refresh the offers by clicking `refresh`.

![ui](./images/ui.png)

3. When you find an offer you'd like to take, press the `take` button to input the amount of ETH you'd like to provide. Then, confirm the offer. If all goes well, you should see the swap complete in the logs of `swapd`.

![ui](./images/ui-take.png)
![ui](./images/ui-swapping.png)

### CLI

1. Search for existing XMR offers using `swapcli`:
```bash
./bin/swapcli discover --provides XMR --search-time 3 --swapd-port 5001
# [[/ip4/127.0.0.1/udp/9934/quic-v1/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv /ip4/127.0.0.1/udp/9934/quic-v1/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv]]
```

2. Query a returned peer as to how much XMR they can provide and their preferred exchange rate (replace `"--peer-id"` field with one of the addresses returned in the above step):
```bash
./bin/swapcli query --peer-id 12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv
# Offer ID=cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9 Provides=XMR MinimumAmount=0.1 MaximumAmount=1 ExchangeRate=0.05
```

> Note: the exchange rate is the ratio of XMR:ETH price. So for example, a ratio of 0.05 would mean 20 XMR to 1 ETH. Since we're on testnet, it's not critical what you set it to. 

3. a. Then, finding an offer you like, take the offer by copying the peer's multiaddress and offer ID into the command below. As well, specify how much SepETH you would like to provide, taking into account the offer's exchange rate and min/max XMR amounts.
```bash
./bin/swapcli take --peer-id 12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv \
  --offer-id cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9 --provides-amount 0.05
# Initiated swap with ID=0
```

3. b. Alternatively, you can take the offer without getting notified of swap status updates:
```bash
./bin/swapcli take --peer-id 12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7 \
  --offer-id 0xcf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9 \
  --provides-amount 0.05 --detached --swapd-port 5001
```

If all goes well, you should see the node execute the swap protocol. If the swap ends successfully, the XMR received will automatically be transferred back your original wallet.

## Maker
 
1. Find your stagenet address:
```bash
./bin/swapcli balances | grep 'Monero address'
```

2. Fund this address with some stagenet XMR. You can try the faucets here:
- https://stagenet-faucet.xmr-tw.org/
- https://community.rino.io/faucet/stagenet/

If you don't have any luck with these, please message me on twitter/reddit (@elizabethereum) with your stagenet address, and I can send you some stagenet XMR.

3. a. Make an offer with `swapcli`:
```bash
./bin/swapcli make --min-amount 0.1 --max-amount 1 --exchange-rate 0.5 --swapd-port 5001
# Published offer with ID cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9
```

4. b. Alternatively, make an offer with `swapcli` without subscribing to updates:
```bash
./bin/swapcli make --min-amount 0.1 --max-amount 1 --exchange-rate 0.5 --swapd-port 5001 --detached
# Published offer with ID cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9
```

> Note: the exchange rate is the ratio of XMR:ETH price. So for example, a ratio of 0.05 would mean 20 XMR to 1 ETH. Since we're on testnet, it's not critical what you set it to. 

When a peer takes your offer, you will see logs in `swapd` notifying you that a swap has been initiated. If all goes well, you should receive the SepETH in the Sepolia account created earlier.

## Troubleshooting

Ideally, the exit case of the swap should be `Success`. If this is not the case, it will either be one of `Refunded` or `Aborted`.
- `Refunded` means that the swap refunded after your funds were already locked. In this case, you would lose  transaction fees.
- `Aborted` means that the swap exited before any funds were locked, so nothing was lost except time.

Neither of these should happen, so if they happen, it indicates an issue either on your side or the remote peer's side.

A few common errors are:
- `Failed to get height`: double check that your `monerod --stagenet` process is running.
- `unlocked balance is less than maximum offer amount`: you will see this if you're a maker and try to make an offer but don't have enough balance. Either get more stagenet XMR or wait for your balance to unlock.

## Trying the swap on a different network

You can also try the swap on another Ethereum or EVM-compatible testnet. However, you'll need to run your own maker nodes. 

To connect to a different Ethereum network, follow [Setup](#setup) steps 4-7 but with your desired network. Then, start `swapd` with your specified private key file, endpoint, and chain ID. Common chain IDs can be found [here](https://besu.hyperledger.org/en/stable/Concepts/NetworkID-And-ChainID/).

> Note: The `--deploy` flag to `swapd` creates a new instance of `SwapCreator.sol` to the
network. You need to have funds in your account to deploy the contract. To use a contract
deployed with `--deploy` in subsequent `swapd` instances, use the flag
`--contract-addr=ADDRESS`. When using `stagenet`, a deployed contract already exists and
our code will use it by default.

## Bug reports

If you find any bugs or unexpected swap occurrences, please [open an issue](https://github.com/athanorlabs/atomic-swap/issues/new) on the repo, detailing exact steps you took to setup `swapd` and what caused the bug to occur. Your OS and environment would be helpful as well. Any bug reports or general improvement suggestions are much appreciated.
