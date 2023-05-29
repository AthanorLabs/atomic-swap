// Copyright 2023 The AthanorLabs/atomic-swap Authors
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

var (
	testExchangeRate = coins.StrToExchangeRate("0.08")
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
		testExchangeRate,
		types.EthAssetETH,
	)
	s, err := xmrtaker.InitiateProtocol(testPeerID, providesAmount, offer)
	return offer, s, err
}

func TestXMRTaker_InitiateProtocol(t *testing.T) {
	a := newTestXMRTaker(t)
	min := coins.StrToDecimal("0.1")
	max := coins.StrToDecimal("1")

	// Provided between minAmount and maxAmount (0.05 ETH / 0.08 = 0.625 XMR)
	offer, s, err := initiate(a, coins.StrToDecimal("0.05"), min, max)
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.ID], s)

	// Exact max is in range (0.08 ETH / 0.08 = 1 XMR)
	offer, s, err = initiate(a, coins.StrToDecimal("0.08"), min, max)
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.ID], s)

	// Exact min is in range (0.008 ETH / 0.08 = 0.1 XMR)
	offer, s, err = initiate(a, coins.StrToDecimal("0.008"), min, max)
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.ID], s)

	// Provided with too many decimals
	_, s, err = initiate(a, apd.New(1, -50), min, max) // 10^-50
	require.ErrorContains(t, err, `"providesAmount" has too many decimal points; found=50 max=18`)
	require.Equal(t, nil, s)

	// Provided with a negative number
	_, s, err = initiate(a, coins.StrToDecimal("-1"), min, max)
	require.ErrorContains(t, err, `"providesAmount" cannot be negative`)
	require.Equal(t, nil, s)

	// Provided over maxAmount (0.09 ETH / 0.08 = 1.125 XMR)
	_, s, err = initiate(a, coins.StrToDecimal("0.09"), min, max)
	expected := `provided ETH converted to XMR is over offer max of 1 XMR (0.09 ETH / 0.08 = 1.125 XMR)`
	require.ErrorContains(t, err, expected)
	require.Equal(t, nil, s)

	// Provided under minAmount (0.00079 ETH / 0.08 = 0.009875 XMR)
	_, s, err = initiate(a, coins.StrToDecimal("0.00079"), min, max)
	expected = `provided ETH converted to XMR is under offer min of 0.1 XMR (0.00079 ETH / 0.08 = 0.009875)`
	require.ErrorContains(t, err, expected)
	require.Equal(t, nil, s)
}
