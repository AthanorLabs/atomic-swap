package common

import (
	"context"
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

const defaultAliceAddress = "0x90F8bf6A479f320ead074411a4B0e7944Ea8c9C1"

func TestWaitForReceipt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ec, err := ethclient.Dial(DefaultEthEndpoint)
	require.NoError(t, err)

	pk, err := ethcrypto.HexToECDSA(DefaultPrivKeyAlice)
	require.NoError(t, err)

	nonce, err := ec.PendingNonceAt(ctx, ethcommon.HexToAddress(defaultAliceAddress))
	require.NoError(t, err)

	to := ethcommon.Address{}
	txInner := &ethtypes.LegacyTx{
		Nonce: nonce,
		To:    &to,
		Value: big.NewInt(99),
		Gas:   21000,
	}

	tx, err := ethtypes.SignNewTx(pk,
		ethtypes.LatestSignerForChainID(big.NewInt(GanacheChainID)),
		txInner,
	)
	require.NoError(t, err)

	err = ec.SendTransaction(ctx, tx)
	require.NoError(t, err)

	receipt, err := WaitForReceipt(ctx, ec, tx.Hash())
	require.NoError(t, err)
	require.Equal(t, tx.Hash(), receipt.TxHash)
}
