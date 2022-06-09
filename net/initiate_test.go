package net

import (
	"testing"
	"time"

	"github.com/noot/atomic-swap/common/types"

	"github.com/stretchr/testify/require"
)

func TestHost_Initiate(t *testing.T) {
	ha := newHost(t, defaultPort)
	err := ha.Start()
	require.NoError(t, err)
	hb := newHost(t, defaultPort+1)
	err = hb.Start()
	require.NoError(t, err)

	defer func() {
		_ = ha.Stop()
		_ = hb.Stop()
	}()

	err = ha.h.Connect(ha.ctx, hb.addrInfo())
	require.NoError(t, err)

	err = ha.Initiate(hb.addrInfo(), &SendKeysMessage{}, new(mockSwapState))
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 500)
	require.NotNil(t, ha.swaps[testID])
	require.NotNil(t, hb.swaps[testID])
}

func TestHost_ConcurrentSwaps(t *testing.T) {
	ha := newHost(t, defaultPort)
	err := ha.Start()
	require.NoError(t, err)
	hb := newHost(t, defaultPort+1)
	err = hb.Start()
	require.NoError(t, err)

	testID2 := types.Hash{98}

	defer func() {
		_ = ha.Stop()
		_ = hb.Stop()
	}()

	err = ha.h.Connect(ha.ctx, hb.addrInfo())
	require.NoError(t, err)

	err = ha.Initiate(hb.addrInfo(), &SendKeysMessage{}, new(mockSwapState))
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 500)
	require.NotNil(t, ha.swaps[testID])
	require.NotNil(t, hb.swaps[testID])

	hb.handler.(*mockHandler).id = testID2

	err = ha.Initiate(hb.addrInfo(), &SendKeysMessage{}, &mockSwapState{testID2})
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 1500)
	require.NotNil(t, ha.swaps[testID2])
	require.NotNil(t, hb.swaps[testID2])
}
