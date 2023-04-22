// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

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

var fakeSwapKey = [32]byte{1}

func TestSwapState_handleEvent_EventContractReady(t *testing.T) {
	_, s := newTestSwapState(t)
	s.nextExpectedEvent = EventContractReadyType

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)
	newSwap(t, s, fakeSwapKey, fakeSwapKey, desiredAmount.BigInt(), duration)

	txOpts, err := s.ETHClient().TxOpts(s.ctx)
	require.NoError(t, err)
	tx, err := s.SwapCreator().SetReady(txOpts, *s.contractSwap)
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
	err = s.setXMRTakerKeys(
		xmrtakerKeysAndProof.PublicKeyPair.SpendKey(),
		xmrtakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrtakerKeysAndProof.Secp256k1PublicKey,
	)
	require.NoError(t, err)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	refundKey := xmrtakerKeysAndProof.Secp256k1PublicKey.Keccak256()
	newSwap(t, s, fakeSwapKey, refundKey, desiredAmount.BigInt(), duration)

	// lock XMR
	err = s.lockFunds(coins.MoneroToPiconero(s.info.ProvidedAmount))
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
