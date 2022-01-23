#!/bin/bash

$SOLC_BIN --abi ethereum/contracts/Swap.sol -o ethereum/abi/ --overwrite
$SOLC_BIN --bin ethereum/contracts/Swap.sol -o ethereum/bin/ --overwrite
abigen --abi ethereum/abi/Swap.abi --pkg swap --type Swap --out swap.go --bin ethereum/bin/Swap.bin
mv swap.go ./swap-contract


$SOLC_BIN --abi ethereum/contracts/SwapFactory.sol -o ethereum/abi/ --overwrite
$SOLC_BIN --bin ethereum/contracts/SwapFactory.sol -o ethereum/bin/ --overwrite
abigen --abi ethereum/abi/SwapFactory.abi --pkg swapfactory --type SwapFactory --out swap_factory.go --bin ethereum/bin/SwapFactory.bin
mv swap_factory.go ./swapfactory
