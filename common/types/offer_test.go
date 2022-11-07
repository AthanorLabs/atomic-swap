package types

import (
	"encoding/json"
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOffer_MarshalJSON(t *testing.T) {
	offer := NewOffer(ProvidesXMR, 100.0, 200.0, 1.5, EthAssetETH)
	id := offer.GetID()
	require.False(t, IsHashZero(id))

	expected := fmt.Sprintf(`{
		"version": "0.1.0",
		"offer_id": "%s",
		"provides": "XMR",
		"min_amount": 100,
		"max_amount": 200,
		"exchange_rate": 1.5,
		"eth_asset": "ETH"
	}`, id)
	jsonData, err := json.Marshal(offer)
	require.NoError(t, err)
	require.JSONEq(t, expected, string(jsonData))
}

func TestOffer_UnmarshalJSON(t *testing.T) {
	idStr := "0x0102030405060708091011121314151617181920212223242526272829303131"
	offerJSON := fmt.Sprintf(`{
		"version": "0.1.0",
		"offer_id": "%s",
		"provides": "XMR",
		"min_amount": 100,
		"max_amount": 200,
		"exchange_rate": 1.5,
		"eth_asset":"0x0000000000000000000000000000000000000001"
	}`, idStr)
	var offer Offer
	err := json.Unmarshal([]byte(offerJSON), &offer)
	require.NoError(t, err)
	assert.Equal(t, idStr, offer.ID.String())
	assert.Equal(t, offer.Provides, ProvidesXMR)
	assert.Equal(t, offer.MinimumAmount, float64(100))
	assert.Equal(t, offer.MaximumAmount, float64(200))
	assert.Equal(t, offer.ExchangeRate, ExchangeRate(1.5))
	assert.Equal(t, ethcommon.Address(offer.EthAsset).Hex(), "0x0000000000000000000000000000000000000001")
}

func TestOffer_UnmarshalJSON_DefaultAsset(t *testing.T) {
	idStr := "0102030405060708091011121314151617181920212223242526272829303131"
	offerJSON := fmt.Sprintf(`{
		"version": "0.1.0",
		"offer_id": "%s",
		"provides": "XMR",
		"min_amount": 100,
		"max_amount": 200,
		"exchange_rate": 1.5
	}`, idStr)
	var offer Offer
	err := json.Unmarshal([]byte(offerJSON), &offer)
	require.NoError(t, err)
	assert.Equal(t, CurOfferVersion, offer.Version.String())
	assert.Equal(t, idStr, offer.ID.String())
	assert.Equal(t, offer.Provides, ProvidesXMR)
	assert.Equal(t, offer.MinimumAmount, float64(100))
	assert.Equal(t, offer.MaximumAmount, float64(200))
	assert.Equal(t, offer.ExchangeRate, ExchangeRate(1.5))
	assert.Equal(t, offer.EthAsset, EthAssetETH)
}

func TestOffer_MarshalJSON_RoundTrip(t *testing.T) {
	offer1 := NewOffer(ProvidesXMR, 100.0, 200.0, 1.5, EthAssetETH)
	offerJSON, err := json.Marshal(offer1)
	require.NoError(t, err)
	var offer2 Offer
	err = json.Unmarshal(offerJSON, &offer2)
	require.NoError(t, err)
	assert.Equal(t, offer1.Version.String(), offer2.Version.String())
	offer2.Version = offer1.Version // make the version pointers equal for the next line
	assert.EqualValues(t, offer1, &offer2)
}

func TestOffer_UnmarshalJSON_BadID(t *testing.T) {
	offerJSON := []byte(`{
		"version": "0.1.0",
		"offer_id": "",
		"provides": "XMR",
		"min_Amount": 100,
		"max_Amount": 200,
		"exchange_rate": 1.5,
		"eth_asset": "ETH"
	}`)
	var offer Offer
	err := json.Unmarshal(offerJSON, &offer)
	require.Error(t, err)
	require.ErrorContains(t, err, "hex string has length 0, want 64")
}
