package bob

import (
	"bytes"
	"context"

	"github.com/noot/atomic-swap/net/message"
	"github.com/noot/atomic-swap/swapfactory"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TODO: redo this (how tf do i not hardcode this)
func checkContractCode(ctx context.Context, ec *ethclient.Client, contractAddr ethcommon.Address) error {
	return nil

	code, err := ec.CodeAt(ctx, contractAddr, nil)
	if err != nil {
		return err
	}

	expectedCode := ethcommon.FromHex(swapfactory.SwapFactoryBin)

	// compiled bytecode isn't the exact same as deployed bycode, need to compare subsets
	// one subset is Swap.sol, one is Secp256k1.sol
	if !bytes.Equal(expectedCode[154:3474], code[:3320]) {
		return errInvalidSwapContract
	}

	if !bytes.Equal(expectedCode[3494:6149], code[3340:]) {
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
