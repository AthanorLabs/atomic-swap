package types

import (
	"encoding/json"
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestEthAsset_MarshalText(t *testing.T) {
	jsonData, err := json.Marshal(EthAssetETH)
	require.NoError(t, err)
	require.Equal(t, `"ETH"`, string(jsonData))

	addr := "0xADd47138bb89c3013B39F2e3B062B408c90E5179"
	asset := EthAsset(ethcommon.HexToAddress(addr))
	jsonData, err = json.Marshal(asset)
	require.NoError(t, err)
	quotedAddr := fmt.Sprintf("%q", addr)
	require.Equal(t, quotedAddr, string(jsonData))
}

func TestEthAsset_UnmarshalText(t *testing.T) {
	asset := EthAsset(ethcommon.Address{0x1}) // any non-zero initial value to make sure we overwrite it
	err := json.Unmarshal([]byte(`"ETH"`), &asset)
	require.NoError(t, err)
	require.Equal(t, EthAssetETH, asset)

	addr := "0xADd47138bb89c3013B39F2e3B062B408c90E5179"
	quotedAddr := fmt.Sprintf("%q", addr)
	err = json.Unmarshal([]byte(quotedAddr), &asset)
	require.NoError(t, err)
	expected := EthAsset(ethcommon.HexToAddress(addr))
	require.Equal(t, expected, asset)

	// Same exact test as above, but without the 0x prefix
	quotedAddr = fmt.Sprintf("%q", addr[2:])
	err = json.Unmarshal([]byte(quotedAddr), &asset)
	require.NoError(t, err)
	expected = EthAsset(ethcommon.HexToAddress(addr))
	require.Equal(t, expected, asset)
}

func TestEthAsset_UnmarshalText_fail(t *testing.T) {
	tooShortQuotedAddr := `"0xA9"`
	asset := EthAsset(ethcommon.Address{0x1})
	err := json.Unmarshal([]byte(tooShortQuotedAddr), &asset)
	require.ErrorContains(t, err, "invalid asset value")
}
