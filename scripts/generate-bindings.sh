#!/bin/bash

SOLC_BIN="/home/elizabeth/Downloads/solc-static-linux"

$SOLC_BIN --abi contracts/contracts/Swap.sol -o contracts/abi/ --overwrite
$SOLC_BIN --bin contracts/contracts/Swap.sol -o contracts/bin/ --overwrite
abigen --abi contracts/abi/Swap.abi --pkg swap --type Swap --out swap.go --bin contracts/bin/Swap.bin
mv swap.go ./swap-contract