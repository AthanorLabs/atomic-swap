package swap

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
)

func TestStatusManager(t *testing.T) {
	offerID1 := types.Hash{0x1}

	statusMgr := newStatusManager()
	ch1 := statusMgr.GetStatusChan(offerID1)
	ch2 := statusMgr.GetStatusChan(offerID1)
	require.Equal(t, ch1, ch2)

	statusMgr.PushNewStatus(offerID1, types.CompletedSuccess)
	status := <-ch1
	require.Equal(t, types.CompletedSuccess, status)

	statusMgr.DeleteStatusChan(offerID1)
	ch3 := statusMgr.GetStatusChan(offerID1)
	require.NotEqual(t, ch1, ch3)
}
