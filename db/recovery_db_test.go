package db

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"

	"github.com/ChainSafe/chaindb"
	"github.com/stretchr/testify/require"
)

func TestRecoveryDB_ContractSwapInfo(t *testing.T) {
	cfg := &chaindb.Config{
		DataDir:  t.TempDir(),
		InMemory: true,
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)

	rdb := db.recoveryDB

	offerID := types.Hash{5, 6, 7, 8}
	si := &swapInfo{
		SwapID: types.Hash{1, 2, 3, 4},
		Swap: &contracts.SwapFactorySwap{
			Owner:   [20]byte{9, 0xa, 0xb, 0xc},
			Claimer: [20]byte{0xd, 0xe, 0xf, 0},
			Value:   big.NewInt(999),
			Nonce:   big.NewInt(888),
		},
	}

	expectedStr := `{"swapID":"0x0102030400000000000000000000000000000000000000000000000000000000","swap":{"Owner":"0x090a0b0c00000000000000000000000000000000","Claimer":"0x0d0e0f0000000000000000000000000000000000","PubKeyClaim":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"PubKeyRefund":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"Timeout0":null,"Timeout1":null,"Asset":"0x0000000000000000000000000000000000000000","Value":999,"Nonce":888}}`
	val, err := json.Marshal(si)
	require.NoError(t, err)
	require.Equal(t, expectedStr, string(val))

	err = rdb.PutContractSwapInfo(offerID, si.SwapID, si.Swap)
	require.NoError(t, err)

	resSwapID, resSwap, err := rdb.GetContractSwapInfo(offerID)
	require.NoError(t, err)
	require.Equal(t, [32]byte(si.SwapID), resSwapID)
	require.Equal(t, si.Swap, resSwap)
}
