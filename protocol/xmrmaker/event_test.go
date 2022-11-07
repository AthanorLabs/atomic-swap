package xmrmaker

import (
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func TestSwapState_handleEvent_EventContractReady(t *testing.T) {
	_, s := newTestInstance(t)

	s.nextExpectedEvent = &EventContractReady{}
	err := s.generateAndSetKeys()
	require.NoError(t, err)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)
	newSwap(t, s, [32]byte{}, [32]byte{}, desiredAmount.BigInt(), duration)

	txOpts, err := s.TxOpts()
	require.NoError(t, err)
	tx, err := s.Contract().SetReady(txOpts, s.contractSwap)
	require.NoError(t, err)
	tests.MineTransaction(t, s, tx)

	// runContractEventWatcher will trigger EventContractReady,
	// which will then set the next expected event to EventExit.
	for status := range s.info.StatusCh() {
		if !status.IsOngoing() {
			break
		}
	}

	require.Equal(t, types.CompletedSuccess, s.info.Status())
}

func TestSwapState_handleEvent_EventETHRefunded(t *testing.T) {
	_, s, db := newTestInstanceAndDB(t)
	db.EXPECT().PutOffer(s.offer)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setXMRTakerPublicKeys(xmrtakerKeysAndProof.PublicKeyPair, xmrtakerKeysAndProof.Secp256k1PublicKey)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	refundKey := xmrtakerKeysAndProof.Secp256k1PublicKey.Keccak256()
	newSwap(t, s, [32]byte{}, refundKey, desiredAmount.BigInt(), duration)

	// lock XMR
	_, err = s.lockFunds(common.MoneroToPiconero(s.info.ProvidedAmount()))
	require.NoError(t, err)

	// call refund w/ XMRTaker's secret
	secret := xmrtakerKeysAndProof.DLEqProof.Secret()
	sk, err := mcrypto.NewPrivateSpendKey(secret[:])
	require.NoError(t, err)

	event := newEventETHRefunded(sk)
	s.handleEvent(event)
	err = <-event.errCh
	require.NoError(t, err)
	require.Equal(t, types.CompletedRefund, s.info.Status())
}
