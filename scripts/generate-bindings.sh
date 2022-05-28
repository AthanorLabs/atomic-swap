#!/bin/bash

$SOLC_BIN --abi ethereum/contracts/SwapFactory.sol -o ethereum/abi/ --overwrite
$SOLC_BIN --bin ethereum/contracts/SwapFactory.sol -o ethereum/bin/ --overwrite

# this requires geth v1.10.17 or lower
abigen --sol ethereum/contracts/SwapFactory.sol --pkg swapfactory --out swap_factory.go 
mv swap_factory.go ./swapfactory
