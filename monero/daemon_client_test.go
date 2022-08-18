package monero

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noot/atomic-swap/common"
)

func Test_GenerateBlocks(t *testing.T) {
	addr := "4BKjy1uVRTPiz4pHyaXXawb82XpzLiowSDd8rEQJGqvN6AD6kWosLQ6VJXW9sghopxXgQSh1RTd54JdvvCRsXiF41xvfeW5"
	cli := NewDaemonClient(common.DefaultMoneroDaemonEndpoint)
	hdr, err := cli.rpc.GetLastBlockHeader()
	require.NoError(t, err)
	prevHeight := hdr.BlockHeader.Height
	for count := uint64(1); count <= 10; count++ {
		resp, err := cli.generateBlocks(addr, count)
		require.NoError(t, err)
		require.GreaterOrEqual(t, resp.Height, prevHeight+count)
		prevHeight = resp.Height
	}
}
