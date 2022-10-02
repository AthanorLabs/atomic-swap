package xmrmaker

import (
	"bytes"
	"context"

	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/protocol/backend"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

func checkContractCode(ctx context.Context, b backend.Backend, contractAddr ethcommon.Address) error {
	code, err := b.CodeAt(ctx, contractAddr, nil)
	if err != nil {
		return err
	}

	expectedCode := ethcommon.FromHex(contracts.SwapFactoryMetaData.Bin)
	if !bytes.Contains(expectedCode, code) {
		return errInvalidSwapContract
	}

	return nil
}

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
