package db

import (
	"encoding/json"

	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"

	"github.com/ChainSafe/chaindb"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const (
	recoveryPrefix             = "recv"
	contractAddrPrefix         = "addr"
	contractSwapInfoPrefix     = "contract"
	swapPrivateKeyPrefix       = "privkey"
	sharedSwapPrivateKeyPrefix = "sprivkey"
)

// RecoveryInfo ...
type RecoveryInfo struct {
	ContractAddress      string
	ContractSwapID       [32]byte
	ContractSwap         contracts.SwapFactorySwap
	PrivateKeyInfo       *mcrypto.PrivateKeyInfo
	SharedSwapPrivateKey *mcrypto.PrivateKeyInfo
}

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

// PutContractAddress stores the given contract address for the given swap ID.
func (db *RecoveryDB) PutContractAddress(id types.Hash, addr ethcommon.Address) error {
	val, err := json.Marshal(addr)
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, contractAddrPrefix)
	return db.db.Put(key[:], val)
}

// PutContractAddress stores the given contract swap ID (which is not the same as the daemon
// swap ID, but is instead a hash of the `SwapFactorySwap` structure)
// and contract swap structure for the given swap ID.
func (db *RecoveryDB) PutContractSwapInfo(id types.Hash, swapID [32]byte, swap contracts.SwapFactorySwap) error {
	val, err := json.Marshal(&struct {
		SwapID [32]byte
		Swap   contracts.SwapFactorySwap
	}{
		SwapID: swapID,
		Swap:   swap,
	})
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, contractSwapInfoPrefix)
	return db.db.Put(key[:], val)
}

// PutSwapPrivateKey stores the given ephemeral swap private key share for the given swap ID.
func (db *RecoveryDB) PutSwapPrivateKey(id types.Hash, k *mcrypto.PrivateKeyInfo) error {
	val, err := json.Marshal(k)
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, swapPrivateKeyPrefix)
	return db.db.Put(key[:], val)
}

// PutSharedSwapPrivateKey stores the shared swap private key for the given swap ID.
func (db *RecoveryDB) PutSharedSwapPrivateKey(id types.Hash, k *mcrypto.PrivateKeyInfo) error {
	val, err := json.Marshal(k)
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, sharedSwapPrivateKeyPrefix)
	return db.db.Put(key[:], val)
}
