# ETH-XMR Atomic Swaps

This is a WIP implementation of ETH-XMR atomic swaps, currently in the pre-production development phase. It currently consists of `swapd` and `swapcli` binaries, the swap daemon and swap CLI tool respectively, which allow for peers to discover each other over the network, query peers for their current available offers, and the ability to make and take swap offers and perform the swap protocol. The `swapd` program has a JSON-RPC endpoint which the user can use to interact with it. `swapcli` is a command-line utility that interacts with `swapd` by performing RPC calls. 

## Disclaimer

**This code is unaudited and under active development and should not be used on mainnet or any networks that hold monetary value!!!**

## Protocol

Please see the [protocol documentation](docs/protocol.md) for how it works.

## Swap instructions

### Trying it locally

To try the swap locally with two nodes (maker and taker) on a development environment, follow the instructions [here](./docs/local.md).

### Trying it on Monero's stagenet and Ethereum's Sepolia testnet

To try the swap on Stagenet/Sepolia, follow the instructions [here](./docs/stagenet.md).

## Additional documentation

### Developer instructions

Please see the [developer docs](docs/developing.md).

### RPC API

The swap process comes with a HTTP JSON-RPC API as well as a Websockets API. You can find the documentation [here](./docs/rpc.md).

## Contributions

If you'd like to contribute, feel free to fork the repo and make a pull request. Please make sure the CI is passing - you can run `make build`, `make lint`, and `make test` to make sure the checks pass locally.

## Contact
 
- [Matrix room](https://matrix.to/#/#ethxmrswap:matrix.org)

## Donations

The work on this project has been funded previously by community grants. It is currently not funded; if you'd like to donate, you can do so at the following address:
- XMR `8AYdE4Tzq3rQYh7QNHfHz8HqcgT9kcTcHMcRHL1LhVtqYwah27zwPYGdesBgK5PATvGBAd4BC1t2NfrqKQqDguybQrC1tZb`
- ETH `0x39D3b8cc9D08fD83360dDaCFe054b7D6e7f2cA08`
