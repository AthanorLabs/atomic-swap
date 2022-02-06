package alice

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAlice_InitiateProtocol(t *testing.T) {
	a := newTestAlice(t)
	s, err := a.InitiateProtocol(3.33)
	require.NoError(t, err)
	require.Equal(t, a.swapState, s)
}
