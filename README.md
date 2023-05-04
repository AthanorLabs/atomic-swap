# ETH-XMR Atomic Swaps

This is an implementation of ETH-XMR atomic swaps, currently in beta. It currently consists of `swapd` and `swapcli` binaries, the swap daemon and swap CLI tool respectively, which allow for nodes to discover each other over the p2p network, to query nodes for their current available offers, and the ability to make and take swap offers and perform the swap protocol. The `swapd` program has a JSON-RPC endpoint which the user can use to interact with it. `swapcli` is a command-line utility that interacts with `swapd` by performing RPC calls. 

## Swap instructions

### Trying it on mainnet

To try the swap on Ethereum and Monero mainnet, follow the instructions [here](./docs/mainnet.md).

### Trying it on Monero's stagenet and Ethereum's Sepolia testnet

To try the swap on Stagenet/Sepolia, follow the instructions [here](./docs/stagenet.md).

### Trying it locally

To try the swap locally with two nodes (maker and taker) on a development environment, follow the instructions [here](./docs/local.md).

## Protocol

Please see the [protocol documentation](docs/protocol.md) for how it works.

## Additional documentation

### Developer instructions

Please see the [developer docs](docs/developing.md).

### RPC API

The swap process comes with a HTTP JSON-RPC API as well as a Websockets API. You can find the documentation [here](./docs/rpc.md).

## Contributions

If you'd like to contribute, feel free to fork the repo and make a pull request. Please make sure the CI is passing - you can run `make build`, `make lint`, and `make test` to make sure the checks pass locally. Please note that any contributions you make will be licensed under LGPLv3.

## Contact
 
- [Matrix room](https://matrix.to/#/#ethxmrswap:matrix.org)

## Donations

The work on this project has been funded previously by community grants. It is currently not funded; if you'd like to donate, you can do so at the following address:
- XMR `8AYdE4Tzq3rQYh7QNHfHz8HqcgT9kcTcHMcRHL1LhVtqYwah27zwPYGdesBgK5PATvGBAd4BC1t2NfrqKQqDguybQrC1tZb`
- ETH `0x39D3b8cc9D08fD83360dDaCFe054b7D6e7f2cA08`

## GPLv3 Disclaimer 

THERE IS NO WARRANTY FOR THE PROGRAM, TO THE EXTENT PERMITTED BY APPLICABLE LAW. EXCEPT WHEN OTHERWISE STATED IN WRITING THE COPYRIGHT HOLDERS AND/OR OTHER PARTIES PROVIDE THE PROGRAM “AS IS” WITHOUT WARRANTY OF ANY KIND, EITHER EXPRESSED OR IMPLIED, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE. THE ENTIRE RISK AS TO THE QUALITY AND PERFORMANCE OF THE PROGRAM IS WITH YOU. SHOULD THE PROGRAM PROVE DEFECTIVE, YOU ASSUME THE COST OF ALL NECESSARY SERVICING, REPAIR OR CORRECTION.

IN NO EVENT UNLESS REQUIRED BY APPLICABLE LAW OR AGREED TO IN WRITING WILL ANY COPYRIGHT HOLDER, OR ANY OTHER PARTY WHO MODIFIES AND/OR CONVEYS THE PROGRAM AS PERMITTED ABOVE, BE LIABLE TO YOU FOR DAMAGES, INCLUDING ANY GENERAL, SPECIAL, INCIDENTAL OR CONSEQUENTIAL DAMAGES ARISING OUT OF THE USE OR INABILITY TO USE THE PROGRAM (INCLUDING BUT NOT LIMITED TO LOSS OF DATA OR DATA BEING RENDERED INACCURATE OR LOSSES SUSTAINED BY YOU OR THIRD PARTIES OR A FAILURE OF THE PROGRAM TO OPERATE WITH ANY OTHER PROGRAMS), EVEN IF SUCH HOLDER OR OTHER PARTY HAS BEEN ADVISED OF THE POSSIBILITY OF SUCH DAMAGES.