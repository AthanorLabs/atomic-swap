package coins

import (
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
	assert.ErrorIs(t, err, errInvalidCoin)
}
