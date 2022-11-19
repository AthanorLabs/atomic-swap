package protocol

import (
	"fmt"
	"strings"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/protocol/backend"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const etherSymbol = "ETH"

// ConvertContractSwapToMsg converts a contracts.SwapFactorySwap to a *message.ContractSwap
func ConvertContractSwapToMsg(swap contracts.SwapFactorySwap) *message.ContractSwap {
	return &message.ContractSwap{
		Owner:        swap.Owner,
		Claimer:      swap.Claimer,
		PubKeyClaim:  swap.PubKeyClaim,
		PubKeyRefund: swap.PubKeyRefund,
		Timeout0:     swap.Timeout0,
		Timeout1:     swap.Timeout1,
		Asset:        swap.Asset,
		Value:        swap.Value,
		Nonce:        swap.Nonce,
	}
}

// AssetSymbol returns the symbol for the given asset.
func AssetSymbol(b backend.Backend, asset types.EthAsset) (string, error) {
	if asset != types.EthAssetETH {
		_, symbol, _, err := b.ETHClient().ERC20Info(b.Ctx(), asset.Address())
		if err != nil {
			return "", fmt.Errorf("failed to get ERC20 info: %w", err)
		}

		return symbol, nil
	}

	return etherSymbol, nil
}

// CheckSwapID checks if the given log is for the given swap ID.
func CheckSwapID(log *ethtypes.Log, eventName string, contractSwapID types.Hash) error {
	abiSF, err := abi.JSON(strings.NewReader(contracts.SwapFactoryMetaData.ABI))
	if err != nil {
		return err
	}

	data := log.Data
	res, err := abiSF.Unpack(eventName, data)
	if err != nil {
		return err
	}

	if len(res) < 1 {
		return errLogMissingParams
	}

	swapID := res[0].([32]byte)
	if swapID != contractSwapID {
		return ErrLogNotForUs
	}

	return nil
}
