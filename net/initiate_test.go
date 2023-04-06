// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package net

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func createSendKeysMessage(t *testing.T) *message.SendKeysMessage {
	keysAndProof, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	return &message.SendKeysMessage{
		OfferID:            types.Hash{0x1},
		ProvidedAmount:     coins.StrToDecimal("0.5"),
		PublicSpendKey:     keysAndProof.PublicKeyPair.SpendKey(),
		PrivateViewKey:     keysAndProof.PrivateKeyPair.ViewKey(),
		DLEqProof:          keysAndProof.DLEqProof.Proof(),
		Secp256k1PublicKey: keysAndProof.Secp256k1PublicKey,
		EthAddress:         crypto.PubkeyToAddress(tests.GetMakerTestKey(t).PublicKey),
	}
}

func TestHost_Initiate(t *testing.T) {
	ha := newHost(t, basicTestConfig(t))
	err := ha.Start()
	require.NoError(t, err)
	hb := newHost(t, basicTestConfig(t))
	err = hb.Start()
	require.NoError(t, err)

	err = ha.h.Connect(ha.ctx, hb.h.AddrInfo())
	require.NoError(t, err)

	err = ha.Initiate(hb.h.AddrInfo(), createSendKeysMessage(t), new(mockSwapState))
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 500)

	ha.swapMu.RLock()
	require.NotNil(t, ha.swaps[testID])
	ha.swapMu.RUnlock()

	hb.swapMu.RLock()
	require.NotNil(t, hb.swaps[testID])
	hb.swapMu.RUnlock()
}

func TestHost_ConcurrentSwaps(t *testing.T) {
	ha := newHost(t, basicTestConfig(t))
	err := ha.Start()
	require.NoError(t, err)

	hbCfg := basicTestConfig(t)
	hbCfg.Bootnodes = []string{ha.Addresses()[0].String()} // get some test coverage on our bootnode code
	hb := newHost(t, hbCfg)
	err = hb.Start()
	require.NoError(t, err)

	testID2 := types.Hash{98}

	err = ha.h.Connect(ha.ctx, hb.h.AddrInfo())
	require.NoError(t, err)

	err = ha.Initiate(hb.h.AddrInfo(), createSendKeysMessage(t), new(mockSwapState))
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 500)

	ha.swapMu.RLock()
	require.NotNil(t, ha.swaps[testID])
	ha.swapMu.RUnlock()

	hb.swapMu.RLock()
	require.NotNil(t, hb.swaps[testID])
	hb.swapMu.RUnlock()

	hb.makerHandler.(*mockMakerHandler).id = testID2

	err = ha.Initiate(hb.h.AddrInfo(), createSendKeysMessage(t), &mockSwapState{testID2})
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 1500)

	ha.swapMu.RLock()
	require.NotNil(t, ha.swaps[testID])
	ha.swapMu.RUnlock()

	hb.swapMu.RLock()
	require.NotNil(t, hb.swaps[testID])
	hb.swapMu.RUnlock()
}
