package xmrmaker

import (
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
)

func convertContractSwap(msg *message.ContractSwap) contracts.SwapFactorySwap {
	return contracts.SwapFactorySwap{
		Owner:        msg.Owner,
		Claimer:      msg.Claimer,
		PubKeyClaim:  msg.PubKeyClaim,
		PubKeyRefund: msg.PubKeyRefund,
		Timeout0:     msg.Timeout0,
		Timeout1:     msg.Timeout1,
		Asset:        msg.Asset,
		Value:        msg.Value,
		Nonce:        msg.Nonce,
	}
}
