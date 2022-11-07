package types

import (
	"encoding/hex"
	"encoding/json"
	"errors"

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

// Address ...
func (asset EthAsset) Address() ethcommon.Address {
	return ethcommon.Address(asset)
}

// MarshalJSON marshals the EthAsset as a 0x-prefixed hex string
func (asset EthAsset) MarshalJSON() ([]byte, error) {
	return json.Marshal(asset.Address())
}

// UnmarshalJSON unmarshals a 0x-prefixed hex string into an EthAsset
func (asset *EthAsset) UnmarshalJSON(data []byte) error {
	var hexStr string
	err := json.Unmarshal(data, &hexStr)
	if err != nil {
		return err
	}

	if len(hexStr) < 2 || hexStr[:2] != "0x" {
		return errors.New("invalid EthAsset string; must be prefixed with `0x`")
	}
	d, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		return err
	}

	copy(asset[:], d[:])
	return nil
}

// EthAssetETH describes regular ETH (rather than an ERC-20 token)
var EthAssetETH = EthAsset(ethcommon.HexToAddress("0x0"))
