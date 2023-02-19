package db

import (
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/common/vjson"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"

	"github.com/ChainSafe/chaindb"
)

const (
	recoveryPrefix                   = "recv"
	contractSwapInfoPrefix           = "ethinfo"
	swapPrivateKeyPrefix             = "privkey"
	counterpartySwapPrivateKeyPrefix = "cspriv"
	relayerInfoPrefix                = "relayer"
	xmrmakerKeysPrefix               = "xmrmaker"
)

// RecoveryDB contains information about ongoing swaps required for recovery
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
	val, err := vjson.MarshalStruct(info)
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
	err = vjson.UnmarshalStruct(value, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// PutContractSwapInfo stores the given contract swap ID (which is not the same as the daemon
// swap ID, but is instead a hash of the `SwapFactorySwap` structure)
// and contract swap structure for the given swap ID.
func (db *RecoveryDB) PutContractSwapInfo(id types.Hash, info *EthereumSwapInfo) error {
	val, err := vjson.MarshalStruct(info)
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
	err = vjson.UnmarshalStruct(value, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// PutSwapPrivateKey stores the given ephemeral swap private key share for the given swap ID.
func (db *RecoveryDB) PutSwapPrivateKey(id types.Hash, sk *mcrypto.PrivateSpendKey) error {
	val, err := vjson.MarshalStruct(sk)
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, swapPrivateKeyPrefix)
	return db.db.Put(key[:], val)
}

// GetSwapPrivateKey returns the swap private key share, if it exists.
func (db *RecoveryDB) GetSwapPrivateKey(id types.Hash) (*mcrypto.PrivateSpendKey, error) {
	key := getRecoveryDBKey(id, swapPrivateKeyPrefix)
	value, err := db.db.Get(key[:])
	if err != nil {
		return nil, err
	}

	privSpendKey := new(mcrypto.PrivateSpendKey)
	err = vjson.UnmarshalStruct(value, privSpendKey)
	if err != nil {
		return nil, err
	}

	return privSpendKey, nil
}

// PutCounterpartySwapPrivateKey stores the counterparty's swap private key for the given swap ID.
func (db *RecoveryDB) PutCounterpartySwapPrivateKey(id types.Hash, kp *mcrypto.PrivateSpendKey) error {
	val, err := vjson.MarshalStruct(kp)
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, counterpartySwapPrivateKeyPrefix)
	return db.db.Put(key[:], val)
}

// GetCounterpartySwapPrivateKey returns the counterparty's swap private key, if it exists.
func (db *RecoveryDB) GetCounterpartySwapPrivateKey(id types.Hash) (*mcrypto.PrivateSpendKey, error) {
	key := getRecoveryDBKey(id, counterpartySwapPrivateKeyPrefix)
	value, err := db.db.Get(key[:])
	if err != nil {
		return nil, err
	}

	sk := new(mcrypto.PrivateSpendKey)
	err = vjson.UnmarshalStruct(value, sk)
	if err != nil {
		return nil, err
	}

	return sk, nil
}

type xmrmakerKeys struct {
	PublicSpendKey *mcrypto.PublicKey      `json:"publicSpendKey" validate:"required"`
	PrivateViewKey *mcrypto.PrivateViewKey `json:"privateViewKey" validate:"required"`
}

// PutXMRMakerSwapKeys is called by the xmrtaker to store the counterparty's swap keys.
func (db *RecoveryDB) PutXMRMakerSwapKeys(id types.Hash, sk *mcrypto.PublicKey, vk *mcrypto.PrivateViewKey) error {
	val, err := vjson.MarshalStruct(&xmrmakerKeys{
		PublicSpendKey: sk,
		PrivateViewKey: vk,
	})
	if err != nil {
		return err
	}

	key := getRecoveryDBKey(id, xmrmakerKeysPrefix)
	return db.db.Put(key[:], val)
}

// GetXMRMakerSwapKeys is called by the xmrtaker during recovery to retrieve the counterparty's
// swap keys.
func (db *RecoveryDB) GetXMRMakerSwapKeys(id types.Hash) (*mcrypto.PublicKey, *mcrypto.PrivateViewKey, error) {
	key := getRecoveryDBKey(id, xmrmakerKeysPrefix)
	value, err := db.db.Get(key)
	if err != nil {
		return nil, nil, err
	}

	var info xmrmakerKeys
	err = vjson.UnmarshalStruct(value, &info)
	if err != nil {
		return nil, nil, err
	}

	return info.PublicSpendKey, info.PrivateViewKey, nil
}

// DeleteSwap deletes all recovery info from the db for the given swap.
func (db *RecoveryDB) DeleteSwap(id types.Hash) error {
	keys := [][]byte{
		getRecoveryDBKey(id, relayerInfoPrefix),
		getRecoveryDBKey(id, contractSwapInfoPrefix),
		getRecoveryDBKey(id, swapPrivateKeyPrefix),
		getRecoveryDBKey(id, counterpartySwapPrivateKeyPrefix),
		getRecoveryDBKey(id, xmrmakerKeysPrefix),
	}

	for _, key := range keys {
		err := db.db.Del(key)
		if err != nil {
			return err
		}
	}

	return nil
}
