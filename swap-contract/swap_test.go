package swap

import (
	"crypto/ecdsa"
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

	pk_a, err := crypto.HexToECDSA(keyAlice)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(1337)) // ganache chainID
	require.NoError(t, err)

	address, tx, swapContract, err := DeploySwap(authAlice, conn, [32]byte{}, [32]byte{})
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

func TestSwap_Claim(t *testing.T) {
	// Alice generates key
	keyPairAlice, err := crypto.GenerateKey()
	require.NoError(t, err)
	pubBytesAlice := encodePublicKey(keyPairAlice.Public().(*ecdsa.PublicKey))
	pubhashAlice := crypto.Keccak256Hash(pubBytesAlice[:])

	// Bob generates key
	keyPairBob, err := crypto.GenerateKey()
	require.NoError(t, err)
	pubBytesBob := encodePublicKey(keyPairBob.Public().(*ecdsa.PublicKey))
	pubhashBob := crypto.Keccak256Hash(pubBytesBob[:])

	secretBob := keyPairBob.D.Bytes()

	// setup
	conn, err := ethclient.Dial("http://127.0.0.1:8545")
	require.NoError(t, err)

	pk_a, err := crypto.HexToECDSA(keyAlice)
	require.NoError(t, err)
	pk_b, err := crypto.HexToECDSA(keyBob)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(1337)) // ganache chainID
	require.NoError(t, err)
	authBob, err := bind.NewKeyedTransactorWithChainID(pk_b, big.NewInt(1337)) // ganache chainID
	require.NoError(t, err)

	_, _, swap, err := DeploySwap(authAlice, conn, pubhashAlice, pubhashBob)
	require.NoError(t, err)

	txOpts := &bind.TransactOpts{
		From:   authAlice.From,
		Signer: authAlice.Signer,
	}

	// callOpts := &bind.CallOpts{From: address}

	// Alice calls set_ready on the contract
	_, err = swap.SetReady(txOpts)
	require.NoError(t, err)

	txOptsBob := &bind.TransactOpts{
		From:   authBob.From,
		Signer: authBob.Signer,
	}

	// Bob tries to claim
	_, err = swap.Claim(txOptsBob, setBigIntLE(secretBob))
	require.NoError(t, err)

	// TODO check whether Bob's account balance has increased
}
