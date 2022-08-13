package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOffer_MarshalJSON(t *testing.T) {
	offer := NewOffer(ProvidesXMR, 100.0, 200.0, 1.5)
	id := offer.GetID()
	require.False(t, id.IsZero())

	expected := fmt.Sprintf(
		`{"ID":"%s","Provides":"XMR","MinimumAmount":100,"MaximumAmount":200,"ExchangeRate":1.5}`, id)
	jsonData, err := json.Marshal(offer)
	require.NoError(t, err)
	require.Equal(t, expected, string(jsonData))
}

func TestOffer_UnmarshalJSON(t *testing.T) {
	idStr := "0102030405060708091011121314151617181920212223242526272829303131"
	offerJSON := fmt.Sprintf(
		`{"ID":"%s","Provides":"XMR","MinimumAmount":100,"MaximumAmount":200,"ExchangeRate":1.5}`, idStr)
	var offer Offer
	err := json.Unmarshal([]byte(offerJSON), &offer)
	require.NoError(t, err)
	assert.Equal(t, idStr, offer.id.String())
	assert.Equal(t, offer.Provides, ProvidesXMR)
	assert.Equal(t, offer.MinimumAmount, float64(100))
	assert.Equal(t, offer.MaximumAmount, float64(200))
	assert.Equal(t, offer.ExchangeRate, ExchangeRate(1.5))
}

func TestOffer_MarshalJSON_RoundTrip(t *testing.T) {
	offer1 := NewOffer(ProvidesXMR, 100.0, 200.0, 1.5)
	offerJSON, err := json.Marshal(offer1)
	require.NoError(t, err)
	var offer2 Offer
	err = json.Unmarshal(offerJSON, &offer2)
	require.NoError(t, err)
	assert.EqualValues(t, offer1, &offer2)
}

func TestOffer_UnmarshalJSON_BadID(t *testing.T) {
	offerJSON := []byte(`{"ID":"","Provides":"XMR","MinimumAmount":100,"MaximumAmount":200,"ExchangeRate":1.5}`)
	var offer Offer
	err := json.Unmarshal(offerJSON, &offer)
	require.Error(t, err)
	require.Equal(t, err.Error(), "offer ID has invalid length=0")
}