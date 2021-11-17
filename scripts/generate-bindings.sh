#!/bin/bash

$SOLC_BIN --abi ethereum/contracts/SwapOnChain.sol -o ethereum/abi/ --overwrite
$SOLC_BIN --bin ethereum/contracts/SwapOnChain.sol -o ethereum/bin/ --overwrite
abigen --abi ethereum/abi/SwapOnChain.abi --pkg swap --type Swap --out swap.go --bin ethereum/bin/SwapOnChain.bin
mv swap.go ./swap-contract

$SOLC_BIN --abi ethereum/contracts/SwapDLEQ.sol -o ethereum/abi/ --overwrite
$SOLC_BIN --bin ethereum/contracts/SwapDLEQ.sol -o ethereum/bin/ --overwrite
abigen --abi ethereum/abi/SwapDLEQ.abi --pkg swapdleq --type SwapDLEQ --out swap_dleq.go --bin ethereum/bin/SwapDLEQ.bin
mv swap_dleq.go ./swap-dleq-contract
