package protocol

import (
	"fmt"
	"path"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/net/message"
	"github.com/noot/atomic-swap/swapfactory"
)

// GetSwapInfoFilepath returns an info file path with the current timestamp.
func GetSwapInfoFilepath(basePath string) string {
	t := time.Now().Format(common.TimeFmtNSecs)
	return path.Join(basePath, t)
}

// GetSwapRecoveryFilepath returns an info file path with the current timestamp.
func GetSwapRecoveryFilepath(basePath string) string {
	t := time.Now().Format(common.TimeFmtNSecs)
	return path.Join(basePath, fmt.Sprintf("recovery-%s.txt", t))
}

// ConvertContractSwapToMsg converts a swapfactory.SwapFactorySwap to a *message.ContractSwap
func ConvertContractSwapToMsg(swap swapfactory.SwapFactorySwap) *message.ContractSwap {
	return &message.ContractSwap{
		Owner:        swap.Owner,
		Claimer:      swap.Claimer,
		PubKeyClaim:  swap.PubKeyClaim,
		PubKeyRefund: swap.PubKeyRefund,
		Timeout0:     swap.Timeout0,
		Timeout1:     swap.Timeout1,
		Value:        swap.Value,
		Nonce:        swap.Nonce,
	}
}
