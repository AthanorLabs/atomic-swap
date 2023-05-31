// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/vjson"
)

func TestOffer_MarshalJSON(t *testing.T) {
	min := apd.New(101, 0)
	max := apd.New(202, 0)
	rate := coins.ToExchangeRate(apd.New(15, -1)) // 1.5
	offer := NewOffer(coins.ProvidesXMR, min, max, rate, EthAssetETH)
	require.False(t, IsHashZero(offer.ID))

	expected := fmt.Sprintf(`{
		"version": "1.0.0",
		"offerID": "%s",
		"provides": "XMR",
		"minAmount": "101",
		"maxAmount": "202",
		"exchangeRate": "1.5",
		"ethAsset": "ETH",
		"nonce": %d
	}`, offer.ID, offer.Nonce)
	jsonData, err := vjson.MarshalStruct(offer)
	require.NoError(t, err)
	require.JSONEq(t, expected, string(jsonData))
}

func TestOffer_UnmarshalJSON(t *testing.T) {
	min := apd.New(100, 0)
	max := apd.New(200, 0)
	rate := coins.ToExchangeRate(apd.New(15, -1)) // 1.5
	ethAsset := EthAsset(
		ethcommon.HexToAddress("0x0000000000000000000000000000000000000001"),
	)
	offer := NewOffer(coins.ProvidesXMR, min, max, rate, ethAsset)
	require.False(t, IsHashZero(offer.ID))
	v, _ := semver.NewVersion("0.1.0")
	offer.Version = *v
	offer.ID = offer.hash()

	offerJSON := fmt.Sprintf(`{
		"version": "0.1.0",
		"offerID": "%s",
		"provides": "XMR",
		"minAmount": "100",
		"maxAmount": "200",
		"exchangeRate": "1.5",
		"ethAsset":"0x0000000000000000000000000000000000000001",
		"nonce": %d
	}`, offer.ID, offer.Nonce)

	var res Offer
	err := vjson.UnmarshalStruct([]byte(offerJSON), &res)
	require.NoError(t, err)
	assert.Equal(t, offer.ID, res.ID)
	assert.Equal(t, res.Provides, coins.ProvidesXMR)
	assert.Equal(t, res.MinAmount.Text('f'), "100")
	assert.Equal(t, res.MaxAmount.Text('f'), "200")
	assert.Equal(t, res.ExchangeRate.String(), "1.5")
	assert.Equal(t, ethAsset, res.EthAsset)
}

func TestOffer_UnmarshalJSON_DefaultAsset(t *testing.T) {
	min := apd.New(100, 0)
	max := apd.New(200, 0)
	rate := coins.ToExchangeRate(apd.New(15, -1)) // 1.5
	offer := NewOffer(coins.ProvidesXMR, min, max, rate, EthAssetETH)
	require.False(t, IsHashZero(offer.ID))

	offerJSON := fmt.Sprintf(`{
		"version": "1.0.0",
		"offerID": "%s",
		"provides": "XMR",
		"minAmount": "100",
		"maxAmount": "200",
		"exchangeRate": "1.5",
		"nonce": %d
	}`, offer.ID, offer.Nonce)

	var res Offer
	err := vjson.UnmarshalStruct([]byte(offerJSON), &res)
	require.NoError(t, err)
	assert.Equal(t, *CurOfferVersion, offer.Version)
	assert.Equal(t, offer.ID, res.ID)
	assert.Equal(t, res.Provides, coins.ProvidesXMR)
	assert.Equal(t, res.MinAmount.Text('f'), "100")
	assert.Equal(t, res.MaxAmount.Text('f'), "200")
	assert.Equal(t, res.ExchangeRate.String(), "1.5")
	assert.Equal(t, EthAssetETH, res.EthAsset)
}

func TestOffer_MarshalJSON_RoundTrip(t *testing.T) {
	min := apd.New(100, 0)
	max := apd.New(200, 0)
	rate := coins.ToExchangeRate(apd.New(15, -1)) // 1.5
	offer1 := NewOffer(coins.ProvidesXMR, min, max, rate, EthAssetETH)
	offerJSON, err := vjson.MarshalStruct(offer1)
	require.NoError(t, err)
	var offer2 Offer
	err = vjson.UnmarshalStruct(offerJSON, &offer2)
	require.NoError(t, err)
	assert.Equal(t, offer1.Version.String(), offer2.Version.String())
	offer2.Version = offer1.Version // make the version pointers equal for the next line
	assert.EqualValues(t, offer1, &offer2)
}

func TestOffer_UnmarshalJSON_BadID(t *testing.T) {
	offerJSON := []byte(`{
		"version": "0.1.0",
		"offerID": "",
		"provides": "XMR",
		"minAmount": "100",
		"maxAmount": "200",
		"exchangeRate": "1.5",
		"ethAsset": "ETH"
	}`)
	_, err := UnmarshalOffer(offerJSON)
	require.Error(t, err)
	require.ErrorContains(t, err, "hex string has length 0, want 64")
}

