package xmrmaker

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/tests"
)

func Test_waitForNewSwapReceipt(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	addr := crypto.PubkeyToAddress(pk.PublicKey)

	_, swapCreator := contracts.DevDeploySwapCreator(t, ec, pk)

	timeoutDuration := big.NewInt(10)
	value := big.NewInt(2e16)
	tx, err := swapCreator.NewSwap(
		tests.TxOptsWithValue(t, pk, value),
		[32]byte{1},
		[32]byte{2},
		addr,
		timeoutDuration,
		timeoutDuration,
		types.EthAssetETH.Address(),
		value,
		contracts.GenerateNewSwapNonce(),
	)
	require.NoError(t, err)

	// Simulate a maker's endpoint not being synchronized with the taker's
	// endpoint by calling waitForNewSwapReceipt without waiting for the
	// transaction to be mined into a block.

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(2 * time.Second)
		_ = tests.MineTransaction(t, ec, tx)
	}()

	receipt, err := waitForNewSwapReceipt(context.Background(), ec, tx.Hash())
	require.NoError(t, err)
	require.Equal(t, tx.Hash(), receipt.TxHash)
}

func Test_waitForNewSwapReceipt_reverted(t *testing.T) {
	ctx := context.Background()
	ec, chainID := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	addr := crypto.PubkeyToAddress(pk.PublicKey)

	swapCreatorAddr, _ := contracts.DevDeploySwapCreator(t, ec, pk)

	// Create a NewSwap transaction using lots of zero values that are checked in the contract
	// to trigger a revert
	callData, err := contracts.SwapCreatorParsedABI.Pack(
		"newSwap",
		[32]byte{0},
		[32]byte{0},
		addr,
		new(big.Int),
		new(big.Int),
		ethcommon.Address{},
		new(big.Int),
		new(big.Int),
	)
	require.NoError(t, err)

	nonce, err := ec.PendingNonceAt(context.Background(), addr)
	require.NoError(t, err)

	gasPrice, err := ec.SuggestGasPrice(context.Background())
	require.NoError(t, err)

	tx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      contracts.MaxNewSwapETHGas,
		To:       &swapCreatorAddr,
		Value:    nil,
		Data:     callData,
	})

	signedTx, err := ethtypes.SignTx(tx, ethtypes.NewEIP155Signer(chainID), pk)
	require.NoError(t, err)

	err = ec.SendTransaction(ctx, signedTx)
	require.NoError(t, err)

	_, err = waitForNewSwapReceipt(ctx, ec, signedTx.Hash())
	require.ErrorContains(t, err, fmt.Sprintf("received newSwap tx=%s was reverted", signedTx.Hash()))
}

func Test_waitForNewSwapReceipt_NotFound(t *testing.T) {
	ctx := context.Background()
	ec, _ := tests.NewEthClient(t)
	txHash := ethcommon.Hash{0x1, 0x2}

	// Requires a 15 second wait
	_, err := waitForNewSwapReceipt(ctx, ec, txHash)
	require.ErrorIs(t, err, ethereum.NotFound)
}
