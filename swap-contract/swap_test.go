package swap

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"

	"github.com/noot/atomic-swap/monero"
)

const (
	keyAlice = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"
	keyBob   = "6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1"
)

func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func setBigIntLE(s []byte) *big.Int {
	s = reverse(s)
	return big.NewInt(0).SetBytes(s)
}

func TestDeploySwap(t *testing.T) {
	conn, err := ethclient.Dial("http://127.0.0.1:8545")
	require.NoError(t, err)

	pk_a, err := crypto.HexToECDSA(keyAlice)
	require.NoError(t, err)

	pk_b, err := crypto.HexToECDSA(keyBob)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(1337)) // ganache chainID
	require.NoError(t, err)

	address, tx, swapContract, err := DeploySwap(authAlice, conn, pk_a.X, pk_a.Y, pk_b.X, pk_b.Y)
	require.NoError(t, err)

	t.Log(address)
	t.Log(tx)
	t.Log(swapContract)
}

func TestSwap_Claim(t *testing.T) {
	// Alice generates key
	keyPairAlice, err := monero.GenerateKeys()
	require.NoError(t, err)
	// pubKeyAlice := keyPairAlice.PublicKeyPair().SpendKey().Bytes()
	pubKeyAliceX, pubKeyAliceY := monero.PublicSpendOnSecp256k1(keyPairAlice.SpendKey())

	// Bob generates key
	keyPairBob, err := monero.GenerateKeys()
	require.NoError(t, err)
	pubKeyBobX, pubKeyBobY := monero.PublicSpendOnSecp256k1(keyPairBob.SpendKey())
	// pubKeyBob := keyPairBob.PublicKeyPair().SpendKey().Bytes()

	secretBob := keyPairBob.SpendKeyBytes()

	// setup
	conn, err := ethclient.Dial("ws://127.0.0.1:8545")
	require.NoError(t, err)

	pk_a, err := crypto.HexToECDSA(keyAlice)
	require.NoError(t, err)
	pk_b, err := crypto.HexToECDSA(keyBob)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(1337)) // ganache chainID
	authAlice.Value = big.NewInt(10)
	require.NoError(t, err)
	authBob, err := bind.NewKeyedTransactorWithChainID(pk_b, big.NewInt(1337)) // ganache chainID
	require.NoError(t, err)

	aliceBalanceBefore, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	fmt.Println("AliceBalanceBefore: ", aliceBalanceBefore)
	// check whether Bob had nothing before the Tx
	bobBalanceBefore, err := conn.BalanceAt(context.Background(), authBob.From, nil)
	fmt.Println("BobBalanceBefore: ", bobBalanceBefore)
	require.NoError(t, err)

	var pkAliceFixed [32]byte
	copy(pkAliceFixed[:], reverse(pubKeyAlice))
	var pkBobFixed [32]byte
	copy(pkBobFixed[:], reverse(pubKeyBob))
	contractAddress, deployTx, swap, err := DeploySwap(authAlice, conn, pkBobFixed, pkAliceFixed)
	fmt.Println("Deploy Tx Gas Cost:", deployTx.Gas())
	aliceBalanceAfter, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	fmt.Println("AliceBalanceAfter: ", aliceBalanceAfter)
	contractBalance, err := conn.BalanceAt(context.Background(), contractAddress, nil)
	require.Equal(t, contractBalance, big.NewInt(10))
	require.NoError(t, err)

	txOpts := &bind.TransactOpts{
		From:   authAlice.From,
		Signer: authAlice.Signer,
	}

	txOptsBob := &bind.TransactOpts{
		From:   authBob.From,
		Signer: authBob.Signer,
	}

	// Bob tries to claim before Alice has called ready, should fail
	s := big.NewInt(0).SetBytes(reverse(secretBob))
	fmt.Println("Secret:", hex.EncodeToString(reverse(secretBob)))
	fmt.Println("PubKey:", hex.EncodeToString(reverse(pubKeyBob)))
	_, err = swap.Claim(txOptsBob, s)
	require.Regexp(t, ".*'isReady == false' cannot claim yet!", err)

	// Alice calls set_ready on the contract
	setReadyTx, err := swap.SetReady(txOpts)
	fmt.Println("setReady Tx Gas Cost:", setReadyTx.Gas())
	require.NoError(t, err)

	// The main transaction that we're testing. Should work
	tx, err := swap.Claim(txOptsBob, s)
	require.NoError(t, err)

	// The Swap contract has self destructed: should have no balance AND no bytecode at the address
	contractBalance, err = conn.BalanceAt(context.Background(), contractAddress, nil)
	require.NoError(t, err)
	require.Equal(t, contractBalance.Uint64(), big.NewInt(0).Uint64())
	bytecode, err := conn.CodeAt(context.Background(), contractAddress, nil) // nil is latest block
	require.NoError(t, err)
	require.Empty(t, bytecode)

	fmt.Println("Tx details are:", tx.Gas())

	// check whether Bob's account balance has increased now
	// bobBalanceAfter, err := conn.BalanceAt(context.Background(), authBob.From, nil)
	// fmt.Println("BobBalanceBefore: ", bobBalanceAfter)
	// require.NoError(t, err)
	// require.Equal(t, big.NewInt(10), big.NewInt(0).Sub(bobBalanceAfter, bobBalanceBefore))
}

