#!/bin/bash

$SOLC_BIN --abi ethereum/contracts/Swap.sol -o ethereum/abi/ --overwrite
$SOLC_BIN --bin ethereum/contracts/Swap.sol -o ethereum/bin/ --overwrite
abigen --abi ethereum/abi/Swap.abi --pkg swap --type Swap --out swap.go --bin ethereum/bin/Swap.bin
mv swap.go ./swap-contract

$SOLC_BIN --abi ethereum/contracts/SwapDLEQ.sol -o ethereum/abi/ --overwrite
$SOLC_BIN --bin ethereum/contracts/SwapDLEQ.sol -o ethereum/bin/ --overwrite
abigen --abi ethereum/abi/SwapDLEQ.abi --pkg swapdleq --type SwapDLEQ --out swap_dleq.go --bin ethereum/bin/SwapDLEQ.bin
mv swap_dleq.go ./swap-dleq-contract
