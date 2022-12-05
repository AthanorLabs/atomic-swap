package net

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
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
	str := "/ip4/192.168.0.101/udp/9934/quic/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv"
	addrInfo, err := StringToAddrInfo(str)
	require.NoError(t, err)
	require.Equal(t, "12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv", addrInfo.ID.String())
}

func Test_getPubIP(t *testing.T) {
	ip, err := getPubIP()
	require.NoError(t, err)
	// simple sanity check regex (not a full-blown validator)
	assert.Regexp(t, regexp.MustCompile(`^(\d+.){3}\d+$`), ip)
}
