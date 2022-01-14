package swap

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/crypto/secp256k1"
	"github.com/noot/atomic-swap/dleq"
	mcrypto "github.com/noot/atomic-swap/monero/crypto"
)

var defaultTimeoutDuration = big.NewInt(60) // 60 seconds

func setupAliceAuth(t *testing.T) (*bind.TransactOpts, *ethclient.Client, *ecdsa.PrivateKey) {
	conn, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)
	pkA, err := crypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(pkA, big.NewInt(common.GanacheChainID))
	require.NoError(t, err)
	return auth, conn, pkA
}

func TestDeploySwap(t *testing.T) {
	auth, conn, _ := setupAliceAuth(t)
	address, tx, swapContract, err := DeploySwap(auth, conn, [32]byte{}, [32]byte{},
		ethcommon.Address{}, defaultTimeoutDuration)
	require.NoError(t, err)
	require.NotEqual(t, ethcommon.Address{}, address)
	require.NotNil(t, tx)
	require.NotNil(t, swapContract)
}

func TestSwap_Claim_vec(t *testing.T) {
	secret, err := hex.DecodeString("D30519BCAE8D180DBFCC94FE0B8383DC310185B0BE97B4365083EBCECCD75759")
	require.NoError(t, err)
	pubX, err := hex.DecodeString("3AF1E1EFA4D1E1AD5CB9E3967E98E901DAFCD37C44CF0BFB6C216997F5EE51DF")
	require.NoError(t, err)
	pubY, err := hex.DecodeString("E4ACAC3E6F139E0C7DB2BD736824F51392BDA176965A1C59EB9C3C5FF9E85D7A")
	require.NoError(t, err)

	var s, x, y [32]byte
	copy(s[:], secret)
	copy(x[:], pubX)
	copy(y[:], pubY)

	pk := secp256k1.NewPublicKey(x, y)
	cmt := pk.Keccak256()

	// deploy swap contract with claim key hash
	auth, conn, pkA := setupAliceAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)
	t.Logf("commitment: 0x%x", cmt)

	_, deployTx, swap, err := DeploySwap(auth, conn, cmt, [32]byte{}, addr,
		defaultTimeoutDuration)
	require.NoError(t, err)
	t.Logf("gas cost to deploy Swap.sol: %d", deployTx.Gas())

	// set contract to Ready
	_, err = swap.SetReady(auth)
	require.NoError(t, err)

	// now let's try to claim
	tx, err := swap.Claim(auth, s)
	require.NoError(t, err)
	t.Log(tx.Hash())
}

func TestSwap_Claim_random(t *testing.T) {
	// generate claim secret and public key
	dleq := &dleq.FarcasterDLEq{}
	proof, err := dleq.Prove()
	require.NoError(t, err)
	res, err := dleq.Verify(proof)
	require.NoError(t, err)

	// hash public key
	cmt := res.Secp256k1PublicKey().Keccak256()

	// deploy swap contract with claim key hash
	auth, conn, pkA := setupAliceAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	_, deployTx, swap, err := DeploySwap(auth, conn, cmt, [32]byte{}, addr,
		defaultTimeoutDuration)
	require.NoError(t, err)
	t.Logf("gas cost to deploy Swap.sol: %d", deployTx.Gas())

	// set contract to Ready
	tx, err := swap.SetReady(auth)
	require.NoError(t, err)
	t.Logf("gas cost to call SetReady: %d", tx.Gas())

	// now let's try to claim
	var s [32]byte
	secret := proof.Secret()
	copy(s[:], common.Reverse(secret[:]))
	tx, err = swap.Claim(auth, s)
	require.NoError(t, err)
	t.Logf("gas cost to call Claim: %d", tx.Gas())
}

func TestSwap_Refund_beforeT0(t *testing.T) {

}

