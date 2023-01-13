#!/usr/bin/env bash

# Use the project root (one directory above this script) as the current working directory:
PROJECT_ROOT="$(dirname "$(dirname "$(realpath "$0")")")"
cd "${PROJECT_ROOT}" || exit 1

ABIGEN="$(go env GOPATH)/bin/abigen"

if [[ -z "${SOLC_BIN}" ]]; then
	SOLC_BIN=solc
fi

compile-contract() {
	local solidity_type_name="${1:?}"
	local go_type_name="${2:?}"
	local go_file_name="${3:?}"

	echo "Generating go bindings for ${solidity_type_name}"

	"${SOLC_BIN}" --abi "ethereum/contracts/${solidity_type_name}.sol" -o ethereum/abi/ --overwrite
	"${SOLC_BIN}" --bin "ethereum/contracts/${solidity_type_name}.sol" -o ethereum/bin/ --overwrite
	"${ABIGEN}" \
		--abi "ethereum/abi/${solidity_type_name}.abi" \
		--bin "ethereum/bin/${solidity_type_name}.bin" \
		--pkg contracts \
		--type "${go_type_name}" \
		--out "ethereum/${go_file_name}.go"
}

compile-contract SwapFactory SwapFactory swap_factory
compile-contract ERC20Mock ERC20Mock erc20_mock
compile-contract IERC20Metadata IERC20 ierc20
compile-contract AggregatorV3Interface AggregatorV3Interface aggregator_v3_interface
