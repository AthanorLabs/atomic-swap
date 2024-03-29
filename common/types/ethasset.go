// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package types

import (
	"fmt"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// EthAsset represents an Ethereum asset (ETH or a token address)
type EthAsset ethcommon.Address

// IsETH returns true of the asset is ETH, otherwise false
func (asset EthAsset) IsETH() bool {
	return asset == EthAssetETH
}

// IsToken returns true if the asset is an ERC20 token, otherwise false
func (asset EthAsset) IsToken() bool {
	return asset != EthAssetETH
}

// String implements fmt.Stringer, returning the asset's address in hex
// prefixed by `ERC20@` if it's an ERC20 token, or ETH for ether.
func (asset EthAsset) String() string {
	if asset.IsETH() {
		return "ETH"
	}

	// TODO: get name of asset from contract?
	return fmt.Sprintf("ERC20@%s", ethcommon.Address(asset).Hex())
}

// MarshalText returns the hex representation of the EthAsset or,
// in some cases, a short string.
func (asset EthAsset) MarshalText() ([]byte, error) {
	if asset.IsETH() {
		return []byte("ETH"), nil
	}
	return []byte(ethcommon.Address(asset).Hex()), nil
}

// UnmarshalText assigns the EthAsset from the input text
func (asset *EthAsset) UnmarshalText(input []byte) error {
	inputStr := string(input)
	if inputStr == "ETH" {
		*asset = EthAssetETH
		return nil
	}

	inputStr = strings.TrimPrefix(inputStr, "ERC20@")
	if ethcommon.IsHexAddress(inputStr) {
		*asset = EthAsset(ethcommon.HexToAddress(inputStr))
		return nil
	}

	return fmt.Errorf("invalid asset value %q", inputStr)
}

// Address ...
func (asset EthAsset) Address() ethcommon.Address {
	return ethcommon.Address(asset)
}

// EthAssetETH describes regular ETH (rather than an ERC-20 token)
var EthAssetETH = EthAsset(ethcommon.Address{})
