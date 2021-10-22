package swap

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

const (
	keyAlice = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"
	keyBob   = "6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1"
)

func setBigIntLE(s []byte) *big.Int {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return big.NewInt(0).SetBytes(s)
}

func TestDeploySwap(t *testing.T) {
	conn, err := ethclient.Dial("http://127.0.0.1:8545")
	require.NoError(t, err)

	pk, err := crypto.HexToECDSA(keyAlice)
	require.NoError(t, err)

	auth, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(1337)) // ganache chainID
	require.NoError(t, err)

	address, tx, swapContract, err := DeploySwap(auth, conn, [32]byte{}, [32]byte{}, [32]byte{})
	require.NoError(t, err)

	t.Log(address)
	t.Log(tx)
	t.Log(swapContract)
}

func encodePublicKey(pub *ecdsa.PublicKey) [64]byte {
	px := pub.X.Bytes()
	py := pub.Y.Bytes()
	var p [64]byte
	copy(p[0:32], px)
	copy(p[32:64], py)
	return p
}

func TestSwap_Redeem(t *testing.T) {
	// Bob generates key
	kp, err := crypto.GenerateKey()
	require.NoError(t, err)

	// Bob's encoded pubkey
	pubBytes := encodePublicKey(kp.Public().(*ecdsa.PublicKey))
	pubhash := crypto.Keccak256Hash(pubBytes[:])

	// Bob's secret key, to be revealed with `Redeem()`
	kb := kp.D.Bytes()
	var sk [32]byte
	copy(sk[:], kb)

	// Alice's refund secret
	var sr [32]byte
	_, err = rand.Read(sr[:])
	require.NoError(t, err)

	// setup
	conn, err := ethclient.Dial("http://127.0.0.1:8545")
	require.NoError(t, err)

	pk, err := crypto.HexToECDSA(keyAlice)
	require.NoError(t, err)
	//alicePub := pk.Public().(*ecdsa.PublicKey)
	//address := crypto.PubkeyToAddress(*alicePub)

	auth, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(1337)) // ganache chainID
	require.NoError(t, err)

	_, _, swap, err := DeploySwap(auth, conn, sk, pubhash, sr)
	require.NoError(t, err)

	txOpts := &bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
	}

	// callOpts := &bind.CallOpts{From: address}

	_, err = swap.Ready(txOpts)
	require.NoError(t, err)

	_, err = swap.Redeem(txOpts, setBigIntLE(kb))
	require.NoError(t, err)

}
