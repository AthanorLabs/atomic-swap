#!/bin/bash

$SOLC_BIN --abi ethereum/contracts/SwapOnChain.sol -o ethereum/abi/ --overwrite
$SOLC_BIN --bin ethereum/contracts/SwapOnChain.sol -o ethereum/bin/ --overwrite
abigen --abi ethereum/abi/SwapOnChain.abi --pkg swapOnChain --type SwapOnChain --out swapOnChain.go --bin ethereum/bin/SwapOnChain.bin
mv swapOnChain.go ./swapOnChain-contract
