package net

import (
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common/types"

	"github.com/stretchr/testify/require"
)

func TestHost_Initiate(t *testing.T) {
	ha := newHost(t, basicTestConfig(t))
	err := ha.Start()
	require.NoError(t, err)
	hb := newHost(t, basicTestConfig(t))
	err = hb.Start()
	require.NoError(t, err)

	err = ha.h.Connect(ha.ctx, hb.h.AddrInfo())
	require.NoError(t, err)

	err = ha.Initiate(hb.h.AddrInfo(), &SendKeysMessage{}, new(mockSwapState))
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 500)
	require.NotNil(t, ha.swaps[testID])
	require.NotNil(t, hb.swaps[testID])
}

func TestHost_ConcurrentSwaps(t *testing.T) {
	ha := newHost(t, basicTestConfig(t))
	err := ha.Start()
	require.NoError(t, err)

	hbCfg := basicTestConfig(t)
	hbCfg.Bootnodes = ha.h.Addresses() // get some test coverage on our bootnode code
	hb := newHost(t, hbCfg)
	err = hb.Start()
	require.NoError(t, err)

	testID2 := types.Hash{98}

	err = ha.h.Connect(ha.ctx, hb.h.AddrInfo())
	require.NoError(t, err)

	err = ha.Initiate(hb.h.AddrInfo(), &SendKeysMessage{}, new(mockSwapState))
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 500)
	require.NotNil(t, ha.swaps[testID])
	require.NotNil(t, hb.swaps[testID])

	hb.handler.(*mockHandler).id = testID2

	err = ha.Initiate(hb.h.AddrInfo(), &SendKeysMessage{}, &mockSwapState{testID2})
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 1500)
	require.NotNil(t, ha.swaps[testID2])
	require.NotNil(t, hb.swaps[testID2])
}
