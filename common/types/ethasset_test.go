// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

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
	// Unmarshal string ETH
	asset := EthAsset(ethcommon.Address{0x1}) // any non-zero initial value to make sure we overwrite it
	err := json.Unmarshal([]byte(`"ETH"`), &asset)
	require.NoError(t, err)
	require.Equal(t, EthAssetETH, asset)

	// Unmarshal 0x prefixed address
	addr := "0xADd47138bb89c3013B39F2e3B062B408c90E5179"
	quotedAddr := fmt.Sprintf("%q", addr)
	err = json.Unmarshal([]byte(quotedAddr), &asset)
	require.NoError(t, err)
	expected := EthAsset(ethcommon.HexToAddress(addr))
	require.Equal(t, expected, asset)

	// Unmarshal address without the 0x prefix
	quotedAddr = fmt.Sprintf("%q", addr[2:])
	err = json.Unmarshal([]byte(quotedAddr), &asset)
	require.NoError(t, err)
	expected = EthAsset(ethcommon.HexToAddress(addr))
	require.Equal(t, expected, asset)

	// Unmarshal addresses with the ERC20@ prefix that our String() method
	// generates
	expected = EthAsset(ethcommon.HexToAddress("0xa1E32d14AC4B6d8c1791CAe8E9baD46a1E15B7a8"))
	quotedAddr = fmt.Sprintf("%q", expected.String()) // will have ERC20@ prefix
	err = json.Unmarshal([]byte(quotedAddr), &asset)
	require.NoError(t, err)
	require.Equal(t, expected, asset)
}

func TestEthAsset_UnmarshalText_fail(t *testing.T) {
	tooShortQuotedAddr := `"0xA9"`
	asset := EthAsset(ethcommon.Address{0x1})
	err := json.Unmarshal([]byte(tooShortQuotedAddr), &asset)
	require.ErrorContains(t, err, "invalid asset value")
}

func TestEthAsset_IsToken(t *testing.T) {
	require.True(t, EthAssetETH.IsETH())
	require.False(t, EthAssetETH.IsToken())

	token := EthAsset(ethcommon.HexToAddress("0xa1E32d14AC4B6d8c1791CAe8E9baD46a1E15B7a8"))
	require.False(t, token.IsETH())
	require.True(t, token.IsToken())
}
