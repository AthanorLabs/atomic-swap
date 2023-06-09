#!/usr/bin/env bash

set -e

# Use the project root (one directory above this script) as the current working directory:
PROJECT_ROOT="$(dirname "$(dirname "$(realpath "$0")")")"
cd "${PROJECT_ROOT}"

ABIGEN="$(go env GOPATH)/bin/abigen"

if [[ -z "${SOLC_BIN}" ]]; then
	SOLC_BIN=solc
fi

compile-contract() {
	local solidity_file_name="${1:?}"
	local go_type_name="${2:?}"
	local go_file_name="${3:?}"

	# strip leading path and extension from to get the solidity type name
	local solidity_type_name
	solidity_type_name="$(basename "${solidity_file_name%.sol}")"

	echo "Generating go bindings for ${solidity_type_name}"

	"${SOLC_BIN}" --optimize --optimize-runs=200 \
		--base-path "ethereum/contracts" \
		--abi "ethereum/contracts/${solidity_file_name}" \
		-o ethereum/abi/ --overwrite
	"${SOLC_BIN}" --optimize --optimize-runs=200 \
		--base-path ethereum/contracts \
		--include-path . \
		--bin "ethereum/contracts/${solidity_file_name}" \
		-o ethereum/bin/ --overwrite

	"${ABIGEN}" \
		--abi "ethereum/abi/${solidity_type_name}.abi" \
		--bin "ethereum/bin/${solidity_type_name}.bin" \
		--pkg contracts \
		--type "${go_type_name}" \
		--out "ethereum/${go_file_name}.go"
}

compile-contract SwapCreator.sol SwapCreator swap_creator
compile-contract TestERC20.sol TestERC20 erc20_token
compile-contract @openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol IERC20 ierc20
compile-contract AggregatorV3Interface.sol AggregatorV3Interface aggregator_v3_interface
