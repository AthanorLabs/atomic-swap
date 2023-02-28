package xmrmaker

import (
	"math/big"
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/db"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func TestSwapStateOngoing_ClaimFunds(t *testing.T) {
	_, swapState := newTestSwapState(t)

	startNum, err := swapState.ETHClient().Raw().BlockNumber(swapState.Backend.Ctx())
	require.NoError(t, err)

	claimKey := swapState.secp256k1Pub.Keccak256()
	newSwap(t, swapState, claimKey,
		[32]byte{}, big.NewInt(33), defaultTimeoutDuration)
	swapState.cancel()

	txOpts, err := swapState.ETHClient().TxOpts(swapState.Backend.Ctx())
	require.NoError(t, err)
	tx, err := swapState.Contract().SetReady(txOpts, *swapState.contractSwap)
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

func TestSwapStateOngoing_Refund(t *testing.T) {
	inst, s, offerDB := newTestSwapStateAndDB(t)
	offerDB.EXPECT().PutOffer(s.offer)

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	err = s.setXMRTakerKeys(
		xmrtakerKeysAndProof.PublicKeyPair.SpendKey(),
		xmrtakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrtakerKeysAndProof.Secp256k1PublicKey,
	)
	require.NoError(t, err)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	startNum, err := s.ETHClient().Raw().BlockNumber(s.Backend.Ctx())
	require.NoError(t, err)

	refundKey := xmrtakerKeysAndProof.Secp256k1PublicKey.Keccak256()
	newSwap(t, s, [32]byte{}, refundKey, desiredAmount.BigInt(), duration)

	// lock XMR
	_, err = s.lockFunds(coins.MoneroToPiconero(s.info.ProvidedAmount))
	require.NoError(t, err)
	s.cancel()

	// call refund w/ XMRTaker's spend key
	secret := xmrtakerKeysAndProof.PrivateKeyPair.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	ctx := s.Backend.Ctx()
	txOpts, err := s.ETHClient().TxOpts(ctx)
	require.NoError(t, err)
	tx, err := s.Contract().Refund(txOpts, *s.contractSwap, sc)
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(ctx, s.ETHClient().Raw(), tx.Hash())
	require.NoError(t, err)
	require.Equal(t, 1, len(receipt.Logs))

	ethSwapInfo := &db.EthereumSwapInfo{
		StartNumber:     big.NewInt(int64(startNum)),
		SwapID:          s.contractSwapID,
		Swap:            s.contractSwap,
		ContractAddress: s.Backend.ContractAddr(),
	}

	s.info.Status = types.XMRLocked
	rdb := inst.backend.RecoveryDB().(*backend.MockRecoveryDB)
	rdb.EXPECT().GetCounterpartySwapKeys(s.ID()).Return(
		xmrtakerKeysAndProof.PublicKeyPair.SpendKey(),
		xmrtakerKeysAndProof.PrivateKeyPair.ViewKey(),
		nil,
	)

	t.Log("creating swap state again...")
	ss, err := newSwapStateFromOngoing(
		s.Backend,
		s.offer,
		s.offerExtra,
		s.offerManager,
		ethSwapInfo,
		s.info,
		s.privkeys,
	)
	require.NoError(t, err)

	select {
	case <-ss.done:
	case <-time.After(time.Second * 10):
		t.Fatal("test timed out")
	}

	require.Equal(t, types.CompletedRefund, ss.info.Status)
}
