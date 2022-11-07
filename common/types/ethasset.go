package types

import (
	"fmt"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// EthAsset represents an Ethereum asset (ETH or a token address)
type EthAsset ethcommon.Address

// String ...
func (asset EthAsset) String() string {
	if ethcommon.Address(asset).Hex() == "0x0000000000000000000000000000000000000000" {
		return "ETH"
	}
	// TODO: get name of asset from contract?
	return ethcommon.Address(asset).Hex()
}

// MarshalText returns the hex representation of the EthAsset
func (asset EthAsset) MarshalText() ([]byte, error) {
	return []byte(asset.String()), nil
}

// UnmarshalText assigns the EthAsset from the input text
func (asset *EthAsset) UnmarshalText(input []byte) error {
	inputStr := string(input)
	switch {
	case strings.EqualFold(inputStr, "ETH"):
		*asset = EthAsset{}
		return nil
	case ethcommon.IsHexAddress(inputStr):
		*asset = EthAsset(ethcommon.HexToAddress(inputStr))
		return nil
	default:
		return fmt.Errorf("invalid asset value %q", inputStr)
	}
}

// Address ...
func (asset EthAsset) Address() ethcommon.Address {
	return ethcommon.Address(asset)
}

// EthAssetETH describes regular ETH (rather than an ERC-20 token)
var EthAssetETH = EthAsset(ethcommon.HexToAddress("0x0"))
