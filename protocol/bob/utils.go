package bob

import (
	"bytes"
	"context"

	"github.com/noot/atomic-swap/net/message"
	"github.com/noot/atomic-swap/swapfactory"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func checkContractCode(ctx context.Context, ec *ethclient.Client, contractAddr ethcommon.Address) error {
	code, err := ec.CodeAt(ctx, contractAddr, nil)
	if err != nil {
		return err
	}

	expectedCode := ethcommon.FromHex(swapfactory.SwapFactoryBin)
	if !bytes.Equal(expectedCode, code) {
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
		Value:        msg.Value,
		Nonce:        msg.Nonce,
	}
}
