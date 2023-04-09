// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/vjson"
)

func TestSwapCreatorSwap_JSON(t *testing.T) {
	sf := &SwapCreatorSwap{
		Owner:        ethcommon.HexToAddress("0xda9dfa130df4de4673b89022ee50ff26f6ea73cf"),
		Claimer:      ethcommon.HexToAddress("0xbe0eb53f46cd790cd13851d5eff43d12404d33e8"),
		PubKeyClaim:  ethcommon.HexToHash("0x5ab9467e70d4e98567991f0179d1f82a3096ed7973f7aff9ea50f649cafa88b9"),
		PubKeyRefund: ethcommon.HexToHash("0x4897bc3b9e02c2a8cd6353b9b29377157bf2694daaf52b59c0b42daa39877f14"),
		Timeout0:     big.NewInt(1672531200),
		Timeout1:     big.NewInt(1672545600),
		Asset:        ethcommon.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"),
		Value:        coins.EtherToWei(apd.New(9876, 0)).BigInt(),
		Nonce:        big.NewInt(1234),
	}
	expectedJSON := `{
		"owner": "0xda9dfa130df4de4673b89022ee50ff26f6ea73cf",
		"claimer": "0xbe0eb53f46cd790cd13851d5eff43d12404d33e8",
		"pubKeyClaim": "0x5ab9467e70d4e98567991f0179d1f82a3096ed7973f7aff9ea50f649cafa88b9",
		"pubKeyRefund": "0x4897bc3b9e02c2a8cd6353b9b29377157bf2694daaf52b59c0b42daa39877f14",
		"timeout0": 1672531200,
		"timeout1": 1672545600,
		"asset": "0xdac17f958d2ee523a2206206994597c13d831ec7",
		"value": 9876000000000000000000,
		"nonce": 1234
	}`
	jsonData, err := vjson.MarshalStruct(sf)
	require.NoError(t, err)
	require.JSONEq(t, expectedJSON, string(jsonData))

	sf2 := &SwapCreatorSwap{}
	err = json.Unmarshal(jsonData, sf2)
	require.NoError(t, err)
	require.EqualValues(t, sf, sf2)
}

// Ensure that our serializable swap type has the same number of fields as the original
// generated type.
func TestSwapCreatorSwap_JSON_fieldCountEqual(t *testing.T) {
	numSwapFields := reflect.TypeOf(swap{}).NumField()
	numSwapCreatorSwapFields := reflect.TypeOf(SwapCreatorSwap{}).NumField()
	require.Equal(t, numSwapCreatorSwapFields, numSwapFields)
}
