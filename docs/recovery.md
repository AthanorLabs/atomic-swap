# Recovery

TODO: probably remove all this

In the case that the swap process crashes in the middle of the swap while funds are still locked, you can use the built-in recovery module to recover your funds manually.

## Building

To build the `swaprecover` binary, follow the instructions in [here](./build.md) but run `make build-all` instead of `make build`.

## Locating recovery file

Depending on whether you were on `dev`, `stagenet`, or `mainnet`, there will be a directory in your home directory `.atomicswap` that contains a directory of the network you were on.

Enter that directory and you should see a file named `info-<date-and-time>.txt`:

```bash
ls ~/.atomicswap/dev/
# info-2022-Apr-19-22:55:22.txt 
```

This file contains all the information you need to recover your funds. 

## Recovering as a maker

If you were in the role of maker during the swap, ie. you had XMR and were swapping for ETH, the following will allow you to either recover your XMR or claim the ETH.

Using the info file, you can recover your funds using the `swaprecover` binary.

For example, on the stagenet-Goerli networks:
```bash
./swaprecover --env stagenet --ethereum-endpoint=<your-goerli-endpoint> --ethereum-privkey=goerli.key --ethereum-chain-id=5 --infofile=/path/to/infofile --xmrmaker
```

The Ethereum private key must be the same one used when you ran `swapd`.

The recovery program will firstly try to claim ETH from the contract if possible. If the time to claim has already passed, and/or the counterparty has refunded the ETH to themselves, the program will refund the XMR to you by creating a Monero wallet containing the funds. If the program logs indicate a refund has occurred, please check your `monero-wallet-rpc` for the wallet containing your XMR.

## Recovering as a taker

If you were in the role of taker during the swap, ie. you had ETH and were swapping for XMR, the following will allow you to either recover your ETH or claim the XMR.

Using the info file, you can recover your funds using the `swaprecover` binary.

For example, on the stagenet-Goerli networks:
```bash
./swaprecover --env stagenet --ethereum-endpoint=<your-goerli-endpoint> --ethereum-privkey=goerli.key --ethereum-chain-id=5 --infofile=/path/to/infofile --xmrtaker
```

The Ethereum private key must be the same one used when you ran `swapd`.

The recovery program will firstly try to claim XMR by checking if the counterparty has claimed the ETH or not. If they haven't, the program will wait until the claim period finishes before trying to refund the ETH. If the program ends up refunding the ETH to you, it will end up back in your account specified by `--ethereum-privkey`. Otherwise, if the counterparty ends up claiming the ETH, you will receive the XMR in a new wallet inside `monero-wallet-rpc`.