// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProvidesCoin(t *testing.T) {
	coin, err := NewProvidesCoin("XMR")
	require.NoError(t, err)
	assert.Equal(t, ProvidesXMR, coin)

	coin, err = NewProvidesCoin("ETH")
	require.NoError(t, err)
	assert.Equal(t, ProvidesETH, coin)

	_, err = NewProvidesCoin("asdf")
	assert.ErrorIs(t, err, ErrInvalidCoin)
}

func TestProvidesCoinMarshal(t *testing.T) {
	type M struct {
		C ProvidesCoin
	}

	data, err := json.Marshal(&M{ProvidesXMR})
	require.NoError(t, err)
	assert.JSONEq(t, `{"C":"XMR"}`, string(data))

	data, err = json.Marshal(&M{ProvidesETH})
	require.NoError(t, err)
	assert.JSONEq(t, `{"C":"ETH"}`, string(data))

	// We are slightly lenient in what we accept, but string in what we generate
	_, err = json.Marshal(&M{"xmr"})
	require.ErrorContains(t, err, "cannot marshal")
	_, err = json.Marshal(&M{"eth"})
	require.ErrorContains(t, err, "cannot marshal")
	_, err = json.Marshal(&M{""})
	require.ErrorContains(t, err, "cannot marshal")
}

func TestProvidesCoinUnMarshal(t *testing.T) {
	type M struct {
		C ProvidesCoin
	}
	m := new(M)

	err := json.Unmarshal([]byte(`{"C":"XMR"}`), m)
	require.NoError(t, err)
	assert.Equal(t, ProvidesXMR, m.C)

	err = json.Unmarshal([]byte(`{"C":"xmr"}`), m)
	require.NoError(t, err)
	assert.Equal(t, ProvidesXMR, m.C)

	err = json.Unmarshal([]byte(`{"C":"ETH"}`), m)
	require.NoError(t, err)
	assert.Equal(t, ProvidesETH, m.C)

	err = json.Unmarshal([]byte(`{"C":"eth"}`), m)
	require.NoError(t, err)
	assert.Equal(t, ProvidesETH, m.C)

}
