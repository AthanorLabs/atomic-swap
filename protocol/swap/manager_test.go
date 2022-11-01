package swap

import (
	"testing"

	"github.com/athanorlabs/atomic-swap/common/types"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	// test that creating a new manager loads all on-disk swaps
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := NewMockDatabase(ctrl)

	db.EXPECT().GetAllSwaps()

	m, err := NewManager(db)
	require.NoError(t, err)

	hashA := types.Hash{0x1}
	infoA := NewInfo(hashA, types.ProvidesXMR, 1, 1, 0.1, types.EthAssetETH, types.ExpectingKeys, nil)
	db.EXPECT().PutSwap(infoA)
	err = m.AddSwap(infoA)
	require.NoError(t, err)

	infoB := NewInfo(types.Hash{0x2}, types.ProvidesXMR, 1, 1, 0.1, types.EthAssetETH, types.CompletedSuccess, nil)
	db.EXPECT().PutSwap(infoB)
	err = m.AddSwap(infoB)
	require.NoError(t, err)

	db.EXPECT().GetAllSwaps().Return([]*Info{infoA, infoB}, nil)
	m, err = NewManager(db)
	require.NoError(t, err)
	require.Equal(t, 1, len(m.ongoing))
	require.Equal(t, infoA, m.ongoing[hashA])
}

func TestManager_AddSwap_Ongoing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := NewMockDatabase(ctrl)

	db.EXPECT().GetAllSwaps()

	m, err := NewManager(db)
	require.NoError(t, err)
	info := NewInfo(types.Hash{}, types.ProvidesXMR, 1, 1, 0.1, types.EthAssetETH, types.ExpectingKeys, nil)

	db.EXPECT().PutSwap(info)
	err = m.AddSwap(info)
	require.NoError(t, err)
	db.EXPECT().PutSwap(info)
	err = m.AddSwap(info)
	require.NoError(t, err)

	s, err := m.GetOngoingSwap(types.Hash{})
	require.NoError(t, err)
	require.Equal(t, info, s)
	require.NotNil(t, m.ongoing)

	db.EXPECT().PutSwap(info)
	m.CompleteOngoingSwap(types.Hash{})
	require.Equal(t, 0, len(m.ongoing))

	db.EXPECT().GetAllSwaps()
	ids, err := m.GetPastIDs()
	require.NoError(t, err)
	require.Equal(t, []types.Hash{{}}, ids)

	m.CompleteOngoingSwap(types.Hash{})
}

func TestManager_AddSwap_Past(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := NewMockDatabase(ctrl)

	db.EXPECT().GetAllSwaps()

	m, err := NewManager(db)
	require.NoError(t, err)

	info := &Info{
		ID:     types.Hash{1},
		Status: types.CompletedSuccess,
	}

	db.EXPECT().PutSwap(info)
	err = m.AddSwap(info)
	require.NoError(t, err)

	s, err := m.GetPastSwap(info.ID)
	require.NoError(t, err)
	require.NotNil(t, s)

	info = &Info{
		ID:     types.Hash{2},
		Status: types.CompletedSuccess,
	}

	db.EXPECT().PutSwap(info)
	err = m.AddSwap(info)
	require.NoError(t, err)

	s, err = m.GetPastSwap(info.ID)
	require.NoError(t, err)
	require.NotNil(t, s)

	db.EXPECT().GetAllSwaps()
	ids, err := m.GetPastIDs()
	require.NoError(t, err)
	require.Equal(t, 2, len(ids))
}
