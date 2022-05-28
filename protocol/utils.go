package protocol

import (
	"fmt"
	"time"

	"github.com/noot/atomic-swap/net/message"
	"github.com/noot/atomic-swap/swapfactory"
)

// GetSwapInfoFilepath returns an info file path with the current timestamp.
func GetSwapInfoFilepath(basepath string) string {
	t := time.Now().Format("2006-Jan-2-15:04:05")
	path := fmt.Sprintf("%s/info-%s.txt", basepath, t)
	return path
}

// GetSwapRecoveryFilepath returns an info file path with the current timestamp.
func GetSwapRecoveryFilepath(basepath string) string {
	t := time.Now().Format("2006-Jan-2-15:04:05")
	path := fmt.Sprintf("%s/recovery-%s.txt", basepath, t)
	return path
}

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
