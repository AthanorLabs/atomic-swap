# Developing 

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