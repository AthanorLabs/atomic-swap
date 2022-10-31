package xmrtaker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
)

func TestNewInstance(t *testing.T) {
	i, err := NewInstance(&Config{
		Backend:        newBackend(t),
		DataDir:        "",
		TransferBack:   true,
		ExternalSender: false,
	})
	require.NoError(t, err)
	assert.Nil(t, i.GetOngoingSwapState(types.EmptyHash))
	assert.Equal(t, i.Provides(), types.ProvidesETH)
	_, err = i.Refund(types.EmptyHash)
	assert.ErrorIs(t, err, errNoOngoingSwap)
}
