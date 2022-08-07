# Developing 

## Building the project

Follow the [build instructions](./build.md) to ensure you have Go installed and can build the project.

## Setting up your local environment

Follow the instructions [here](local.md) to set up your local Alice (ETH-holder, XMR-wanter) and Bob (XMR-holder, ETH-wanter) nodes. 

You can also run `bash scripts/setup-env.sh` to quickly set up the local Monero and ganache environment.

## Deploying or using deployed SwapFactory.sol

The swap program uses a "factory" contract for the Ethereum side to reduce gas costs from deploying a new contract for each swap. The contract can be found in [here](../ethereum/contracts/SwapFactory.sol). For each new swap, the eth-holding party will call `NewSwap` on the factory contract, initiating a swap instance inside the contract.

If you're developing on a local network, running a `swapd` instance with the `--dev-xmrtaker` flag will automatically deploy an instance of `SwapFactory.sol` for you. You should see the following log shortly after starting `./swapd --dev-xmrtaker`:
```bash
# 2022-01-26T18:39:04.600-0500	INFO	cmd	daemon/contract.go:35	deployed SwapFactory.sol: address=0x3F2aF34E4250de94242Ac2B8A38550fd4503696d tx hash=0x638caf280178b3cfe06854b8a76a4ce355d38c5d81187836f0733cad1287b657
```

If you wish to use an instance of `SwapFactory.sol` that's already deployed on-chain, you can use the `--contract-address` flag to specify the address. For example:
```bash
$ ./swapd --dev-xmrtaker --contract-address 0x3F2aF34E4250de94242Ac2B8A38550fd4503696d
# 2022-01-26T18:56:31.627-0500	INFO	cmd	daemon/contract.go:42	loaded SwapFactory.sol from address 0x3F2aF34E4250de94242Ac2B8A38550fd4503696d
```

If you want to deploy the contract without running `swapd`, you can use hardhat. You will need node.js installed.
```bash
cd ethereum
npm install --save-dev hardhat
```

Then, you can run the deployment script against your desired network to deploy the contract.

```bash
$ npx hardhat run --network localhost scripts/deploy.js 
# Compiling 2 files with 0.8.5
# Compilation finished successfully
# SwapFactory deployed to: 0xB0f6EC177Ab867E372479C0bfdBB0068bb5E1554
```

## Compiling DLEq binaries

The program utilizes a Rust DLEq library implemented by Farcaster.

To compile the farcaster-dleq binaries used, you can run:
```
make build-dleq
```

This will install Rust (if it isn't already installed) and build the binaries. The resulting binaries will be in `./farcaster-dleq/target/release/`.

## Compiling contract bindings

If you update the `Swap.sol` contract for some reason, you will need to re-generate the Go bindings for the contract. **Note:** you do *not* need to do this to try out the swap; only if you want to edit the contract for development purposes.

Download solc v0.8.9: https://github.com/ethereum/solidity/releases/tag/v0.8.9

Set `SOLC_BIN` to the downloaded binary
```
export SOLC_BIN=solc
```

Install `abigen`
```
git clone https://github.com/ethereum/go-ethereum.git && cd go-ethereum/cmd/abigen
go install
```

Generate the bindings
```
./scripts/generate-bindings.sh
```
Note: you may need to add `$GOPATH` and `$GOPATH/bin` to your path.

## Testing
To setup the test environment and run all unit tests, execute:
```
make test
```

This include tests for main protocol functionality, such as:
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
