package db

import (
	"encoding/json"
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/ChainSafe/chaindb"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
)

func newTestRecoveryDB(t *testing.T) *RecoveryDB {
	cfg := &chaindb.Config{
		DataDir:  t.TempDir(),
		InMemory: true,
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)

	return db.recoveryDB
}

func TestRecoveryDB_ContractSwapInfo(t *testing.T) {
	rdb := newTestRecoveryDB(t)

	offerID := types.Hash{5, 6, 7, 8}
	si := &EthereumSwapInfo{
		StartNumber: big.NewInt(12345),
		SwapID:      types.Hash{1, 2, 3, 4},
		Swap: contracts.SwapFactorySwap{
			Owner:        ethcommon.HexToAddress("0xda9dfa130df4de4673b89022ee50ff26f6ea73cf"),
			Claimer:      ethcommon.HexToAddress("0xbe0eb53f46cd790cd13851d5eff43d12404d33e8"),
			PubKeyClaim:  ethcommon.HexToHash("0x5ab9467e70d4e98567991f0179d1f82a3096ed7973f7aff9ea50f649cafa88b9"),
			PubKeyRefund: ethcommon.HexToHash("0x4897bc3b9e02c2a8cd6353b9b29377157bf2694daaf52b59c0b42daa39877f14"),
			Timeout0:     big.NewInt(1672531200),
			Timeout1:     big.NewInt(1672545600),
			Asset:        types.EthAssetETH.Address(),
			Value:        big.NewInt(9876),
			Nonce:        big.NewInt(1234),
		},
		ContractAddress: ethcommon.HexToAddress("0xd2b5d6252d0645e4cf4bb547e82a485f527befb7"),
	}

	expectedStr := `{
		"start_number": 12345,
		"swap_id":      "0x0102030400000000000000000000000000000000000000000000000000000000",
		"swap": {
			"owner":          "0xda9dfa130df4de4673b89022ee50ff26f6ea73cf",
			"claimer":        "0xbe0eb53f46cd790cd13851d5eff43d12404d33e8",
			"pub_key_claim":  "0x5ab9467e70d4e98567991f0179d1f82a3096ed7973f7aff9ea50f649cafa88b9",
			"pub_key_refund": "0x4897bc3b9e02c2a8cd6353b9b29377157bf2694daaf52b59c0b42daa39877f14",
			"timeout0":       1672531200,
			"timeout1":       1672545600,
			"asset":          "0x0000000000000000000000000000000000000000",
			"value":          9876,
			"nonce":          1234
		},
		"contract_address": "0xd2b5d6252d0645e4cf4bb547e82a485f527befb7"
	}`
	jsonData, err := json.Marshal(si)
	require.NoError(t, err)
	require.JSONEq(t, expectedStr, string(jsonData))

	err = rdb.PutContractSwapInfo(offerID, si)
	require.NoError(t, err)

	res, err := rdb.GetContractSwapInfo(offerID)
	require.NoError(t, err)
	require.Equal(t, si, res)
}

func TestRecoveryDB_SwapRelayerInfo(t *testing.T) {
	rdb := newTestRecoveryDB(t)
	offerID := types.Hash{5, 6, 7, 8}

	commission := coins.StrToDecimal("0.0135")
	info := &types.OfferExtra{
		RelayerEndpoint:   "endpoint",
		RelayerCommission: commission,
	}

	err := rdb.PutSwapRelayerInfo(offerID, info)
	require.NoError(t, err)

	res, err := rdb.GetSwapRelayerInfo(offerID)
	require.NoError(t, err)
	require.Equal(t, info, res)
}

func TestRecoveryDB_SwapPrivateKey(t *testing.T) {
	rdb := newTestRecoveryDB(t)
	offerID := types.Hash{5, 6, 7, 8}

	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	err = rdb.PutSwapPrivateKey(offerID, kp.SpendKey())
	require.NoError(t, err)

	res, err := rdb.GetSwapPrivateKey(offerID)
	require.NoError(t, err)
	require.Equal(t, kp.SpendKey().Hex(), res.Hex())
}

func TestRecoveryDB_SharedSwapPrivateKey(t *testing.T) {
	rdb := newTestRecoveryDB(t)
	offerID := types.Hash{5, 6, 7, 8}

	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	err = rdb.PutCounterpartySwapPrivateKey(offerID, kp.SpendKey())
	require.NoError(t, err)

	res, err := rdb.GetCounterpartySwapPrivateKey(offerID)
	require.NoError(t, err)
	require.Equal(t, kp.SpendKey().Hex(), res.Hex())
}