func TestOffer_UnmarshalJSON_MissingID(t *testing.T) {
	offerJSON := []byte(`{
		"version": "0.1.0",
		"provides": "XMR",
		"minAmount": "100",
		"maxAmount": "200",
		"exchangeRate": "1.5",
		"ethAsset": "ETH",
		"nonce": 1234
	}`)
	_, err := UnmarshalOffer(offerJSON)
	require.ErrorContains(t, err, "Field validation for 'ID' failed on the 'required' tag")
}

func TestOffer_UnmarshalJSON_BadAmountsOrRate(t *testing.T) {
	offerJSON := `{
		"offerID": "0x0102030405060708091011121314151617181920212223242526272829303131",
		"provides": "XMR",
		"minAmount": %s,
		"maxAmount": %s,
		"exchangeRate": %s,
		"ethAsset": "ETH",
		"nonce": 1234
	}`
	type entry struct {
		jsonData    string
		errContains string
	}
	testEntries := []entry{
		// Min amount checks
		{
			jsonData:    fmt.Sprintf(offerJSON, `null`, `"1"`, `"0.1"`),
			errContains: `Field validation for 'MinAmount' failed on the 'required' tag`,
		},
		{
			jsonData:    fmt.Sprintf(offerJSON, `"0"`, `"1"`, `"0.1"`),
			errContains: `"minAmount" must be non-zero`,
		},
		{
			jsonData:    fmt.Sprintf(offerJSON, `"-1"`, `"1"`, `"0.1"`),
			errContains: `"minAmount" cannot be negative`,
		},
		{
			// 0.01 relayer fee is 0.1 XMR with exchange rate of 0.1
			jsonData:    fmt.Sprintf(offerJSON, `"0.01"`, `"10"`, `"0.1"`),
			errContains: `min amount must be greater than 0.01 ETH when converted (10 XMR * 0.1 = 0.001 ETH)`,
		},
		// Max Amount checks
		{
			jsonData:    fmt.Sprintf(offerJSON, `"1"`, `null`, `"0.1"`),
			errContains: `Field validation for 'MaxAmount' failed on the 'required' tag`,
		},
		{
			jsonData:    fmt.Sprintf(offerJSON, `"1"`, `"-0"`, `"0.1"`),
			errContains: `"maxAmount" must be non-zero`,
		},
		{
			jsonData:    fmt.Sprintf(offerJSON, `"1"`, `"-1E1"`, `"0.1"`),
			errContains: `"maxAmount" cannot be negative`,
		},
		{
			jsonData:    fmt.Sprintf(offerJSON, `"100"`, `"1000.1"`, `"0.1"`),
			errContains: `1000.1 XMR exceeds max offer amount of 1000 XMR`,
		},
		// Combo min/max check
		{
			jsonData:    fmt.Sprintf(offerJSON, `"0.11"`, `"0.1"`, `"0.1"`),
			errContains: `"minAmount" must be less than or equal to "maxAmount"`,
		},
		// Exchange rate checks
		{
			jsonData:    fmt.Sprintf(offerJSON, `"1"`, `"1"`, `null`),
			errContains: `Field validation for 'ExchangeRate' failed on the 'required' tag`,
		},
		{
			jsonData:    fmt.Sprintf(offerJSON, `"1"`, `"1"`, `"0"`),
			errContains: `"exchangeRate" must be non-zero`,
		},
		{
			jsonData:    fmt.Sprintf(offerJSON, `"1"`, `"1"`, `"-0.1"`),
			errContains: `"exchangeRate" cannot be negative`,
		},
	}
	for _, e := range testEntries {
		o := new(Offer)
		err := vjson.UnmarshalStruct([]byte(e.jsonData), o)
		assert.ErrorContains(t, err, e.errContains)
	}
}

func TestOffer_UnmarshalJSON_BadProvides(t *testing.T) {
	offerJSON := []byte(`{
		"offerID": "0x0102030405060708091011121314151617181920212223242526272829303131",
		"provides": "",
		"minAmount": "0.1",
		"maxAmount": "0.2",
		"exchangeRate": "0.5",
		"ethAsset": "ETH"
	}`)
	err := vjson.UnmarshalStruct(offerJSON, new(Offer))
	assert.ErrorIs(t, err, coins.ErrInvalidCoin)
}

func TestUnmarshalOffer_MissingVersion(t *testing.T) {
	_, err := UnmarshalOffer([]byte(`{}`))
	require.ErrorIs(t, err, errOfferVersionMissing)
}

func TestUnmarshalOffer_VersionTooNew(t *testing.T) {
	unsupportedVersion := CurOfferVersion.IncMajor()
	offerJSON := fmt.Sprintf(`{
		"version": "%s",
		"some_unsupported_field": ""
	}`, unsupportedVersion)
	_, err := UnmarshalOffer([]byte(offerJSON))
	require.ErrorContains(t, err, fmt.Sprintf("offer version %q not supported", unsupportedVersion))
}

func TestOfferExtra_JSON(t *testing.T) {
	// Marshal test
	extra := NewOfferExtra(true)
	data, err := vjson.MarshalStruct(extra)
	require.NoError(t, err)
	require.JSONEq(t, `{"useRelayer":true}`, string(data))

	// Unmarshal test
	extra = new(OfferExtra)
	err = json.Unmarshal(data, extra)
	require.NoError(t, err)
	require.True(t, extra.UseRelayer)
}
