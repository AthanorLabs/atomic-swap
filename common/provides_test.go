package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewProvidesCoin(t *testing.T) {
	coin, err := NewProvidesCoin("XMR")
	require.NoError(t, err)
	require.Equal(t, ProvidesXMR, coin)

	coin, err = NewProvidesCoin("ETH")
	require.NoError(t, err)
	require.Equal(t, ProvidesETH, coin)

	_, err = NewProvidesCoin("asdf")
	require.NotNil(t, err)
}
