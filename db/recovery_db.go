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
	swapPrivateKeyPrefix       = "privkey"
	sharedSwapPrivateKeyPrefix = "sprivkey"
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

// PutContractSwapInfo stores the given contract swap ID (which is not the same as the daemon
// swap ID, but is instead a hash of the `SwapFactorySwap` structure)
// and contract swap structure for the given swap ID.
func (db *RecoveryDB) PutContractSwapInfo(id types.Hash, info *EthereumSwapInfo) error {
	val, err := json.Marshal(info)
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, contractSwapInfoPrefix)
	return db.db.Put(key[:], val)
}

// GetContractSwapInfo returns the contract swap ID (a hash of the `SwapFactorySwap` structure) and
// and contract swap structure for the given swap ID.
func (db *RecoveryDB) GetContractSwapInfo(id types.Hash) (*EthereumSwapInfo, error) {
	key := getRecoveryDBKey(id, contractSwapInfoPrefix)
	value, err := db.db.Get(key[:])
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
