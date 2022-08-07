package xmrmaker

import (
	"context"
	"math/big"
	"testing"

	"github.com/noot/atomic-swap/swapfactory"
	"github.com/noot/atomic-swap/tests"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCheckContractCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	b := NewMockBackend(ctrl)

	ec, chainID := tests.NewEthClient(t)
	ctx := context.Background()
	pk := tests.GetMakerTestKey(t)

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)

	_, tx, _, err := swapfactory.DeploySwapFactory(txOpts, ec)
	require.NoError(t, err)

	addr, err := bind.WaitDeployed(ctx, ec, tx)
	require.NoError(t, err)

	b.EXPECT().CodeAt(context.Background(), addr, nil).
		DoAndReturn(func(ctx context.Context, account ethcommon.Address, _ *big.Int) ([]byte, error) {
			return ec.CodeAt(ctx, account, nil)
		})

	err = checkContractCode(ctx, b, addr)
	require.NoError(t, err)
}
