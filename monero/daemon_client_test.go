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
	for i := 0; i < 100; i++ {
		resp, err := cli.generateBlocks(addr, 1)
		require.NoError(t, err)
		require.Greater(t, resp.Height, prevHeight)
		prevHeight = resp.Height
	}
}
