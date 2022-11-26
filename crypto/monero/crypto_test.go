package mcrypto

import (
	"encoding/hex"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/crypto"
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

	// test MoneroAddrBase58ToBytes
	address := "49oFJna6jrkJYvmupQktXKXmhnktf1aCvUmwp8HJGvY7fdXpLMTVeqmZLWQLkyHXuU9Z8mZ78LordCmp3Nqx5T9GFdEGueB"
	addressBytes, err := MoneroAddrBase58ToBytes(address)
	require.NoError(t, err)
	require.Equal(t, psk, addressBytes[1:33])
	require.Equal(t, pvk, addressBytes[33:65])

	// test that Address() gives the correct address bytes
	// implicitly tests that the *PrivateSpendKey.Public() and *PrivateViewKey.Public()
	// give the correct public keys
	kp, err := NewPrivateKeyPairFromBytes(sk, vk)
	require.NoError(t, err)
	require.Equal(t, addressBytes, kp.AddressBytes(common.Mainnet))
	require.Equal(t, Address(address), kp.Address(common.Mainnet))

	// check public key derivation
	require.Equal(t, pskBytes, kp.sk.Public().Hex())
	require.Equal(t, pvkBytes, kp.vk.Public().Hex())
}

func TestGeneratePrivateKeyPair(t *testing.T) {
	_, err := GenerateKeys()
	require.NoError(t, err)
}

func TestKeccak256(t *testing.T) {
	res := ethcrypto.Keccak256([]byte{1})
	res2 := crypto.Keccak256([]byte{1})
	require.Equal(t, res, res2[:])
}

func TestNewPrivateSpendKey(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)

	sk, err := NewPrivateSpendKey(kp.sk.Bytes())
	require.NoError(t, err)
	require.Equal(t, kp.sk.key, sk.key)
}

func TestNewPrivateViewKeyFromHex(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)

	vk, err := NewPrivateViewKeyFromHex(kp.vk.Hex())
	require.NoError(t, err)
	require.Equal(t, kp.vk.key, vk.key)
}

func TestNewPublicKeyFromHex(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)

	pk, err := NewPublicKeyFromHex(kp.sk.Public().Hex())
	require.NoError(t, err)
	require.Equal(t, kp.sk.Public().key.Bytes(), pk.key.Bytes())
}

func TestNewPublicKeyPairFromHex(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)

	kp2, err := NewPublicKeyPairFromHex(kp.sk.Public().Hex(), kp.vk.Public().Hex())
	require.NoError(t, err)
	require.Equal(t, kp.sk.Public().key.Bytes(), kp2.sk.key.Bytes())
	require.Equal(t, kp.vk.Public().key.Bytes(), kp2.vk.key.Bytes())
}

func TestPrivateSpendKey_View(t *testing.T) {
	type testData struct {
		spendKey string
		viewKey  string
	}
	// Dataset is from wallets created by monero-wallet-rpc using query_key to get the values
	tests := []testData{
		{
			spendKey: "6d0fffe9979b3091e25996293fc638999daf54a6a6cedb6ed95217ab409fcd07",
			viewKey:  "534b59efaf46832581250d0b39202c62beb258b0bff3257bbdfa45a545220100",
		},
		{
			spendKey: "ca9fd8206ab9f98cfa36d44153764e9f301844e66d21e0f4938b9a70623fd40c",
			viewKey:  "bb88581844c4995f61ccdf0e2be820a849afdb9ed4b44f875e6db31ce41d5802",
		},
		{
			spendKey: "e1b9c2b0a0933372e33817bc583895cd23ae7b5bdc3d255090b0a122828e8c07",
			viewKey:  "afc7d62084fa263a25ace85f267a3d7b3f7dcef73b8bb82f5977c0c62758d500",
		},
		{
			spendKey: "6d231208674247aeeeff247a7d27251e0fdb07719f704df36b3baa20f2695006",
			viewKey:  "78d140e53b2bac980665e6bb6d83cfab71911de1bf24119e861649b077b49d0a",
		},
		{
			spendKey: "bcaf422a13e3a93affc42ad99aadd2c99e2a7317068ada36f2bb3bf42decfd0a",
			viewKey:  "22f223704c73be57b6b968088f03ad0d760479743399864343d9fbd499a85702",
		},
		{
			spendKey: "ba4bb6fc845860a556333e80204957c163ccdf4378a53110ec76209cc742fc03",
			viewKey:  "b7a6cda35038613a34ab46f01636fdd1dc17ea35f0516f60c0ec896c2169ac0b",
		},
		{
			spendKey: "92a0fb1f1abeabed9c58bdee54b2b1089b836f721e0c530c722c289f1e505909",
			viewKey:  "5b6813c1fb9313ad18aaa7894259e343976e98e89b43e2491a1e7f72fdfede0d",
		},
		{
			spendKey: "adb6b821ab2aab5b84bc0e6dc743d8147da292cc604c4a382422e2c7d359a200",
			viewKey:  "a934f103f78dab73ef7560c8c9bf9df664087974b5cb9783b16c295900e5c80a",
		},
		{
			spendKey: "4397e2a5edb21ebdc73158a4f25ed0d5df47aebfc069755dabac6b8fa0ef7c03",
			viewKey:  "2d56357641936ba8c7a195eddb35e4878dba2134959143fbd10a35b334687903",
		},
		{
			spendKey: "ffe38d5bbdf17d56373486173eff41710b21f795f04dda17b14e3bcaeac5f204",
			viewKey:  "d31647011c5fda8170ff5917445b528be80549c98a1eb83682e988cebde55f0b",
		},
	}
	for _, tt := range tests {
		psk := &PrivateSpendKey{}
		err := psk.UnmarshalText([]byte(tt.spendKey))
		require.NoError(t, err)
		vk, err := psk.View()
		require.NoError(t, err)
		require.Equal(t, tt.viewKey, vk.Hex())
	}
}
