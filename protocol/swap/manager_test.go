package swap

import (
	"testing"

	"github.com/cockroachdb/apd/v3"

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
	infoA := NewInfo(
		hashA,
		types.ProvidesXMR,
		apd.New(1, 0),
		apd.New(10, 0),
		(*types.ExchangeRate)(apd.New(1, -1)), // 0.1
		types.EthAssetETH,
		types.ExpectingKeys,
		100,
		nil,
	)
	db.EXPECT().PutSwap(infoA)
	err = m.AddSwap(infoA)
	require.NoError(t, err)

	infoB := NewInfo(
		types.Hash{2},
		types.ProvidesXMR,
		apd.New(1, 0),
		apd.New(10, 0),
		(*types.ExchangeRate)(apd.New(1, -1)), // 0.1
		types.EthAssetETH,
		types.CompletedSuccess,
		100,
		nil,
	)
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
	info := NewInfo(
		types.Hash{},
		types.ProvidesXMR,
		apd.New(1, 0),
		apd.New(10, 0),
		(*types.ExchangeRate)(apd.New(1, -1)), // 0.1
		types.EthAssetETH,
		types.ExpectingKeys,
		100,
		nil,
	)

	db.EXPECT().PutSwap(info)
	err = m.AddSwap(info)
	require.NoError(t, err)
	db.EXPECT().PutSwap(info)
	err = m.AddSwap(info)
	require.NoError(t, err)

	s, err := m.GetOngoingSwap(types.Hash{})
	require.NoError(t, err)
	require.Equal(t, info, &s)
	require.NotNil(t, m.ongoing)

	db.EXPECT().PutSwap(info)
	err = m.CompleteOngoingSwap(info)
	require.NoError(t, err)
	require.Equal(t, 0, len(m.ongoing))

	db.EXPECT().GetAllSwaps()
	ids, err := m.GetPastIDs()
	require.NoError(t, err)
	require.Equal(t, []types.Hash{{}}, ids)

	//err = m.CompleteOngoingSwap(info)
	//require.NoError(t, err)
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

	info = &Info{
		ID:     types.Hash{3},
		Status: types.ExpectingKeys,
	}

	db.EXPECT().PutSwap(info)
	err = m.AddSwap(info)
	require.NoError(t, err)

	db.EXPECT().GetAllSwaps()
	ids, err := m.GetPastIDs()
	require.NoError(t, err)
	require.Equal(t, 2, len(ids))
}
