package types

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOffer_MarshalJSON(t *testing.T) {
	offer := NewOffer(ProvidesXMR, 100.0, 200.0, 1.5, EthAssetETH)
	//id := offer.ID
	require.False(t, offer.ID.IsZero())

	expected := fmt.Sprintf(
		`{"ID":"%s","Provides":"XMR","MinimumAmount":100,"MaximumAmount":200,"ExchangeRate":1.5,`+
			`"EthAsset":"0x0000000000000000000000000000000000000000"}`, offer.ID)
	jsonData, err := json.Marshal(offer)
	require.NoError(t, err)
	require.Equal(t, expected, string(jsonData))
}

func TestOffer_UnmarshalJSON(t *testing.T) {
	idStr := "0102030405060708091011121314151617181920212223242526272829303131"
	offerJSON := fmt.Sprintf(
		`{"ID":"%s","Provides":"XMR","MinimumAmount":100,"MaximumAmount":200,"ExchangeRate":1.5,`+
			`"EthAsset":"0x0000000000000000000000000000000000000001"}`, idStr)
	var offer Offer
	err := json.Unmarshal([]byte(offerJSON), &offer)
	require.NoError(t, err)
	assert.Equal(t, idStr, offer.ID.String())
	assert.Equal(t, offer.Provides, ProvidesXMR)
	assert.Equal(t, offer.MinimumAmount, float64(100))
	assert.Equal(t, offer.MaximumAmount, float64(200))
	assert.Equal(t, offer.ExchangeRate, ExchangeRate(1.5))
	assert.Equal(t, "0x0000000000000000000000000000000000000001", ethcommon.Address(offer.EthAsset).Hex())
}

func TestOffer_UnmarshalJSON_DefaultAsset(t *testing.T) {
	idStr := "0102030405060708091011121314151617181920212223242526272829303131"
	offerJSON := fmt.Sprintf(
		`{"ID":"%s","Provides":"XMR","MinimumAmount":100,"MaximumAmount":200,"ExchangeRate":1.5}`, idStr)
	var offer Offer
	err := json.Unmarshal([]byte(offerJSON), &offer)
	require.NoError(t, err)
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
	assert.EqualValues(t, offer1, &offer2)
}

func TestOffer_UnmarshalJSON_BadID(t *testing.T) {
	offerJSON := []byte(`{"ID":"","Provides":"XMR","MinimumAmount":100,"MaximumAmount":200,"ExchangeRate":1.5,` +
		`"EthAsset":"0x0000000000000000000000000000000000000000"}`)
	var offer Offer
	err := json.Unmarshal(offerJSON, &offer)
	require.Error(t, err)
	require.True(t, errors.Is(errInvalidHashString, err))
}

func TestHash_JSON(t *testing.T) {
	hashStr := "6ea4b13eb0b4f48bfbff416f78d817e43226a215ccaddc1ce90fb3cea893e0f3"
	b, err := hex.DecodeString(hashStr)
	require.NoError(t, err)

	var hash Hash
	copy(hash[:], b[:])

	enc, err := json.Marshal(hash)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("\"%s\"", hashStr), string(enc))

	var res Hash
	err = json.Unmarshal(enc, &res)
	require.NoError(t, err)
	require.Equal(t, hash, res)
}
