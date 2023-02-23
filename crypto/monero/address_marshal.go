package mcrypto

// MarshalText serializes the Monero Address type with some extra validation.
func (a *Address) MarshalText() ([]byte, error) {
	if err := a.validateDecoded(); err != nil {
		return nil, err
	}
	return []byte(moneroAddrBytesToBase58(a.decoded[:])), nil
}

// UnmarshalText validates that the string represents a properly formatted
// monero address. The encoding, length and checksum are all validated, but not
// the network, as it is unknown by the JSON parser. Empty strings are not
// allowed. Use an address pointer in your serialized types if the Address is
// optional.
func (a *Address) UnmarshalText(base58Input []byte) error {
	base58Str := string(base58Input)
	addrBytes, err := moneroAddrBase58ToBytes(base58Str)
	if err != nil {
		return err
	}

	newAddr := new(Address)
	n := copy(newAddr.decoded[:], addrBytes)
	if n != addressBytesLen {
		// moneroAddrBase58ToBytes already verified the decoded length
		panic("bytes to address conversion is broken")
	}

	if err := newAddr.validateDecoded(); err != nil {
		return err
	}

	// No more errors possible, overwrite the existing value
	*a = *newAddr
	return nil
}