func TestRecoveryDB_XMRMakerSwapKeys(t *testing.T) {
	rdb := newTestRecoveryDB(t)
	offerID := types.Hash{5, 6, 7, 8}

	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	err = rdb.PutXMRMakerSwapKeys(offerID, kp.SpendKey().Public(), kp.ViewKey())
	require.NoError(t, err)

	resSk, resVk, err := rdb.GetXMRMakerSwapKeys(offerID)
	require.NoError(t, err)
	require.Equal(t, kp.SpendKey().Public().Hex(), resSk.Hex())
	require.Equal(t, kp.ViewKey().Hex(), resVk.Hex())
}

func TestRecoveryDB_XMRTakerSwapKeys(t *testing.T) {
	rdb := newTestRecoveryDB(t)
	offerID := types.Hash{5, 6, 7, 8}

	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	err = rdb.PutXMRTakerSwapKeys(offerID, kp.PublicKeyPair())
	require.NoError(t, err)

	res, err := rdb.GetXMRTakerSwapKeys(offerID)
	require.NoError(t, err)
	require.Equal(t, res.SpendKey().Hex(), kp.SpendKey().Public().Hex())
	require.Equal(t, res.ViewKey().Hex(), kp.ViewKey().Public().Hex())
}

func TestRecoveryDB_DeleteSwap(t *testing.T) {
	rdb := newTestRecoveryDB(t)
	offerID := types.Hash{5, 6, 7, 8}

	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	si := &EthereumSwapInfo{
		StartNumber: big.NewInt(12345),
		SwapID:      types.Hash{1, 2, 3, 4},
		Swap: contracts.SwapFactorySwap{
			Owner:        ethcommon.HexToAddress("0xda9dfa130df4de4673b89022ee50ff26f6ea73cf"),
			Claimer:      ethcommon.HexToAddress("0xbe0eb53f46cd790cd13851d5eff43d12404d33e8"),
			PubKeyClaim:  ethcommon.HexToHash("0x5ab9467e70d4e98567991f0179d1f82a3096ed7973f7aff9ea50f649cafa88b9"),
			PubKeyRefund: ethcommon.HexToHash("0x4897bc3b9e02c2a8cd6353b9b29377157bf2694daaf52b59c0b42daa39877f14"),
			Timeout0:     big.NewInt(1672531200),
			Timeout1:     big.NewInt(1672545600),
			Asset:        types.EthAssetETH.Address(),
			Value:        big.NewInt(9876),
			Nonce:        big.NewInt(1234),
		},
		ContractAddress: ethcommon.HexToAddress("0xd2b5d6252d0645e4cf4bb547e82a485f527befb7"),
	}

	info := &types.OfferExtra{
		RelayerEndpoint:   "endpoint",
		RelayerCommission: coins.StrToDecimal("0.0135"),
	}

	err = rdb.PutContractSwapInfo(offerID, si)
	require.NoError(t, err)
	err = rdb.PutSwapRelayerInfo(offerID, info)
	require.NoError(t, err)
	err = rdb.PutSwapPrivateKey(offerID, kp.SpendKey())
	require.NoError(t, err)
	err = rdb.PutCounterpartySwapPrivateKey(offerID, kp.SpendKey())
	require.NoError(t, err)
	err = rdb.PutXMRMakerSwapKeys(offerID, kp.SpendKey().Public(), kp.ViewKey())
	require.NoError(t, err)

	err = rdb.DeleteSwap(offerID)
	require.NoError(t, err)
	_, err = rdb.GetContractSwapInfo(offerID)
	require.EqualError(t, chaindb.ErrKeyNotFound, err.Error())
	_, err = rdb.GetSwapRelayerInfo(offerID)
	require.EqualError(t, chaindb.ErrKeyNotFound, err.Error())
	_, err = rdb.GetSwapPrivateKey(offerID)
	require.EqualError(t, chaindb.ErrKeyNotFound, err.Error())
	_, err = rdb.GetCounterpartySwapPrivateKey(offerID)
	require.EqualError(t, chaindb.ErrKeyNotFound, err.Error())
	_, _, err = rdb.GetXMRMakerSwapKeys(offerID)
	require.EqualError(t, chaindb.ErrKeyNotFound, err.Error())
}
