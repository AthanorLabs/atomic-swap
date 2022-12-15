package net

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateAndSaveKey(t *testing.T) {
	tempDir := t.TempDir()
	_, err := generateKey(1234, tempDir)
	require.NoError(t, err)

	_, err = generateKey(1234, tempDir)
	require.NoError(t, err)
}

func Test_stringsToAddrInfos(t *testing.T) {
	bootnodes := []string{
		"/ip4/192.168.0.101/udp/9934/quic-v1/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv",
		"/ip4/192.168.0.101/tcp/9934/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv",
	}
	addrInfos, err := stringsToAddrInfos(bootnodes)
	require.NoError(t, err)
	require.Len(t, addrInfos, 1) // both were combined into one AddrInfo
	require.Equal(t, "12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv",
		addrInfos[0].ID.String())
	require.Len(t, addrInfos[0].Addrs, 2)
	require.Equal(t, "/ip4/192.168.0.101/udp/9934/quic", addrInfos[0].Addrs[0].String())
	require.Equal(t, "/ip4/192.168.0.101/tcp/9934", addrInfos[0].Addrs[1].String())
}
