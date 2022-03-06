package swap

import (
	"testing"

	"github.com/noot/atomic-swap/common/types"

	"github.com/stretchr/testify/require"
)

func TestManager_AddSwap_Ongoing(t *testing.T) {
	m := NewManager()
	info := NewInfo(types.ProvidesXMR, 1, 1, 0.1, types.ExpectingKeys, nil)

	err := m.AddSwap(info)
	require.NoError(t, err)
	err = m.AddSwap(info)
	require.Equal(t, errHaveOngoingSwap, err)
	require.Equal(t, info, m.GetOngoingSwap())
	require.NotNil(t, m.ongoing)

	m.CompleteOngoingSwap()
	require.Nil(t, m.ongoing)
	require.Equal(t, []uint64{0}, m.GetPastIDs())
	require.Equal(t, uint64(1), nextID)

	m.CompleteOngoingSwap()
}

func TestManager_AddSwap_Past(t *testing.T) {
	m := NewManager()

	info := &Info{
		id:     1,
		status: types.CompletedSuccess,
	}

	err := m.AddSwap(info)
	require.NoError(t, err)
	require.NotNil(t, m.GetPastSwap(1))

	info = &Info{
		id:     2,
		status: types.CompletedSuccess,
	}

	err = m.AddSwap(info)
	require.NoError(t, err)
	require.NotNil(t, m.GetPastSwap(2))

	ids := m.GetPastIDs()
	require.Equal(t, 2, len(ids))
}
