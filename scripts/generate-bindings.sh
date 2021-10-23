#!/bin/bash

SOLC_BIN="/home/elizabeth/Downloads/solc-static-linux"

$SOLC_BIN --abi ethereum/contracts/Swap.sol -o ethereum/abi/ --overwrite
$SOLC_BIN --bin ethereum/contracts/Swap.sol -o ethereum/bin/ --overwrite
abigen --abi ethereum/abi/Swap.abi --pkg swap --type Swap --out swap.go --bin ethereum/bin/Swap.bin
mv swap.go ./swap-contract
