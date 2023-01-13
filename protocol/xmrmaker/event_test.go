package xmrmaker

import (
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func TestSwapState_handleEvent_EventContractReady(t *testing.T) {
	_, s := newTestSwapState(t)
	s.nextExpectedEvent = EventContractReadyType

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)
	newSwap(t, s, [32]byte{}, [32]byte{}, desiredAmount.BigInt(), duration)

	txOpts, err := s.ETHClient().TxOpts(s.ctx)
	require.NoError(t, err)
	tx, err := s.Contract().SetReady(txOpts, s.contractSwap)
	require.NoError(t, err)
	tests.MineTransaction(t, s.ETHClient().Raw(), tx)

	// runContractEventWatcher will trigger EventContractReady,
	// which will then set the next expected event to EventExit.
	for status := range s.info.StatusCh() {
		if !status.IsOngoing() {
			break
		}
	}

	require.Equal(t, types.CompletedSuccess, s.info.Status)
}

func TestSwapState_handleEvent_EventETHRefunded(t *testing.T) {
	_, s, db := newTestSwapStateAndDB(t)
	db.EXPECT().PutOffer(s.offer)

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setXMRTakerPublicKeys(xmrtakerKeysAndProof.PublicKeyPair, xmrtakerKeysAndProof.Secp256k1PublicKey)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	refundKey := xmrtakerKeysAndProof.Secp256k1PublicKey.Keccak256()
	newSwap(t, s, [32]byte{}, refundKey, desiredAmount.BigInt(), duration)

	// lock XMR
	_, err = s.lockFunds(coins.MoneroToPiconero(s.info.ProvidedAmount))
	require.NoError(t, err)

	// call refund w/ XMRTaker's secret
	secret := xmrtakerKeysAndProof.DLEqProof.Secret()
	sk, err := mcrypto.NewPrivateSpendKey(common.Reverse(secret[:]))
	require.NoError(t, err)

	event := newEventETHRefunded(sk)
	s.handleEvent(event)
	err = <-event.errCh
	require.NoError(t, err)
	require.Equal(t, types.CompletedRefund, s.info.Status)
}
