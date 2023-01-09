package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
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
		"minAmount": "100",
		"maxAmount": "200",
		"exchangeRate": "1.5",
		"ethAsset":"0x0000000000000000000000000000000000000001"
	}`, idStr)
	var offer Offer
	err := json.Unmarshal([]byte(offerJSON), &offer)
	require.NoError(t, err)
	assert.Equal(t, idStr, offer.ID.String())
	assert.Equal(t, offer.Provides, coins.ProvidesXMR)
	assert.Equal(t, offer.MinAmount.Text('f'), "100")
	assert.Equal(t, offer.MaxAmount.Text('f'), "200")
	assert.Equal(t, offer.ExchangeRate.String(), "1.5")
	assert.Equal(t, "0x0000000000000000000000000000000000000001", ethcommon.Address(offer.EthAsset).Hex())
}

func TestOffer_UnmarshalJSON_DefaultAsset(t *testing.T) {
	idStr := "0x0102030405060708091011121314151617181920212223242526272829303131"
	offerJSON := fmt.Sprintf(`{
		"version": "1.0.0",
		"offerID": "%s",
		"provides": "XMR",
		"minAmount": "100",
		"maxAmount": "200",
		"exchangeRate": "1.5"
	}`, idStr)
	var offer Offer
	err := json.Unmarshal([]byte(offerJSON), &offer)
	require.NoError(t, err)
	assert.Equal(t, *CurOfferVersion, offer.Version)
	assert.Equal(t, idStr, offer.ID.String())
	assert.Equal(t, offer.Provides, coins.ProvidesXMR)
	assert.Equal(t, offer.MinAmount.Text('f'), "100")
	assert.Equal(t, offer.MaxAmount.Text('f'), "200")
	assert.Equal(t, offer.ExchangeRate.String(), "1.5")
	assert.Equal(t, offer.EthAsset, EthAssetETH)
}

func TestOffer_MarshalJSON_RoundTrip(t *testing.T) {
	min := apd.New(100, 0)
	max := apd.New(200, 0)
	rate := coins.ToExchangeRate(apd.New(15, -1)) // 1.5
	offer1 := NewOffer(coins.ProvidesXMR, min, max, rate, EthAssetETH)
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
		"ethAsset": "ETH"
	}`)
	_, err := UnmarshalOffer(offerJSON)
	require.ErrorIs(t, err, errOfferIDNotSet)
}

func TestOffer_UnmarshalJSON_BadAmountsOrRate(t *testing.T) {
	offerJSON := `{
		"offerID": "0x0102030405060708091011121314151617181920212223242526272829303131",
		"provides": "XMR",
		"minAmount": %s,
		"maxAmount": %s,
		"exchangeRate": %s,
		"ethAsset": "ETH"
	}`
	type entry struct {
		jsonData    string
		errContains string
	}
	testEntries := []entry{
		// Min amount checks
		{
			jsonData:    fmt.Sprintf(offerJSON, `null`, `"1"`, `"0.1"`),
			errContains: `"minAmount" is not set`,
		},
		{
			jsonData:    fmt.Sprintf(offerJSON, `"0"`, `"1"`, `"0.1"`),
			errContains: `"minAmount" must be non-zero`,
		},
		{
			jsonData:    fmt.Sprintf(offerJSON, `"-1"`, `"1"`, `"0.1"`),
			errContains: `"minAmount" cannot be negative`,
		},
		// Max Amount checks
		{
			jsonData:    fmt.Sprintf(offerJSON, `"1"`, `null`, `"0.1"`),
			errContains: `"maxAmount" is not set`,
		},
		{
			jsonData:    fmt.Sprintf(offerJSON, `"1"`, `"-0"`, `"0.1"`),
			errContains: `"maxAmount" must be non-zero`,
		},
		{
			jsonData:    fmt.Sprintf(offerJSON, `"1"`, `"-1E1"`, `"0.1"`),
			errContains: `"maxAmount" cannot be negative`,
		},
		// Combo min/max check
		{
			jsonData:    fmt.Sprintf(offerJSON, `"0.11"`, `"0.1"`, `"0.1"`),
			errContains: `"minAmount" must be less than or equal to "maxAmount"`,
		},
		// Exchange rate checks
		{
			jsonData:    fmt.Sprintf(offerJSON, `"1"`, `"1"`, `null`),
			errContains: `"exchangeRate" is not set`,
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
		err := json.Unmarshal([]byte(e.jsonData), o)
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
	err := json.Unmarshal(offerJSON, new(Offer))
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
