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
	require.False(t, IsHashZero(offer.ID))

	expected := fmt.Sprintf(`{
		"version": "0.1.0",
		"offerID": "%s",
		"provides": "XMR",
		"minAmount": 100,
		"maxAmount": 200,
		"exchangeRate": 1.5,
		"ethAsset": "ETH"
	}`, offer.ID)
	jsonData, err := json.Marshal(offer)
	require.NoError(t, err)
	require.JSONEq(t, expected, string(jsonData))
}

func TestOffer_UnmarshalJSON(t *testing.T) {
	idStr := "0x0102030405060708091011121314151617181920212223242526272829303131"
	offerJSON := fmt.Sprintf(`{
		"version": "0.1.0",
		"offerID": "%s",
		"provides": "XMR",
		"minAmount": 100,
		"maxAmount": 200,
		"exchangeRate": 1.5,
		"ethAsset":"0x0000000000000000000000000000000000000001"
	}`, idStr)
	var offer Offer
	err := json.Unmarshal([]byte(offerJSON), &offer)
	require.NoError(t, err)
	assert.Equal(t, idStr, offer.ID.String())
	assert.Equal(t, offer.Provides, ProvidesXMR)
	assert.Equal(t, offer.MinAmount, float64(100))
	assert.Equal(t, offer.MaxAmount, float64(200))
	assert.Equal(t, offer.ExchangeRate, ExchangeRate(1.5))
	assert.Equal(t, "0x0000000000000000000000000000000000000001", ethcommon.Address(offer.EthAsset).Hex())
}

func TestOffer_UnmarshalJSON_DefaultAsset(t *testing.T) {
	idStr := "0x0102030405060708091011121314151617181920212223242526272829303131"
	offerJSON := fmt.Sprintf(`{
		"version": "0.1.0",
		"offerID": "%s",
		"provides": "XMR",
		"minAmount": 100,
		"maxAmount": 200,
		"exchangeRate": 1.5
	}`, idStr)
	var offer Offer
	err := json.Unmarshal([]byte(offerJSON), &offer)
	require.NoError(t, err)
	assert.Equal(t, *CurOfferVersion, offer.Version)
	assert.Equal(t, idStr, offer.ID.String())
	assert.Equal(t, offer.Provides, ProvidesXMR)
	assert.Equal(t, offer.MinAmount, float64(100))
	assert.Equal(t, offer.MaxAmount, float64(200))
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
		"offerID": "",
		"provides": "XMR",
		"minAmount": 100,
		"maxAmount": 200,
		"exchangeRate": 1.5,
		"ethAsset": "ETH"
	}`)
	_, err := UnmarshalOffer(offerJSON)
	require.Error(t, err)
	require.ErrorContains(t, err, "hex string has length 0, want 64")
}

func TestUnmarshalOffer_missingVersion(t *testing.T) {
	_, err := UnmarshalOffer([]byte(`{}`))
	require.ErrorIs(t, err, errOfferVersionMissing)
}

func TestUnmarshalOffer_versionTooNew(t *testing.T) {
	unsupportedVersion := CurOfferVersion.IncMajor()
	offerJSON := fmt.Sprintf(`{
		"version": "%s",
		"some_unsupported_field": ""
	}`, unsupportedVersion)
	_, err := UnmarshalOffer([]byte(offerJSON))
	require.ErrorContains(t, err, fmt.Sprintf("offer version %q not supported", unsupportedVersion))
}
