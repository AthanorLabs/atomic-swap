#!/bin/bash

# Generate the contract Go bindings into a file named ut_contract_test.go of the
# parent directory for use by unit tests.

# Use the contract's directory as the current working directory
cd "$(dirname "$(readlink -f "$0")")"

if [[ -z "${SOLC_BIN}" ]]; then
	SOLC_BIN=solc
fi

"${SOLC_BIN}" --abi  UTContract.sol -o . --overwrite
"${SOLC_BIN}" --bin  UTContract.sol -o . --overwrite

# Use abigen 1.10.17-stable to match how we compile the SwapFactory contract
abigen --sol UTContract.sol --pkg block --out ../ut_contract_test.go

rm -f UTContract.abi UTContract.bin