func TestSwap_Refund_Within_T0(t *testing.T) {
	// Alice generates key
	keyPairAlice, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	pubKeyAlice := keyPairAlice.PublicKeyPair().SpendKey().Bytes()

	// Bob generates key
	keyPairBob, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	pubKeyBob := keyPairBob.PublicKeyPair().SpendKey().Bytes()

	secretAlice := keyPairAlice.SpendKeyBytes()

	var pkAliceFixed, pkBobFixed [32]byte
	copy(pkAliceFixed[:], common.Reverse(pubKeyAlice))
	copy(pkBobFixed[:], common.Reverse(pubKeyBob))

	// setup
	conn, err := ethclient.Dial("ws://127.0.0.1:8545")
	require.NoError(t, err)

	pkA, err := crypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)
	pkB, err := crypto.HexToECDSA(common.DefaultPrivKeyBob)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pkA, big.NewInt(1337)) // ganache chainID
	require.NoError(t, err)
	authAlice.Value = big.NewInt(10)

	aliceBalanceBefore, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	require.NoError(t, err)
	fmt.Println("AliceBalanceBefore: ", aliceBalanceBefore)

	bobPub := pkB.Public().(*ecdsa.PublicKey)
	bobAddr := crypto.PubkeyToAddress(*bobPub)
	contractAddress, _, swap, err := DeploySwap(authAlice, conn, pkBobFixed, pkAliceFixed, bobAddr, defaultTimeoutDuration)
	require.NoError(t, err)

	txOpts := &bind.TransactOpts{
		From:   authAlice.From,
		Signer: authAlice.Signer,
	}

	// Alice never calls set_ready on the contract, instead she just tries to Refund immediately
	var sa [32]byte
	copy(sa[:], common.Reverse(secretAlice))
	_, err = swap.Refund(txOpts, sa)
	require.NoError(t, err)

	// The Swap contract has self destructed: should have no balance AND no bytecode at the address
	contractBalance, err := conn.BalanceAt(context.Background(), contractAddress, nil)
	require.NoError(t, err)
	require.Equal(t, contractBalance.Uint64(), big.NewInt(0).Uint64())

	// bytecode, err := conn.CodeAt(context.Background(), contractAddress, nil) // nil is latest block
	// require.NoError(t, err)
	// require.Empty(t, bytecode)
}

func TestSwap_Refund_After_T1(t *testing.T) {
	// Alice generates key
	keyPairAlice, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	pubKeyAlice := keyPairAlice.PublicKeyPair().SpendKey().Bytes()

	// Bob generates key
	keyPairBob, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	pubKeyBob := keyPairBob.PublicKeyPair().SpendKey().Bytes()

	secretAlice := keyPairAlice.SpendKeyBytes()

	var pkAliceFixed, pkBobFixed [32]byte
	copy(pkAliceFixed[:], common.Reverse(pubKeyAlice))
	copy(pkBobFixed[:], common.Reverse(pubKeyBob))

	// setup
	conn, err := ethclient.Dial("ws://127.0.0.1:8545")
	require.NoError(t, err)

	pkA, err := crypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)
	pkB, err := crypto.HexToECDSA(common.DefaultPrivKeyBob)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pkA, big.NewInt(1337)) // ganache chainID
	authAlice.Value = big.NewInt(10)
	require.NoError(t, err)

	aliceBalanceBefore, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	require.NoError(t, err)
	fmt.Println("AliceBalanceBefore: ", aliceBalanceBefore)

	bobPub := pkB.Public().(*ecdsa.PublicKey)
	bobAddr := crypto.PubkeyToAddress(*bobPub)
	contractAddress, _, swap, err := DeploySwap(authAlice, conn, pkBobFixed, pkAliceFixed, bobAddr, defaultTimeoutDuration)
	require.NoError(t, err)

	txOpts := &bind.TransactOpts{
		From:   authAlice.From,
		Signer: authAlice.Signer,
	}

	// Alice calls set_ready on the contract, and immediately tries to Refund
	// After waiting T1, Alice should be able to refund now
	_, err = swap.SetReady(txOpts)
	require.NoError(t, err)

	var sa [32]byte
	copy(sa[:], common.Reverse(secretAlice))
	_, err = swap.Refund(txOpts, sa)
	require.Regexp(t, ".*It's Bob's turn now, please wait!", err)

	// wait some, then try again
	var result int64
	rpcClient, err := rpc.Dial("http://127.0.0.1:8545")
	require.NoError(t, err)

	ret := rpcClient.Call(&result, "evm_increaseTime", 3600*25)
	require.NoError(t, ret)
	_, err = swap.Refund(txOpts, sa)
	require.NoError(t, err)

	// The Swap contract has self destructed: should have no balance AND no bytecode at the address
	contractBalance, err := conn.BalanceAt(context.Background(), contractAddress, nil)
	require.NoError(t, err)
	require.Equal(t, contractBalance.Uint64(), big.NewInt(0).Uint64())

	// bytecode, err := conn.CodeAt(context.Background(), contractAddress, nil) // nil is latest block
	// require.NoError(t, err)
	// require.Empty(t, bytecode)
}
