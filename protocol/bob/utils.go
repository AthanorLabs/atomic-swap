package bob

import (
	"bytes"
	"context"

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
