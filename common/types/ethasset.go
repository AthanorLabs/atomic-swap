package types

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
)

// EthAsset represents an Ethereum asset (ETH or a token address)
type EthAsset ethcommon.Address

func (asset EthAsset) String() string {
	if ethcommon.Address(asset).Hex() == "0x0000000000000000000000000000000000000000" {
		return "ETH"
	}
	return ethcommon.Address(asset).Hex()
}

// TODO: is there a cleaner way to create a constant like this?
var EthAssetETH = EthAsset(ethcommon.HexToAddress("0x0"))
