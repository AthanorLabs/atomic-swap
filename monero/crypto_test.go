package monero

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/stretchr/testify/require"
)

func TestPrivateKeyPairToAddress(t *testing.T) {
	skBytes := "1186c5edbfce9f003e157f35d01af91745aa075f913362b604be10992b27490c"
	pskBytes := "d7b83c2acd596568ed699644aa2092b7f75d6b9b8cd0d7a60e2c5b14f1d328e6"
	vkBytes := "e41d9be1691ae364d96f2f655fc869f0d6c51502c90baebd3fa91bf805feae04"
	pvkBytes := "f3e9adeaed9400c15386162207093b3324273fc643d2de466240e2b51b87d781"

	sk, err := hex.DecodeString(skBytes)
	require.NoError(t, err)

	psk, err := hex.DecodeString(pskBytes)
	require.NoError(t, err)

	vk, err := hex.DecodeString(vkBytes)
	require.NoError(t, err)

	pvk, err := hex.DecodeString(pvkBytes)
	require.NoError(t, err)

	// test DecodeMoneroBase58
	address := "49oFJna6jrkJYvmupQktXKXmhnktf1aCvUmwp8HJGvY7fdXpLMTVeqmZLWQLkyHXuU9Z8mZ78LordCmp3Nqx5T9GFdEGueB"
	addressBytes := DecodeMoneroBase58(address)
	require.Equal(t, psk, addressBytes[1:33])
	require.Equal(t, pvk, addressBytes[33:65])

	// test that Address() gives the correct address bytes
	// implicitly tests that the *PrivateSpendKey.Public() and *PrivateViewKey.Public()
	// give the correct public keys
	kp, err := NewPrivateKeyPairFromBytes(sk, vk)
	require.NoError(t, err)
	require.Equal(t, addressBytes, kp.AddressBytes())
	require.Equal(t, Address(address), kp.Address())

	// check public key derivation
	require.Equal(t, pskBytes, kp.sk.Public().Hex())
	require.Equal(t, pvkBytes, kp.vk.Public().Hex())
}

func TestGeneratePrivateKeyPair(t *testing.T) {
	_, err := GenerateKeys()
	require.NoError(t, err)
}

func TestKeccak256(t *testing.T) {
	res := crypto.Keccak256([]byte{1})
	res2 := Keccak256([]byte{1})
	require.Equal(t, res, res2[:])
}
