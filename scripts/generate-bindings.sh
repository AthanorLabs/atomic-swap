#!/bin/bash

# Use the project root (one directory above this script) as the current working directory:
PROJECT_ROOT="$(dirname "$(dirname "$(readlink -f "$0")")")"
cd "${PROJECT_ROOT}" || exit 1

ABIGEN="$(go env GOPATH)/bin/abigen"

if [[ -z "${SOLC_BIN}" ]]; then
	SOLC_BIN=solc
fi

"${SOLC_BIN}" --abi ethereum/contracts/SwapFactory.sol -o ethereum/abi/ --overwrite
"${SOLC_BIN}" --bin ethereum/contracts/SwapFactory.sol -o ethereum/bin/ --overwrite

"${ABIGEN}" \
	--abi ethereum/abi/SwapFactory.abi \
	--bin ethereum/bin/SwapFactory.bin \
	--pkg swapfactory \
	--type SwapFactory \
	--out swapfactory/swap_factory.go

"${SOLC_BIN}" --abi ethereum/contracts/ERC20Mock.sol -o ethereum/abi/ --overwrite
"${SOLC_BIN}" --bin ethereum/contracts/ERC20Mock.sol -o ethereum/bin/ --overwrite

"${ABIGEN}" \
	--abi ethereum/abi/ERC20Mock.abi \
	--bin ethereum/bin/ERC20Mock.bin \
	--pkg swapfactory \
	--type ERC20Mock \
	--out swapfactory/erc20_mock.go

"${SOLC_BIN}" --abi ethereum/contracts/IERC20Metadata.sol -o ethereum/abi/ --overwrite
"${ABIGEN}" \
	--abi ethereum/abi/IERC20Metadata.abi \
	--pkg swapfactory \
	--type IERC20 \
	--out swapfactory/ierc20.go