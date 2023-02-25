package mcrypto

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
)

const (
	// addressBytesLen is the length (69) of a Monero address in raw bytes:
	//  1 - Network byte
	// 32 - Public spend key
	// 32 - Public view key
	//  4 - First 4 bytes of keccak-256 checksum of previous bytes
	addressBytesLen = 1 + 32 + 32 + 4

	// encodedAddressLen is the length (95) of a base58 encoded Monero address:
	// 88 - Eight, 11-symbol base58 blocks each representing 8 binary bytes (64 binary bytes total)
	//  7 - Remaining base58 block representing 5 binary bytes
	encodedAddressLen = 8*11 + 1*7

	// encodedIntegratedAddrLen is only for giving better error messages. We don't support
	// integrated addresses. In the byte form, they have an additional 8-byte payment ID
	// between the public view key and the checksum. The additional 8 bytes converts to
	// an additional 11 bytes in base58.
	encodedIntegratedAddrLen = encodedAddressLen + 11
)

// addrBytesToBase58 takes a 69-byte binary monero address (including the 4-byte
// checksum) and returns it encoded using Monero's unique base58 algorithm. It is the
// caller's responsibility to only pass 65 byte input slices.
func addrBytesToBase58(addrBytes []byte) string {
	if len(addrBytes) != addressBytesLen {
		panic("addrBytesToBase58 passed non-addrBytes value")
	}

	var encodedAddr string

	// Handle the first 64 binary bytes in 8 byte chunks yielding exactly 88 (8 * 11)
	// base58 characters.
	for i := 0; i < 8; i++ {
		// Each encoded block will be 11 characters or fewer. If less, we pad to 11.
		block := base58.Encode(addrBytes[i*8 : i*8+8]) // yields 11 or fewer characters
		if len(block) < 11 {
			// Prepend "1"'s (zero in base58) as padding to get exactly 11 characters.
			block = strings.Repeat("1", 11-len(block)) + block
		}
		encodedAddr += block
	}
	// Last block is 5 bytes which converts to 7 characters or fewer in base58. We always
	// pad to 7 characters giving an encoded address size of 95 characters.
	//
	// Note: If you wanted to write a general purpose, monero-specific, base58 encoder,
	// you'd keep a table of modulus-8 values mapped to their maximum base58 encoded
	// length like this: https://github.com/monero-rs/base58-monero/blob/v1.0.0/src/base58.rs#L92-L93
	// It's not functionality that we would use, so all we need to know is that 5 binary
	// bytes maps to 7 or fewer base58 characters.
	lastBlock := base58.Encode(addrBytes[64:])
	if len(lastBlock) < 7 {
		// Prepend "1"'s (zero in base58) as padding to get exactly 7 characters.
		lastBlock = strings.Repeat("1", 7-len(lastBlock)) + lastBlock
	}
	encodedAddr += lastBlock

	return encodedAddr
}

// addrBase58ToBytes decodes a monero base58 encoded address into a byte slice.
// Only decoding is done here, the checksum should be verified after this decoding.
func addrBase58ToBytes(encodedAddress string) ([]byte, error) {
	if len(encodedAddress) != encodedAddressLen {
		err := errInvalidAddressLength
		if len(encodedAddress) == encodedIntegratedAddrLen {
			err = fmt.Errorf("integrated addresses not supported: %w", err)
		}
		return nil, err
	}

	result := make([]byte, 0, addressBytesLen)

	// Handle the first 88 bytes in 11-byte base58 chunks. Each 11 byte chunk converts to
	// 8 binary bytes.
	for i := 0; i < 8; i++ {
		block := base58.Decode(encodedAddress[i*11 : i*11+11])
		if len(block) == 0 {
			return nil, errInvalidAddressEncoding
		}
		// The decoder will never return less than 8 bytes from 11 base58 input
		// characters, but it can return up to 11 bytes, because it adds a leading zero
		// byte for every sequential "1" symbol on the left of the input. So in the edge
		// case of passing 11 1's "11111111111", you'll get back 11 bytes of zeros.
		block = block[len(block)-8:] // strip any leading zeros
		result = append(result, block...)
	}
	// Handle the final 7 bytes, which convert to 5 binary bytes
	lastBlock := base58.Decode(encodedAddress[88:])
	if len(lastBlock) == 0 {
		return nil, errInvalidAddressEncoding
	}
	// See above. We can decode up to 7 bytes with leading zeros, but never less than 5.
	lastBlock = lastBlock[len(lastBlock)-5:] // strip any leading zeros
	result = append(result, lastBlock...)

	if len(result) != addressBytesLen {
		panic("base58 address decoder is broken")
	}

	return result, nil
}
