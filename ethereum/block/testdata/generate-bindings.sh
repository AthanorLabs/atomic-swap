#!/bin/bash
# Generate the UTContract.sol Go bindings into a file named ut_contract_test.go of the
# parent directory for use by unit tests.

# Use the contract's directory (where this script is) as the current working directory:
cd "$(dirname "$(readlink -f "$0")")"

ABIGEN="$(go env GOPATH)/bin/abigen"

if [[ -z "${SOLC_BIN}" ]]; then
	SOLC_BIN=solc
fi

"${SOLC_BIN}" --abi  UTContract.sol -o . --overwrite
"${SOLC_BIN}" --bin  UTContract.sol -o . --overwrite

# Use abigen 1.10.17-stable to match how we compile the SwapFactory contract
"${ABIGEN}" --abi UTContract.abi --bin UTContract.bin --pkg block --type UTContract --out ../ut_contract_test.go

rm -f UTContract.abi UTContract.bin
