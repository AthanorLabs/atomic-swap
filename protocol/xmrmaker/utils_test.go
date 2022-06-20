package xmrmaker

import (
	"context"
	"math/big"
	"testing"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/swapfactory"
	"github.com/noot/atomic-swap/tests"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCheckContractCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	b := NewMockBackend(ctrl)

	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	pk, err := ethcrypto.HexToECDSA(tests.GetMakerTestKey(t))
	require.NoError(t, err)

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(common.GanacheChainID))
	require.NoError(t, err)

	addr, _, _, err := swapfactory.DeploySwapFactory(txOpts, ec)
	require.NoError(t, err)

	b.EXPECT().CodeAt(context.Background(), addr, nil).
		DoAndReturn(func(ctx context.Context, account ethcommon.Address, _ *big.Int) ([]byte, error) {
			return ec.CodeAt(ctx, account, nil)
		})

	err = checkContractCode(context.Background(), b, addr)
	require.NoError(t, err)
}
