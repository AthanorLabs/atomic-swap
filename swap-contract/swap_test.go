package swap

import (
	"context"
	"crypto/ecdsa"
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
	"github.com/noot/atomic-swap/monero"
)

var defaultTimeoutDuration = big.NewInt(60) // 60 seconds

func TestDeploySwap(t *testing.T) {
	conn, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	pk_a, err := crypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(common.GanacheChainID))
	require.NoError(t, err)

	address, tx, swapContract, err := DeploySwap(authAlice, conn, [32]byte{}, [32]byte{}, ethcommon.Address{}, defaultTimeoutDuration)
	require.NoError(t, err)
	require.NotEqual(t, ethcommon.Address{}, address)
	require.NotNil(t, tx)
	require.NotNil(t, swapContract)
}

func TestSwap_Claim(t *testing.T) {
	// Alice generates key
	keyPairAlice, err := monero.GenerateKeys()
	require.NoError(t, err)

	// Bob generates key
	keyPairBob, err := monero.GenerateKeys()
	require.NoError(t, err)
	secretBob := keyPairBob.SpendKeyBytes()

	// setup
	conn, err := ethclient.Dial("ws://127.0.0.1:8545")
	require.NoError(t, err)

	pk_a, err := crypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)
	pk_b, err := crypto.HexToECDSA(common.DefaultPrivKeyBob)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(common.GanacheChainID))
	authAlice.Value = big.NewInt(1000000000000)
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

	bobPub := pk_b.Public().(*ecdsa.PublicKey)
	bobAddr := crypto.PubkeyToAddress(*bobPub)
	claimHash := keyPairBob.SpendKey().Hash()
	refundHash := keyPairAlice.SpendKey().Hash()

	contractAddress, deployTx, swap, err := DeploySwap(authAlice, conn, claimHash, refundHash, bobAddr, defaultTimeoutDuration)
	require.NoError(t, err)
	fmt.Println("Deploy Tx Gas Cost:", deployTx.Gas())

	aliceBalanceAfter, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	require.NoError(t, err)
	fmt.Println("AliceBalanceAfter: ", aliceBalanceAfter)

	contractBalance, err := conn.BalanceAt(context.Background(), contractAddress, nil)
	require.NoError(t, err)
	require.Equal(t, contractBalance, big.NewInt(1000000000000))

	txOpts := &bind.TransactOpts{
		From:   authAlice.From,
		Signer: authAlice.Signer,
	}

	txOptsBob := &bind.TransactOpts{
		From:   authBob.From,
		Signer: authBob.Signer,
	}

	// Bob tries to claim before Alice has called ready, should fail
	var sb [32]byte
	copy(sb[:], secretBob)
	_, err = swap.Claim(txOptsBob, sb)
	require.Regexp(t, ".*too late or early to claim!", err)

	// Alice calls set_ready on the contract
	setReadyTx, err := swap.SetReady(txOpts)
	fmt.Println("setReady Tx Gas Cost:", setReadyTx.Gas())
	require.NoError(t, err)

	// The main transaction that we're testing. Should work
	tx, err := swap.Claim(txOptsBob, sb)
	require.NoError(t, err)

	// The Swap contract has self destructed: should have no balance AND no bytecode at the address
	contractBalance, err = conn.BalanceAt(context.Background(), contractAddress, nil)
	require.NoError(t, err)
	require.Equal(t, contractBalance.Uint64(), big.NewInt(0).Uint64())
	bytecode, err := conn.CodeAt(context.Background(), contractAddress, nil) // nil is latest block
	require.NoError(t, err)
	require.Empty(t, bytecode)

	fmt.Println("Tx details are:", tx.Gas())

	// TODO: check whether Bob's account balance has updated
	// bobBalanceAfter, err := conn.BalanceAt(context.Background(), authBob.From, nil)
	// fmt.Println("BobBalanceAfter: ", bobBalanceAfter)
	// require.NoError(t, err)

	// expected := big.NewInt(0).Sub(big.NewInt(1000000000000), tx.Cost())
	// require.Equal(t, expected, big.NewInt(0).Sub(bobBalanceAfter, bobBalanceBefore))
}

func TestSwap_Refund_Within_T0(t *testing.T) {
	// Alice generates key
	keyPairAlice, err := monero.GenerateKeys()
	require.NoError(t, err)

	// Bob generates key
	keyPairBob, err := monero.GenerateKeys()
	require.NoError(t, err)

	secretAlice := keyPairAlice.SpendKeyBytes()

	// setup
	conn, err := ethclient.Dial("ws://127.0.0.1:8545")
	require.NoError(t, err)

	pk_a, err := crypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)
	pk_b, err := crypto.HexToECDSA(common.DefaultPrivKeyBob)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(1337)) // ganache chainID
	require.NoError(t, err)
	authAlice.Value = big.NewInt(10)

	aliceBalanceBefore, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	require.NoError(t, err)
	fmt.Println("AliceBalanceBefore: ", aliceBalanceBefore)

	bobPub := pk_b.Public().(*ecdsa.PublicKey)
	bobAddr := crypto.PubkeyToAddress(*bobPub)
	claimHash := keyPairBob.SpendKey().Hash()
	refundHash := keyPairAlice.SpendKey().Hash()
	contractAddress, _, swap, err := DeploySwap(authAlice, conn, claimHash, refundHash, bobAddr, defaultTimeoutDuration)
	require.NoError(t, err)

	txOpts := &bind.TransactOpts{
		From:   authAlice.From,
		Signer: authAlice.Signer,
	}

	// Alice never calls set_ready on the contract, instead she just tries to Refund immidiately
	var sa [32]byte
	copy(sa[:], secretAlice)
	_, err = swap.Refund(txOpts, sa)
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

	// Bob generates key
	keyPairBob, err := monero.GenerateKeys()
	require.NoError(t, err)

	secretAlice := keyPairAlice.SpendKeyBytes()

	// setup
	conn, err := ethclient.Dial("ws://127.0.0.1:8545")
	require.NoError(t, err)

	pk_a, err := crypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)
	pk_b, err := crypto.HexToECDSA(common.DefaultPrivKeyBob)
	require.NoError(t, err)

	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(1337)) // ganache chainID
	authAlice.Value = big.NewInt(10)
	require.NoError(t, err)

	aliceBalanceBefore, err := conn.BalanceAt(context.Background(), authAlice.From, nil)
	require.NoError(t, err)
	fmt.Println("AliceBalanceBefore: ", aliceBalanceBefore)

	bobPub := pk_b.Public().(*ecdsa.PublicKey)
	bobAddr := crypto.PubkeyToAddress(*bobPub)
	claimHash := keyPairBob.SpendKey().Hash()
	refundHash := keyPairAlice.SpendKey().Hash()
	contractAddress, _, swap, err := DeploySwap(authAlice, conn, claimHash, refundHash, bobAddr, defaultTimeoutDuration)
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
	copy(sa[:], secretAlice)
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

	bytecode, err := conn.CodeAt(context.Background(), contractAddress, nil) // nil is latest block
	require.NoError(t, err)
	require.Empty(t, bytecode)
}
