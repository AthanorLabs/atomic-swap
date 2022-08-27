package xmrmaker

import (
	"bytes"
	"context"

	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/swapfactory"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

func checkContractCode(ctx context.Context, b backend.Backend, contractAddr ethcommon.Address) error {
	code, err := b.CodeAt(ctx, contractAddr, nil)
	if err != nil {
		return err
	}

	expectedCode := ethcommon.FromHex(swapfactory.SwapFactoryMetaData.Bin)
	if !bytes.Contains(expectedCode, code) {
		return errInvalidSwapContract
	}

	return nil
}

func convertContractSwap(msg *message.ContractSwap) swapfactory.SwapFactorySwap {
	return swapfactory.SwapFactorySwap{
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
