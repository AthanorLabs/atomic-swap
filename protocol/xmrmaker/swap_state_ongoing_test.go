package xmrmaker

import (
	"math/big"
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/db"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func TestSwapStateOngoing_ClaimFunds(t *testing.T) {
	_, swapState := newTestSwapState(t)
	err := swapState.generateAndSetKeys()
	require.NoError(t, err)

	startNum, err := swapState.ETHClient().Raw().BlockNumber(swapState.Backend.Ctx())
	require.NoError(t, err)

	claimKey := swapState.secp256k1Pub.Keccak256()
	newSwap(t, swapState, claimKey,
		[32]byte{}, big.NewInt(33), defaultTimeoutDuration)
	swapState.cancel()

	txOpts, err := swapState.ETHClient().TxOpts(swapState.Backend.Ctx())
	require.NoError(t, err)
	tx, err := swapState.Contract().SetReady(txOpts, swapState.contractSwap)
	require.NoError(t, err)
	tests.MineTransaction(t, swapState.ETHClient().Raw(), tx)

	ethSwapInfo := &db.EthereumSwapInfo{
		StartNumber:     big.NewInt(int64(startNum)),
		SwapID:          swapState.contractSwapID,
		Swap:            swapState.contractSwap,
		ContractAddress: swapState.Backend.ContractAddr(),
	}

	swapState.info.Status = types.XMRLocked

	t.Log("creating swap state again...")
	ss, err := newSwapStateFromOngoing(
		swapState.Backend,
		swapState.offer,
		swapState.offerExtra,
		swapState.offerManager,
		ethSwapInfo,
		1,
		swapState.info,
		swapState.privkeys,
	)
	require.NoError(t, err)

	select {
	case <-ss.done:
	case <-time.After(time.Second * 10):
		t.Fatal("test timed out")
	}

	require.Equal(t, types.CompletedSuccess, swapState.info.Status)
}
