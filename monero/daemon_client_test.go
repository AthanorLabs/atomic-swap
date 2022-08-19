package monero

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noot/atomic-swap/common"
)

func Test_GenerateBlocks(t *testing.T) {
	addr := "4BKjy1uVRTPiz4pHyaXXawb82XpzLiowSDd8rEQJGqvN6AD6kWosLQ6VJXW9sghopxXgQSh1RTd54JdvvCRsXiF41xvfeW5"
	cli := NewDaemonClient(common.DefaultMoneroDaemonEndpoint)

	for count := uint64(1); count <= 100; count += 10 {
		hdrBefore, err := cli.rpc.GetLastBlockHeader()
		require.NoError(t, err)

		err = cli.GenerateBlocks(addr, count)
		require.NoError(t, err)

		hdrAfter, err := cli.rpc.GetLastBlockHeader()
		require.NoError(t, err)
		require.GreaterOrEqual(t, hdrAfter.BlockHeader.Height, hdrBefore.BlockHeader.Height+count)
		// Normally, these values are equal, but a diff of more than 1 probably means another
		// call to GenerateBlocks created a longer chain and the blocks we generated were discarded.
		diff := hdrAfter.BlockHeader.Height - (hdrBefore.BlockHeader.Height + count)
		if diff > 1 {
			t.Logf("WARNING: requested %d blocks, but height difference is %d", count, diff)
		}
	}
}
