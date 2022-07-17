#!/bin/bash

if [[ -z "${SOLC_BIN}" ]]; then
	SOLC_BIN=solc
fi

"${SOLC_BIN}" --abi ethereum/contracts/SwapFactory.sol -o ethereum/abi/ --overwrite
"${SOLC_BIN}" --bin ethereum/contracts/SwapFactory.sol -o ethereum/bin/ --overwrite

# this requires geth v1.10.17 or lower
abigen --sol ethereum/contracts/SwapFactory.sol --pkg swapfactory --out swapfactory/swap_factory.go
