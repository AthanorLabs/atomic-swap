package db

import (
	"encoding/json"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"

	"github.com/ChainSafe/chaindb"
)

const (
	recoveryPrefix             = "recv"
	contractSwapInfoPrefix     = "ethinfo"
	moneroHeightPrefix         = "xmrheight"
	swapPrivateKeyPrefix       = "privkey"
	sharedSwapPrivateKeyPrefix = "sprivkey"
	relayerInfoPrefix          = "relayer"
)

// RecoveryDB contains information about ongoing swaps requires for recovery
// in case of shutdown.
type RecoveryDB struct {
	db chaindb.Database
}

func newRecoveryDB(db chaindb.Database) *RecoveryDB {
	return &RecoveryDB{
		db: db,
	}
}

func getRecoveryDBKey(id types.Hash, additional string) []byte {
	return append(id[:], []byte(additional)...)
}

func (db *RecoveryDB) close() error {
	return db.db.Close()
}

// PutSwapRelayerInfo ...
func (db *RecoveryDB) PutSwapRelayerInfo(id types.Hash, info *types.OfferExtra) error {
	val, err := json.Marshal(info)
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, relayerInfoPrefix)
	return db.db.Put(key, val)
}

// GetSwapRelayerInfo ...
func (db *RecoveryDB) GetSwapRelayerInfo(id types.Hash) (*types.OfferExtra, error) {
	key := getRecoveryDBKey(id, relayerInfoPrefix)
	value, err := db.db.Get(key)
	if err != nil {
		return nil, err
	}

	var s types.OfferExtra
	err = json.Unmarshal(value, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// PutContractSwapInfo stores the given contract swap ID (which is not the same as the daemon
// swap ID, but is instead a hash of the `SwapFactorySwap` structure)
// and contract swap structure for the given swap ID.
func (db *RecoveryDB) PutContractSwapInfo(id types.Hash, info *EthereumSwapInfo) error {
	val, err := json.Marshal(info)
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, contractSwapInfoPrefix)
	return db.db.Put(key, val)
}

// GetContractSwapInfo returns the contract swap ID (a hash of the `SwapFactorySwap` structure) and
// and contract swap structure for the given swap ID.
func (db *RecoveryDB) GetContractSwapInfo(id types.Hash) (*EthereumSwapInfo, error) {
	key := getRecoveryDBKey(id, contractSwapInfoPrefix)
	value, err := db.db.Get(key)
	if err != nil {
		return nil, err
	}

	var s EthereumSwapInfo
	err = json.Unmarshal(value, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// PutMoneroStartHeight stores the monero chain height at the start of the given swap.
func (db *RecoveryDB) PutMoneroStartHeight(id types.Hash, height uint64) error {
	val, err := json.Marshal(height)
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, moneroHeightPrefix)
	return db.db.Put(key, val)
}

// GetMoneroStartHeight ...
func (db *RecoveryDB) GetMoneroStartHeight(id types.Hash) (uint64, error) {
	key := getRecoveryDBKey(id, moneroHeightPrefix)
	value, err := db.db.Get(key)
	if err != nil {
		return 0, err
	}

	var s uint64
	err = json.Unmarshal(value, &s)
	if err != nil {
		return 0, err
	}

	return s, nil
}

// PutSwapPrivateKey stores the given ephemeral swap private key share for the given swap ID.
func (db *RecoveryDB) PutSwapPrivateKey(id types.Hash, keys *mcrypto.PrivateKeyPair, env common.Environment) error {
	k := keys.Info(env)
	val, err := json.Marshal(k)
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, swapPrivateKeyPrefix)
	return db.db.Put(key[:], val)
}

// GetSwapPrivateKey returns the swap private key share, if it exists.
func (db *RecoveryDB) GetSwapPrivateKey(id types.Hash) (*mcrypto.PrivateKeyPair, error) {
	key := getRecoveryDBKey(id, swapPrivateKeyPrefix)
	value, err := db.db.Get(key[:])
	if err != nil {
		return nil, err
	}

	var info mcrypto.PrivateKeyInfo
	err = json.Unmarshal(value, &info)
	if err != nil {
		return nil, err
	}

	return mcrypto.NewPrivateKeyPairFromHex(info.PrivateSpendKey, info.PrivateViewKey)
}

// PutSharedSwapPrivateKey stores the shared swap private key for the given swap ID.
func (db *RecoveryDB) PutSharedSwapPrivateKey(
	id types.Hash,
	keys *mcrypto.PrivateKeyPair,
	env common.Environment,
) error {
	k := keys.Info(env)
	val, err := json.Marshal(k)
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, sharedSwapPrivateKeyPrefix)
	return db.db.Put(key[:], val)
}

// GetSharedSwapPrivateKey returns the shared swap private key, if it exists.
func (db *RecoveryDB) GetSharedSwapPrivateKey(id types.Hash) (*mcrypto.PrivateKeyPair, error) {
	key := getRecoveryDBKey(id, sharedSwapPrivateKeyPrefix)
	value, err := db.db.Get(key[:])
	if err != nil {
		return nil, err
	}

	var info mcrypto.PrivateKeyInfo
	err = json.Unmarshal(value, &info)
	if err != nil {
		return nil, err
	}

	return mcrypto.NewPrivateKeyPairFromHex(info.PrivateSpendKey, info.PrivateViewKey)
}

// DeleteSwap deletes all recovery info from the db for the given swap.
func (db *RecoveryDB) DeleteSwap(id types.Hash) error {
	keys := [][]byte{
		getRecoveryDBKey(id, contractSwapInfoPrefix),
		getRecoveryDBKey(id, moneroHeightPrefix),
		getRecoveryDBKey(id, swapPrivateKeyPrefix),
		getRecoveryDBKey(id, sharedSwapPrivateKeyPrefix),
	}

	for _, key := range keys {
		err := db.db.Del(key)
		if err != nil {
			return err
		}
	}

	return nil
}
