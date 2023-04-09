// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package swap

import (
	"fmt"
	"testing"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/common/vjson"
)

var testPeerID, _ = peer.Decode("12D3KooWQQRJuKTZ35eiHGNPGDpQqjpJSdaxEMJRxi6NWFrrvQVi")

func Test_InfoMarshal(t *testing.T) {
	offerIDStr := "0x0102030405060708091011121314151617181920212223242526272829303132"
	offerID := ethcommon.HexToHash(offerIDStr)
	info := NewInfo(
		testPeerID,
		offerID,
		coins.ProvidesXMR,
		apd.New(125, -2), // 1.25
		apd.New(1, 0),
		coins.ToExchangeRate(apd.New(33, -2)), // 0.33
		types.EthAssetETH,
		types.CompletedSuccess,
		200,
		make(chan types.Status),
	)
	err := info.StartTime.UnmarshalJSON([]byte("\"2023-02-20T17:29:43.471020297-05:00\""))
	require.NoError(t, err)
	info.LastStatusUpdateTime = info.StartTime

	infoBytes, err := vjson.MarshalStruct(info)
	require.NoError(t, err)

	expectedJSON := `{
		"version": "0.3.0",
		"peerID": "12D3KooWQQRJuKTZ35eiHGNPGDpQqjpJSdaxEMJRxi6NWFrrvQVi",
		"offerID": "0x0102030405060708091011121314151617181920212223242526272829303132",
		"provides": "XMR",
		"providedAmount": "1.25",
		"expectedAmount": "1",
		"exchangeRate": "0.33",
		"ethAsset": "ETH",
		"moneroStartHeight": 200,
		"status": "Success",
		"lastStatusUpdateTime": "2023-02-20T17:29:43.471020297-05:00",
		"startTime": "2023-02-20T17:29:43.471020297-05:00"
	}`
	require.JSONEq(t, expectedJSON, string(infoBytes))
}

func TestUnmarshalInfo_missingVersion(t *testing.T) {
	_, err := UnmarshalInfo([]byte(`{}`))
	require.ErrorIs(t, err, errInfoVersionMissing)
}

func TestUnmarshalInfo_versionTooNew(t *testing.T) {
	unsupportedVersion := CurInfoVersion.IncMajor()
	offerJSON := fmt.Sprintf(`{
		"version": "%s",
		"some_unsupported_field": ""
	}`, unsupportedVersion)
	_, err := UnmarshalInfo([]byte(offerJSON))
	require.ErrorContains(t, err, fmt.Sprintf("info version %q not supported", unsupportedVersion))
}
