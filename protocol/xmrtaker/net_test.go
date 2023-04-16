// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package xmrtaker

import (
	"path"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
)

func newTestXMRTaker(t *testing.T) *Instance {
	b := newBackend(t)
	cfg := &Config{
		Backend: b,
		DataDir: path.Join(t.TempDir(), "xmrtaker"),
	}

	xmrtaker, err := NewInstance(cfg)
	require.NoError(t, err)
	return xmrtaker
}

func TestXMRTaker_InitiateProtocol(t *testing.T) {
	a := newTestXMRTaker(t)
	zero := new(apd.Decimal)
	one := apd.New(1, 0)

	// Provided between minAmount and maxAmount
	offer := types.NewOffer(coins.ProvidesETH, zero, one, coins.ToExchangeRate(one), types.EthAssetETH)
	providesAmount := apd.New(1, -1) // 0.1
	s, err := a.InitiateProtocol(testPeerID, providesAmount, offer)
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.ID], s)

	// Provided over maxAmount
	offer = types.NewOffer(coins.ProvidesETH, one, one, coins.ToExchangeRate(one), types.EthAssetETH)
	providesAmount = apd.New(2, 0) // 2
	s, err = a.InitiateProtocol(testPeerID, providesAmount, offer)
	require.Error(t, err)
	require.Equal(t, nil, s)

	// Provided under maxAmount
	offer = types.NewOffer(coins.ProvidesETH, one, one, coins.ToExchangeRate(one), types.EthAssetETH)
	providesAmount = apd.New(1, -1) // 0.1
	s, err = a.InitiateProtocol(testPeerID, providesAmount, offer)
	require.Error(t, err)
	require.Equal(t, nil, s)
}
