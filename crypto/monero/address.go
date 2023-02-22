package mcrypto

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/crypto"
)

// Network is the Monero network type
type Network string

// Monero networks
const (
	Mainnet  Network = "mainnet"
	Stagenet Network = "stagenet"
	Testnet  Network = "testnet"
)

// AddressType is the type of Monero address: Standard, Integrated, Subaddress
type AddressType string

// Monero address types
const (
	Standard   AddressType = "standard"
	Integrated AddressType = "integrated"
	Subaddress AddressType = "subaddress"
)

// Network prefix byte. The 1st decoded byte of a monero address defines both
// the network (mainnet, stagenet, testnet) and the type of address (standard,
// integrated, and subaddress).
const (
	netPrefixStdAddrMainnet     = 18
	netPrefixIntegrAddrMainnet  = 19
	netPrefixSubAddrMainnet     = 42
	netPrefixStdAddrStagenet    = 24
	netPrefixIntegrAddrStagenet = 25
	netPrefixSubAddrStagenet    = 36
	netPrefixStdAddrTestnet     = 53
	netPrefixIntegrAddrTestnet  = 54
	netPrefixSubAddrTestnet     = 63
)

var (
	errChecksumMismatch         = errors.New("invalid address checksum")
	errInvalidAddressLength     = errors.New("invalid monero address length")
	errInvalidAddressEncoding   = errors.New("invalid monero address encoding")
	errInvalidPrefixGotMainnet  = errors.New("invalid monero address: expected stagenet, got mainnet")
	errInvalidPrefixGotStagenet = errors.New("invalid monero address: expected mainnet, got stagenet")
	errInvalidPrefixGotTestnet  = errors.New("invalid monero address: monero testnet not yet supported")
)

// Address represents a base58-encoded string
type Address struct {
	decoded [AddressBytesLen]byte
}

// NewAddress converts a string to a monero Address with validation.
func NewAddress(addrStr string, env common.Environment) (*Address, error) {
	addr := &Address{}
	if err := addr.UnmarshalText([]byte(addrStr)); err != nil {
		return nil, err
	}

	return addr, addr.Validate(env)
}

func (a *Address) String() string {
	return moneroAddrBytesToBase58(a.decoded[:])
}

// Network returns the Monero network of the address
func (a *Address) Network() Network {
	switch a.decoded[0] {
	case netPrefixStdAddrMainnet,
		netPrefixIntegrAddrMainnet,
		netPrefixSubAddrMainnet:
		return Mainnet
	case netPrefixStdAddrStagenet,
		netPrefixIntegrAddrStagenet,
		netPrefixSubAddrStagenet:
		return Stagenet
	case netPrefixStdAddrTestnet,
		netPrefixIntegrAddrTestnet,
		netPrefixSubAddrTestnet:
		return Testnet
	default:
		// Our methods to deserialize and create Address values all verify
		// that the address byte is valid
		panic("address has invalid network prefix")
	}
}

// Type returns the Address type
func (a *Address) Type() AddressType {
	switch a.decoded[0] {
	case netPrefixStdAddrMainnet,
		netPrefixStdAddrStagenet,
		netPrefixStdAddrTestnet:
		return Standard
	case netPrefixIntegrAddrMainnet,
		netPrefixIntegrAddrStagenet,
		netPrefixIntegrAddrTestnet:
		return Integrated
	case netPrefixSubAddrTestnet,
		netPrefixSubAddrStagenet,
		netPrefixSubAddrMainnet:
		return Subaddress
	default:
		// Our methods to deserialize and create Address values all verify
		// that the address byte is valid
		panic("address has invalid network prefix")
	}
}

// Validate validates that the monero network matches the passed environment.
func (a *Address) Validate(env common.Environment) error {
	moneroNet := a.Network()
	switch moneroNet {
	case Mainnet:
		if env != common.Mainnet && env != common.Development {
			return errInvalidPrefixGotMainnet
		}
	case Stagenet:
		if env != common.Stagenet {
			return errInvalidPrefixGotStagenet
		}
	case Testnet:
		return errInvalidPrefixGotTestnet
	}

	checksum := getChecksum(a.decoded[:65])
	if !bytes.Equal(checksum[:], a.decoded[65:69]) {
		return errChecksumMismatch
	}

	return nil
}

func getChecksum(data ...[]byte) (result [4]byte) {
	keccak256 := crypto.Keccak256(data...)
	copy(result[:], keccak256[:4])
	return
}

// Address returns the address as bytes for a PublicKeyPair with the given environment (ie. mainnet or stagenet)
func (kp *PublicKeyPair) Address(env common.Environment) *Address {
	address := new(Address)

	var prefix byte
	switch env {
	case common.Mainnet, common.Development:
		prefix = netPrefixStdAddrMainnet
	case common.Stagenet:
		prefix = netPrefixStdAddrStagenet
	default:
		panic(fmt.Sprintf("unhandled env %d", env))
	}

	// address encoding is:
	// (network_prefix) + (32-byte public spend key) + (32-byte-byte public view key)
	// + first_4_Bytes(Hash(network_prefix + (32-byte public spend key) + (32-byte public view key)))
	address.decoded[0] = prefix                 // 1-byte network prefix
	copy(address.decoded[1:33], kp.sk.Bytes())  // 32-byte public spend key
	copy(address.decoded[33:65], kp.vk.Bytes()) // 32-byte public view key
	checksum := getChecksum(address.decoded[0:65])
	copy(address.decoded[65:69], checksum[:])

	return address
}
