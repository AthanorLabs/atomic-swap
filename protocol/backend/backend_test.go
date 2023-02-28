package backend

import (
	"context"
	"math/big"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/tests"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestWaitForReceipt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ec, chainID := tests.NewEthClient(t)
	privKey := tests.GetTakerTestKey(t)

	gasPrice, err := ec.SuggestGasPrice(ctx)
	require.NoError(t, err)

	to := ethcommon.Address{}
	txInner := &ethtypes.LegacyTx{
		To:       &to,
		Value:    big.NewInt(99),
		Gas:      21000,
		GasPrice: gasPrice,
	}

	tx, err := ethtypes.SignNewTx(privKey,
		ethtypes.LatestSignerForChainID(chainID),
		txInner,
	)
	require.NoError(t, err)

	err = ec.SendTransaction(ctx, tx)
	require.NoError(t, err)

	env := common.Development

	extendedEC, err := extethclient.NewEthClient(ctx, env, ec, privKey)
	require.NoError(t, err)

	b := &backend{
		ethClient: extendedEC,
	}

	receipt, err := b.ETHClient().WaitForReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	require.Equal(t, tx.Hash(), receipt.TxHash)
}
