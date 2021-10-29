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

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
)

func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func setBigIntLE(s []byte) *big.Int { //nolint
	s = reverse(s)
	return big.NewInt(0).SetBytes(s)
}

func TestDeploySwap(t *testing.T) {
	conn, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	pk_a, err := crypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(common.GanacheChainID))
	require.NoError(t, err)

	address, tx, swapContract, err := DeploySwap(authAlice, conn, [32]byte{}, [32]byte{})
	require.NoError(t, err)

	t.Log(address)
	t.Log(tx)
	t.Log(swapContract)
}

func TestSwap_Claim(t *testing.T) {
	// Alice generates key
	keyPairAlice, err := monero.GenerateKeys()
	require.NoError(t, err)
	pubKeyAlice := keyPairAlice.PublicKeyPair().SpendKey().Bytes()

	// Bob generates key
	keyPairBob, err := monero.GenerateKeys()
	require.NoError(t, err)
	pubKeyBob := keyPairBob.PublicKeyPair().SpendKey().Bytes()

	secretBob := keyPairBob.SpendKeyBytes()

	// setup
	conn, err := ethclient.Dial("ws://127.0.0.1:8545")
	require.NoError(t, err)

	pk_a, err := crypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)
	pk_b, err := crypto.HexToECDSA(common.DefaultPrivKeyBob)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(common.GanacheChainID))
	authAlice.Value = big.NewInt(10)
	require.NoError(t, err)
	authBob, err := bind.NewKeyedTransactorWithChainID(pk_b, big.NewInt(common.GanacheChainID))
	require.NoError(t, err)

	aliceBalanceBefore, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	require.NoError(t, err)
	fmt.Println("AliceBalanceBefore: ", aliceBalanceBefore)

	// check whether Bob had nothing before the Tx
	bobBalanceBefore, err := conn.BalanceAt(context.Background(), authBob.From, nil)
	require.NoError(t, err)
	fmt.Println("BobBalanceBefore: ", bobBalanceBefore)

	var pkAliceFixed [32]byte
	copy(pkAliceFixed[:], reverse(pubKeyAlice))
	var pkBobFixed [32]byte
	copy(pkBobFixed[:], reverse(pubKeyBob))
	contractAddress, deployTx, swap, err := DeploySwap(authAlice, conn, pkBobFixed, pkAliceFixed)
	require.NoError(t, err)
	fmt.Println("Deploy Tx Gas Cost:", deployTx.Gas())

	aliceBalanceAfter, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	require.NoError(t, err)
	fmt.Println("AliceBalanceAfter: ", aliceBalanceAfter)

	contractBalance, err := conn.BalanceAt(context.Background(), contractAddress, nil)
	require.NoError(t, err)
	require.Equal(t, contractBalance, big.NewInt(10))

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

	pk_a, err := crypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(1337)) // ganache chainID
	require.NoError(t, err)
	authAlice.Value = big.NewInt(10)

	aliceBalanceBefore, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	require.NoError(t, err)
	fmt.Println("AliceBalanceBefore: ", aliceBalanceBefore)

	var pkAliceFixed [32]byte
	copy(pkAliceFixed[:], reverse(pubKeyAlice))
	var pkBobFixed [32]byte
	copy(pkBobFixed[:], reverse(pubKeyBob))
	contractAddress, _, swap, err := DeploySwap(authAlice, conn, pkBobFixed, pkAliceFixed)
	require.NoError(t, err)

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

	pk_a, err := crypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(1337)) // ganache chainID
	authAlice.Value = big.NewInt(10)
	require.NoError(t, err)

	aliceBalanceBefore, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	require.NoError(t, err)
	fmt.Println("AliceBalanceBefore: ", aliceBalanceBefore)

	var pkAliceFixed [32]byte
	copy(pkAliceFixed[:], reverse(pubKeyAlice))
	var pkBobFixed [32]byte
	copy(pkBobFixed[:], reverse(pubKeyBob))
	contractAddress, _, swap, err := DeploySwap(authAlice, conn, pkBobFixed, pkAliceFixed)
	require.NoError(t, err)

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
	require.NoError(t, err)

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
