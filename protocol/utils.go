package protocol

import (
	"fmt"
	"path"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
)

const etherSymbol = "ETH"

// GetSwapInfoFilepath returns an info file path with the current timestamp.
func GetSwapInfoFilepath(dataDir string, offerID string) string {
	t := time.Now().Format(common.TimeFmtNSecs)
	return path.Join(dataDir, fmt.Sprintf("info-%s-%s", t, offerID))
}

// GetSwapRecoveryFilepath returns an info file path with the current timestamp.
func GetSwapRecoveryFilepath(dataDir string) string {
	t := time.Now().Format(common.TimeFmtNSecs)
	return path.Join(dataDir, fmt.Sprintf("recovery-%s.json", t))
}

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
		_, symbol, _, err := b.ERC20Info(b.Ctx(), asset.Address())
		if err != nil {
			return "", fmt.Errorf("failed to get ERC20 info: %w", err)
		}

		return symbol, nil
	}

	return etherSymbol, nil
}