func TestSwap_Refund_Within_T0(t *testing.T) {
	// Alice generates key
	keyPairAlice, err := monero.GenerateKeys()
	require.NoError(t, err)
	pubKeyAlice := keyPairAlice.PublicKeyPair().SpendKey().Bytes()

	// Bob generates key
	keyPairBob, err := monero.GenerateKeys()
	require.NoError(t, err)
	pubKeyBob := keyPairBob.PublicKeyPair().SpendKey().Bytes()

	secretAlice := keyPairAlice.SpendKeyBytes()

	// setup
	conn, err := ethclient.Dial("ws://127.0.0.1:8545")
	require.NoError(t, err)

	pk_a, err := crypto.HexToECDSA(keyAlice)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(1337)) // ganache chainID
	authAlice.Value = big.NewInt(10)
	require.NoError(t, err)

	aliceBalanceBefore, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	fmt.Println("AliceBalanceBefore: ", aliceBalanceBefore)

	var pkAliceFixed [32]byte
	copy(pkAliceFixed[:], reverse(pubKeyAlice))
	var pkBobFixed [32]byte
	copy(pkBobFixed[:], reverse(pubKeyBob))
	contractAddress, _, swap, err := DeploySwap(authAlice, conn, pkBobFixed, pkAliceFixed)

	txOpts := &bind.TransactOpts{
		From:   authAlice.From,
		Signer: authAlice.Signer,
	}

	// Alice never calls set_ready on the contract, instead she just tries to Refund immidiately
	s := big.NewInt(0).SetBytes(reverse(secretAlice))
	_, err = swap.Refund(txOpts, s)
	require.NoError(t, err)

	// The Swap contract has self destructed: should have no balance AND no bytecode at the address
	contractBalance, err := conn.BalanceAt(context.Background(), contractAddress, nil)
	require.NoError(t, err)
	require.Equal(t, contractBalance.Uint64(), big.NewInt(0).Uint64())

	bytecode, err := conn.CodeAt(context.Background(), contractAddress, nil) // nil is latest block
	require.NoError(t, err)
	require.Empty(t, bytecode)

}

func TestSwap_Refund_After_T1(t *testing.T) {
	// Alice generates key
	keyPairAlice, err := monero.GenerateKeys()
	require.NoError(t, err)
	pubKeyAlice := keyPairAlice.PublicKeyPair().SpendKey().Bytes()

	// Bob generates key
	keyPairBob, err := monero.GenerateKeys()
	require.NoError(t, err)
	pubKeyBob := keyPairBob.PublicKeyPair().SpendKey().Bytes()

	secretAlice := keyPairAlice.SpendKeyBytes()

	// setup
	conn, err := ethclient.Dial("ws://127.0.0.1:8545")
	require.NoError(t, err)

	pk_a, err := crypto.HexToECDSA(keyAlice)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(1337)) // ganache chainID
	authAlice.Value = big.NewInt(10)
	require.NoError(t, err)

	aliceBalanceBefore, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	fmt.Println("AliceBalanceBefore: ", aliceBalanceBefore)

	var pkAliceFixed [32]byte
	copy(pkAliceFixed[:], reverse(pubKeyAlice))
	var pkBobFixed [32]byte
	copy(pkBobFixed[:], reverse(pubKeyBob))
	contractAddress, _, swap, err := DeploySwap(authAlice, conn, pkBobFixed, pkAliceFixed)

	txOpts := &bind.TransactOpts{
		From:   authAlice.From,
		Signer: authAlice.Signer,
	}

	// Alice calls set_ready on the contract, and immediately tries to Refund
	// After waiting T1, Alice should be able to refund now
	s := big.NewInt(0).SetBytes(reverse(secretAlice))
	_, err = swap.SetReady(txOpts)
	require.NoError(t, err)

	_, err = swap.Refund(txOpts, s)
	require.Regexp(t, ".*It's Bob's turn now, please wait!", err)

	// wait some, then try again
	var result int64
	rpcClient, err := rpc.Dial("http://127.0.0.1:8545")

	ret := rpcClient.Call(&result, "evm_increaseTime", 3600*25)
	require.NoError(t, ret)
	_, err = swap.Refund(txOpts, s)
	require.NoError(t, err)

	// The Swap contract has self destructed: should have no balance AND no bytecode at the address
	contractBalance, err := conn.BalanceAt(context.Background(), contractAddress, nil)
	require.NoError(t, err)
	require.Equal(t, contractBalance.Uint64(), big.NewInt(0).Uint64())

	bytecode, err := conn.CodeAt(context.Background(), contractAddress, nil) // nil is latest block
	require.NoError(t, err)
	require.Empty(t, bytecode)
}
