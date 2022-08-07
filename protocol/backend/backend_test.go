package backend

import (
	"context"
	"math/big"
	"testing"

	"github.com/noot/atomic-swap/tests"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestWaitForReceipt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ec, chainID := tests.NewEthClient(t)
	privKey := tests.GetTakerTestKey(t)

	to := ethcommon.Address{}
	txInner := &ethtypes.LegacyTx{
		To:       &to,
		Value:    big.NewInt(99),
		Gas:      21000,
		GasPrice: big.NewInt(2000000000),
	}

	tx, err := ethtypes.SignNewTx(privKey,
		ethtypes.LatestSignerForChainID(chainID),
		txInner,
	)
	require.NoError(t, err)

	err = ec.SendTransaction(ctx, tx)
	require.NoError(t, err)

	b := &backend{
		ethClient: ec,
	}

	receipt, err := b.WaitForReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	require.Equal(t, tx.Hash(), receipt.TxHash)
}
