# Developing 

## Building the project

Follow the [build instructions](./build.md) to ensure you have Go installed and can build the project.

## Setting up your local environment

Follow the instructions [here](local.md) to set up your local Alice (ETH-holder, XMR-wanter) and Bob (XMR-holder, ETH-wanter) nodes. 

You can use the script `scripts/setup-env.sh` to quickly set up local monerod-test and
ganache instances. If you need to later kill `ganache`, `monerod`, `monero-wallet-rpc`, or
`swapd` instances, you can use `scripts/cleanup-test-processes.sh`.

## Deploying or using deployed SwapCreator.sol

The swap program uses a "creator" contract for the Ethereum side to reduce gas
costs from deploying a new contract for each swap. The contract can be found in
[here](../ethereum/contracts/SwapCreator.sol). For each new swap, the
eth-holding party will call `NewSwap` on the creator contract, initiating a swap
instance inside the contract.

If you're developing on a local network, running a `swapd` instance with the
`--dev-xmrtaker` flag will automatically deploy an instance of `SwapCreator.sol`
for you. You should see the following log shortly after starting
`./swapd --dev-xmrtaker`:
```bash
# 2022-01-26T18:39:04.600-0500	INFO	cmd	daemon/contract.go:35	deployed SwapCreator.sol: address=0x3F2aF34E4250de94242Ac2B8A38550fd4503696d tx hash=0x638caf280178b3cfe06854b8a76a4ce355d38c5d81187836f0733cad1287b657
```

If you wish to use an instance of `SwapCreator.sol` that's already deployed
on-chain, you can use the `--contract-address` flag to specify the address. For
example:
```bash
$ ./swapd --dev-xmrtaker --contract-address 0x3F2aF34E4250de94242Ac2B8A38550fd4503696d
# 2022-01-26T18:56:31.627-0500	INFO	cmd	daemon/contract.go:42	loaded SwapCreator.sol from address 0x3F2aF34E4250de94242Ac2B8A38550fd4503696d
```

## Compiling contract bindings

If you update the `Swap.sol` contract for some reason, you will need to re-generate the Go bindings
for the contract. **Note:** you do *not* need to do this to try out the swap; only if you want to
edit the contract for development purposes.

Download solc v0.8.19: https://github.com/ethereum/solidity/releases/tag/v0.8.19

If `solc` with the needed version is not in your path (or not first in your path), set the
`SOLC_BIN` environment variable to the correct version:
```
export SOLC_BIN=solc
```

We install the `abigen` into the `bin` directory of your GOPATH (`$HOME/go/bin` for most users).
The version installed is matched to the go-ethereum version that the project currently links with.
See `scripts/install-abigen.sh` for details.

Generate the bindings
```
make bindings
```

## Testing
To setup the test environment and run all unit tests, execute:
```
make test
```

This includes tests for main protocol functionality, such as:
1. Success case, where both parties obey the protocol
2. Case where Bob never locks monero on his side. Alice can Refund
3. Case where Bob locks monero, but never claims his ether from the contract

You can also run 
```
make test-integration
```

to run integration tests which spin up 3 local nodes and execute calls between them.

## Mocks

The unit tests use mocks. You need to install mockgen to generate new mocks:
```bash
go install github.com/golang/mock/mockgen@v1.6.0
```

Then, if you update an interface, generate new mocks using:
```bash
go generate -run mockgen ./...
```

## Linting and Formatting

There are `make format` and `make lint` targets to format and check code for
errors. These targets require two programs that can be installed using `apt`
on Ubuntu for Bash formatting and linting:
```
sudo apt install -y shfmt shellcheck
```
The solidity formatting uses an npm executable named `prettier` which, for solidity,
needs a plugin called `prettier-plugin-solidity`. For solidity linting, we use the
`solhint` program.
```
npm install --global --save-dev prettier prettier-plugin-solidity
npm install --global --save-dev solhint
```

Go linting uses `golangci-lint`. If it is not already installed in your user's
GOBIN directory, the `make lint` command will install it for you using
`scripts/install-lint.sh`.
