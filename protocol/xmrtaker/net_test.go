// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package xmrtaker

import (
	"path"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
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

func initiate(
	xmrtaker *Instance,
	providesAmount *apd.Decimal,
	minAmount *apd.Decimal,
	maxAmount *apd.Decimal,
) (*types.Offer, common.SwapState, error) {
	offer := types.NewOffer(
		coins.ProvidesETH,
		minAmount,
		maxAmount,
		coins.ToExchangeRate(apd.New(1, 0)),
		types.EthAssetETH,
	)
	s, err := xmrtaker.InitiateProtocol(testPeerID, providesAmount, offer)
	return offer, s, err
}

func TestXMRTaker_InitiateProtocol(t *testing.T) {
	a := newTestXMRTaker(t)
	zero := new(apd.Decimal)
	one := apd.New(1, 0)

	// Provided between minAmount and maxAmount
	offer, s, err := initiate(a, apd.New(1, -1), zero, one) // 0.1
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.ID], s)

	// Provided with too many decimals
	_, s, err = initiate(a, apd.New(1, -50), zero, one) // 10^-50
	require.Error(t, err)
	require.Equal(t, nil, s)

	// Provided with a negative number
	_, s, err = initiate(a, apd.New(-1, 0), zero, one) // -1
	require.Error(t, err)
	require.Equal(t, nil, s)

	// Provided over maxAmount
	_, s, err = initiate(a, apd.New(2, 0), one, one) // 2
	require.Error(t, err)
	require.Equal(t, nil, s)

	// Provided under minAmount
	_, s, err = initiate(a, apd.New(1, -1), one, one) // 0.1
	require.Error(t, err)
	require.Equal(t, nil, s)
}
