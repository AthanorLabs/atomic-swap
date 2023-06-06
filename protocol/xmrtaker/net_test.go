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
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
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
	providesAsset types.EthAsset,
	minAmount *apd.Decimal,
	maxAmount *apd.Decimal,
	exchangeRate *coins.ExchangeRate,
) (*types.Offer, common.SwapState, error) {
	offer := types.NewOffer(
		coins.ProvidesETH,
		minAmount,
		maxAmount,
		exchangeRate,
		providesAsset,
	)
	s, err := xmrtaker.InitiateProtocol(testPeerID, providesAmount, offer)
	return offer, s, err
}

func TestXMRTaker_InitiateProtocol_ETH(t *testing.T) {
	a := newTestXMRTaker(t)
	min := coins.StrToDecimal("0.1")
	max := coins.StrToDecimal("1")
	exRate := coins.StrToExchangeRate("0.08")
	asset := types.EthAssetETH

	// Provided between minAmount and maxAmount (0.05 ETH / 0.08 = 0.625 XMR)
	offer, s, err := initiate(a, coins.StrToDecimal("0.05"), asset, min, max, exRate)
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.ID], s)

	// Exact max is in range (0.08 ETH / 0.08 = 1 XMR)
	offer, s, err = initiate(a, coins.StrToDecimal("0.08"), asset, min, max, exRate)
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.ID], s)

	// Exact min is in range (0.008 ETH / 0.08 = 0.1 XMR)
	offer, s, err = initiate(a, coins.StrToDecimal("0.008"), asset, min, max, exRate)
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.ID], s)

	// Provided with too many decimals
	_, s, err = initiate(a, apd.New(1, -50), asset, min, max, exRate) // 10^-50
	require.ErrorContains(t, err, `"providesAmount" has too many decimal points; found=50 max=18`)
	require.Equal(t, nil, s)

	// Provided with a negative number
	_, s, err = initiate(a, coins.StrToDecimal("-1"), asset, min, max, exRate)
	require.ErrorContains(t, err, `"providesAmount" cannot be negative`)
	require.Equal(t, nil, s)

	// Provided over maxAmount (0.09 ETH / 0.08 = 1.125 XMR)
	_, s, err = initiate(a, coins.StrToDecimal("0.09"), asset, min, max, exRate)
	expected := `provided ETH converted to XMR is over offer max of 1 XMR (0.09 ETH / 0.08 = 1.125 XMR)`
	require.ErrorContains(t, err, expected)
	require.Equal(t, nil, s)

	// Provided under minAmount (0.00079 ETH / 0.08 = 0.009875 XMR)
	_, s, err = initiate(a, coins.StrToDecimal("0.00079"), asset, min, max, exRate)
	expected = `provided ETH converted to XMR is under offer min of 0.1 XMR (0.00079 ETH / 0.08 = 0.009875)`
	require.ErrorContains(t, err, expected)
	require.Equal(t, nil, s)
}

func TestXMRTaker_InitiateProtocol_token(t *testing.T) {
	a := newTestXMRTaker(t)
	min := coins.StrToDecimal("1")
	max := coins.StrToDecimal("2")
	ec := a.backend.ETHClient()
	token := contracts.GetMockTether(t, ec.Raw(), ec.PrivateKey())
	asset := types.EthAsset(token.Address)
	exRate := coins.StrToExchangeRate("160")

	// Provided between minAmount and maxAmount (200 USDT / 160 = 1.25 XMR)
	offer, s, err := initiate(a, coins.StrToDecimal("200"), asset, min, max, exRate)
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.ID], s)

	// Exact max is in range (320 USDT / 160 = 2 XMR)
	offer, s, err = initiate(a, coins.StrToDecimal("320"), asset, min, max, exRate)
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.ID], s)

	// Exact min is in range (160 USDT / 160 = 1 XMR)
	offer, s, err = initiate(a, coins.StrToDecimal("160"), asset, min, max, exRate)
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.ID], s)

	// Provided with too many decimals
	_, s, err = initiate(a, apd.New(1, -7), asset, min, max, exRate) // 10^-7
	require.ErrorContains(t, err, `"providesAmount" has too many decimal points; found=7 max=6`)
	require.Equal(t, nil, s)

	// Provided with a negative number
	_, s, err = initiate(a, coins.StrToDecimal("-1"), asset, min, max, exRate)
	require.ErrorContains(t, err, `"providesAmount" cannot be negative`)
	require.Equal(t, nil, s)

	// Provided over maxAmount (320.5 USDT / 160 = 2.003125 XMR)
	_, s, err = initiate(a, coins.StrToDecimal("320.5"), asset, min, max, exRate)
	expected := `provided "USDT" converted to XMR is over offer max of 2 XMR (320.5 "USDT" / 160 = 2.003125 XMR)`
	require.ErrorContains(t, err, expected)
	require.Equal(t, nil, s)

	// Provided under minAmount (159.98 USDT / 160 = 0.999875 XMR)
	_, s, err = initiate(a, coins.StrToDecimal("159.98"), asset, min, max, exRate)
	expected = `provided "USDT" converted to XMR is under offer min of 1 XMR (159.98 "USDT" / 160 = 0.999875)`
	require.ErrorContains(t, err, expected)
	require.Equal(t, nil, s)
}
