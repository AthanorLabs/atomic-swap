package mcrypto

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/crypto"
)

const (
	addressPrefixMainnet  byte = 18
	addressPrefixStagenet byte = 24
)

var (
	errChecksumMismatch         = errors.New("invalid address checksum")
	errInvalidAddressLength     = errors.New("invalid monero address length")
	errInvalidAddressEncoding   = errors.New("invalid monero address encoding")
	errInvalidPrefixGotMainnet  = errors.New("invalid monero address: expected stagenet, got mainnet")
	errInvalidPrefixGotStagenet = errors.New("invalid monero address: expected mainnet, got stagenet")
)

// Address represents a base58-encoded string
type Address string

// ValidateAddress checks if the given address is valid
func ValidateAddress(addr string, env common.Environment) error {
	b, err := MoneroAddrBase58ToBytes(addr)
	if err != nil {
		return err
	}

	switch env {
	case common.Mainnet, common.Development:
		if b[0] != addressPrefixMainnet {
			return errInvalidPrefixGotStagenet
		}
	case common.Stagenet:
		if b[0] != addressPrefixStagenet {
			return errInvalidPrefixGotMainnet
		}
	}

	checksum := getChecksum(b[:65])
	if !bytes.Equal(checksum[:], b[65:69]) {
		return errChecksumMismatch
	}

	return nil
}

func getChecksum(data ...[]byte) (result [4]byte) {
	keccak256 := crypto.Keccak256(data...)
	copy(result[:], keccak256[:4])
	return
}

// AddressBytes returns the address as bytes for a PrivateKeyPair with the given environment (ie. mainnet or stagenet)
func (kp *PrivateKeyPair) AddressBytes(env common.Environment) []byte {
	return kp.PublicKeyPair().AddressBytes(env)
}

// Address returns the base58-encoded address for a PrivateKeyPair with the given environment
// (ie. mainnet or stagenet)
func (kp *PrivateKeyPair) Address(env common.Environment) Address {
	return Address(MoneroAddrBytesToBase58(kp.AddressBytes(env)))
}

// AddressBytes returns the address as bytes for a PublicKeyPair with the given environment (ie. mainnet or stagenet)
func (kp *PublicKeyPair) AddressBytes(env common.Environment) []byte {
	psk := kp.sk.key.Bytes()
	pvk := kp.vk.key.Bytes()

	var prefix byte
	switch env {
	case common.Mainnet, common.Development:
		prefix = addressPrefixMainnet
	case common.Stagenet:
		prefix = addressPrefixStagenet
	default:
		panic(fmt.Sprintf("unhandled env %d", env))
	}

	// address encoding is:
	// (network_prefix) + (32-byte public spend key) + (32-byte-byte public view key)
	// + first_4_Bytes(Hash(network_prefix + (32-byte public spend key) + (32-byte public view key)))
	addr := append(append([]byte{prefix}, psk...), pvk...)
	checksum := getChecksum(addr)
	addrWithChecksum := append(addr, checksum[:4]...)
	if len(addrWithChecksum) != 69 { // 1 (prefix) + 32 (pub spend key) + 32 (pub view key) + 4 (checksum)
		panic(fmt.Sprintf("monero address %d instead of 69", len(addrWithChecksum)))
	}
	return addrWithChecksum
}

// Address returns the base58-encoded address for a PublicKeyPair with the given environment
// (ie. mainnet or stagenet)
func (kp *PublicKeyPair) Address(env common.Environment) Address {
	return Address(MoneroAddrBytesToBase58(kp.AddressBytes(env)))
}
