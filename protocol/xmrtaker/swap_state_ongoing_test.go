package xmrtaker

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/db"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
)

func setupSwapStateUntilETHLocked(t *testing.T) (*swapState, uint64) {
	s := newTestSwapState(t)
	defer s.cancel()
	s.SetSwapTimeout(time.Minute * 2)

	rdb := s.Backend.RecoveryDB().(*backend.MockRecoveryDB)

	startNum, err := s.ETHClient().Raw().BlockNumber(s.Backend.Ctx())
	require.NoError(t, err)

	makerKeys, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)
	s.setXMRMakerKeys(
		makerKeys.PublicKeyPair.SpendKey(),
		makerKeys.PrivateKeyPair.ViewKey(),
		makerKeys.Secp256k1PublicKey,
	)

	_, err = s.lockAsset()
	require.NoError(t, err)

	s.info.Status = types.ETHLocked

	// shutdown swap state, re-create from ongoing
	s.cancel()

	rdb.EXPECT().GetXMRMakerSwapKeys(s.info.ID).Return(
		makerKeys.PublicKeyPair.SpendKey(),
		makerKeys.PrivateKeyPair.ViewKey(),
		nil,
	)
	return s, startNum
}

// tests the case where the swap is stopped at the stage where the next
// expected event is EventXMRLockedType.
// in this case, an EventXMRLocked should cause the contract to be set to ready.
func TestSwapStateOngoing_handleEvent_EventXMRLocked(t *testing.T) {
	s, startNum := setupSwapStateUntilETHLocked(t)

	ethInfo := &db.EthereumSwapInfo{
		StartNumber:     big.NewInt(int64(startNum)),
		SwapID:          s.contractSwapID,
		Swap:            s.contractSwap,
		ContractAddress: s.Backend.ContractAddr(),
	}

	ss, err := newSwapStateFromOngoing(
		s.Backend,
		s.info,
		s.transferBack,
		ethInfo,
		s.privkeys,
	)
	require.NoError(t, err)
	require.Equal(t, EventXMRLockedType, ss.nextExpectedEvent)

	xmrAddr, _ := ss.expectedXMRLockAccount()
	lockXMRAndCheckForReadyLog(t, ss, xmrAddr)
}

// tests the case where the swap is stopped at the stage where the next
// expected event is EventETHClaimedType.
// in this case, an EventETHClaimed should allow the swap to complete successfully.
func TestSwapStateOngoing_handleEvent_EventETHClaimed(t *testing.T) {
	s, startNum := setupSwapStateUntilETHLocked(t)

	ethInfo := &db.EthereumSwapInfo{
		StartNumber:     big.NewInt(int64(startNum)),
		SwapID:          s.contractSwapID,
		Swap:            s.contractSwap,
		ContractAddress: s.Backend.ContractAddr(),
	}

	ss, err := newSwapStateFromOngoing(
		s.Backend,
		s.info,
		s.transferBack,
		ethInfo,
		s.privkeys,
	)
	require.NoError(t, err)
	require.Equal(t, EventXMRLockedType, ss.nextExpectedEvent)

	// simulate xmrmaker calling claim
	secret := s.privkeys.SpendKeyBytes()
	sk, err := mcrypto.NewPrivateSpendKey(secret[:])
	require.NoError(t, err)
	ss.nextExpectedEvent = EventETHClaimedType

	// handled the claimed message should result in the monero wallet being created
	event := newEventETHClaimed(sk)
	ss.eventCh <- event
	err = <-event.errCh
	require.NoError(t, err)
	require.Equal(t, types.CompletedSuccess, ss.info.Status)
}
