package swap

import (
	"testing"

	"github.com/athanorlabs/atomic-swap/common/types"

	"github.com/stretchr/testify/require"
)

func TestManager_AddSwap_Ongoing(t *testing.T) {
	m := NewManager().(*manager)
	info := NewInfo(types.Hash{}, types.ProvidesXMR, 1, 1, 0.1, types.ExpectingKeys, nil)

	err := m.AddSwap(info)
	require.NoError(t, err)
	err = m.AddSwap(info)
	require.NoError(t, err)
	require.Equal(t, info, m.GetOngoingSwap(types.Hash{}))
	require.NotNil(t, m.ongoing)

	m.CompleteOngoingSwap(types.Hash{})
	require.Equal(t, 0, len(m.ongoing))
	require.Equal(t, []types.Hash{{}}, m.GetPastIDs())

	m.CompleteOngoingSwap(types.Hash{})
}

func TestManager_AddSwap_Past(t *testing.T) {
	m := NewManager().(*manager)

	info := &Info{
		id:     types.Hash{1},
		status: types.CompletedSuccess,
	}

	err := m.AddSwap(info)
	require.NoError(t, err)
	require.NotNil(t, m.GetPastSwap(types.Hash{1}))

	info = &Info{
		id:     types.Hash{2},
		status: types.CompletedSuccess,
	}

	err = m.AddSwap(info)
	require.NoError(t, err)
	require.NotNil(t, m.GetPastSwap(types.Hash{2}))

	ids := m.GetPastIDs()
	require.Equal(t, 2, len(ids))
}
