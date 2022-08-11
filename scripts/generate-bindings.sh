#!/bin/bash

# Use the project root (one directory above this script) as the current working directory:
cd "$(dirname "$(readlink -f "$0")")/.."

ABIGEN="$(go env GOPATH)/bin/abigen"

if [[ -z "${SOLC_BIN}" ]]; then
	SOLC_BIN=solc
fi

"${SOLC_BIN}" --abi ethereum/contracts/SwapFactory.sol -o ethereum/abi/ --overwrite
"${SOLC_BIN}" --bin ethereum/contracts/SwapFactory.sol -o ethereum/bin/ --overwrite

"${ABIGEN}" --abi ethereum/abi/SwapFactory.abi \
            --bin ethereum/bin/SwapFactory.bin \
            --pkg swapfactory \
            --type SwapFactory \
            --out swapfactory/swap_factory.go
