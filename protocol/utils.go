package protocol

import (
	"fmt"
	"path"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
)

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
