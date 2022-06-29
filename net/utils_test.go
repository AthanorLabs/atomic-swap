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

func TestStringToAddrInfo(t *testing.T) {
	str := "/ip4/192.168.0.101/tcp/9934/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv"
	addrInfo, err := StringToAddrInfo(str)
	require.NoError(t, err)
	require.Equal(t, "12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv", addrInfo.ID.String())
}
