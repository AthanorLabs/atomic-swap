package swapfactory

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
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
	"github.com/noot/atomic-swap/tests"
)

var defaultTimeoutDuration = big.NewInt(60) // 60 seconds

func setupXMRTakerAuth(t *testing.T) (*bind.TransactOpts, *ethclient.Client, *ecdsa.PrivateKey) {
	conn, chainID := tests.NewEthClient(t)
	pkA, err := crypto.HexToECDSA(tests.GetTakerTestKey(t))
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(pkA, chainID)
	require.NoError(t, err)
	return auth, conn, pkA
}

func TestSwapFactory_NewSwap(t *testing.T) {
	auth, conn, _ := setupXMRTakerAuth(t)
	defer conn.Close()
	address, tx, contract, err := DeploySwapFactory(auth, conn)
	require.NoError(t, err)
	require.NotEqual(t, ethcommon.Address{}, address)
	require.NotNil(t, tx)
	require.NotNil(t, contract)
	t.Logf("gas cost to deploy SwapFactory.sol: %d", tx.Gas())
	_, err = bind.WaitDeployed(context.Background(), conn, tx)
	require.NoError(t, err)

	nonce := big.NewInt(0)
	tx, err = contract.NewSwap(auth, [32]byte{}, [32]byte{},
		ethcommon.Address{}, defaultTimeoutDuration, nonce)
	require.NoError(t, err)
	tests.MineTransaction(t, conn, tx)

	t.Logf("gas cost to call new_swap: %d", tx.Gas())
}

func TestSwapFactory_Claim_vec(t *testing.T) {
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
	auth, conn, pkA := setupXMRTakerAuth(t)
	defer conn.Close()
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	_, tx, contract, err := DeploySwapFactory(auth, conn)
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapFactory.sol: %d", tx.Gas())
	_, err = bind.WaitDeployed(context.Background(), conn, tx)
	require.NoError(t, err)

	nonce := big.NewInt(0)
	tx, err = contract.NewSwap(auth, cmt, [32]byte{}, addr,
		defaultTimeoutDuration, nonce)
	require.NoError(t, err)
	t.Logf("gas cost to call new_swap: %d", tx.Gas())
	receipt := tests.MineTransaction(t, conn, tx)

	require.Equal(t, 1, len(receipt.Logs))
	id, err := GetIDFromLog(receipt.Logs[0])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[0])
	require.NoError(t, err)

	swap := SwapFactorySwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  cmt,
		PubKeyRefund: [32]byte{},
		Timeout0:     t0,
		Timeout1:     t1,
		Value:        big.NewInt(0),
		Nonce:        nonce,
	}

	// set contract to Ready
	tx, err = contract.SetReady(auth, swap)
	require.NoError(t, err)
	t.Logf("gas cost to call set_ready: %d", tx.Gas())
	tests.MineTransaction(t, conn, tx)

	// now let's try to claim
	tx, err = contract.Claim(auth, swap, s)
	require.NoError(t, err)
	t.Logf("gas cost to call claim: %d", tx.Gas())
	tests.MineTransaction(t, conn, tx)

	callOpts := &bind.CallOpts{
		From:    crypto.PubkeyToAddress(*pub),
		Context: context.Background(),
	}

	stage, err := contract.Swaps(callOpts, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}

func TestSwap_Claim_random(t *testing.T) {
	// generate claim secret and public key
	dleq := &dleq.CGODLEq{}
	proof, err := dleq.Prove()
	require.NoError(t, err)
	res, err := dleq.Verify(proof)
	require.NoError(t, err)

	// hash public key
	cmt := res.Secp256k1PublicKey().Keccak256()

	// deploy swap contract with claim key hash
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	_, tx, contract, err := DeploySwapFactory(auth, conn)
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapFactory.sol: %d", tx.Gas())
	_, err = bind.WaitDeployed(context.Background(), conn, tx)
	require.NoError(t, err)

	nonce := big.NewInt(0)
	tx, err = contract.NewSwap(auth, cmt, [32]byte{}, addr,
		defaultTimeoutDuration, nonce)
	require.NoError(t, err)
	receipt := tests.MineTransaction(t, conn, tx)
	t.Logf("gas cost to call new_swap: %d", tx.Gas())

	require.Equal(t, 1, len(receipt.Logs))
	id, err := GetIDFromLog(receipt.Logs[0])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[0])
	require.NoError(t, err)

	swap := SwapFactorySwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  cmt,
		PubKeyRefund: [32]byte{},
		Timeout0:     t0,
		Timeout1:     t1,
		Value:        big.NewInt(0),
		Nonce:        nonce,
	}

	// set contract to Ready
	tx, err = contract.SetReady(auth, swap)
	require.NoError(t, err)
	t.Logf("gas cost to call SetReady: %d", tx.Gas())
	tests.MineTransaction(t, conn, tx)

	// now let's try to claim
	var s [32]byte
	secret := proof.Secret()
	copy(s[:], common.Reverse(secret[:]))
	tx, err = contract.Claim(auth, swap, s)
	require.NoError(t, err)
	tests.MineTransaction(t, conn, tx)
	t.Logf("gas cost to call Claim: %d", tx.Gas())

	callOpts := &bind.CallOpts{
		From:    crypto.PubkeyToAddress(*pub),
		Context: context.Background(),
	}

	stage, err := contract.Swaps(callOpts, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}

func TestSwap_Refund_beforeT0(t *testing.T) {
	// generate refund secret and public key
	dleq := &dleq.CGODLEq{}
	proof, err := dleq.Prove()
	require.NoError(t, err)
	res, err := dleq.Verify(proof)
	require.NoError(t, err)

	// hash public key
	cmt := res.Secp256k1PublicKey().Keccak256()

	// deploy swap contract with refund key hash
	auth, conn, pkA := setupXMRTakerAuth(t)
	defer conn.Close()
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	_, tx, contract, err := DeploySwapFactory(auth, conn)
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapFactory.sol: %d", tx.Gas())
	_, err = bind.WaitDeployed(context.Background(), conn, tx)
	require.NoError(t, err)

	nonce := big.NewInt(0)
	tx, err = contract.NewSwap(auth, [32]byte{}, cmt, addr,
		defaultTimeoutDuration, nonce)
	require.NoError(t, err)
	receipt := tests.MineTransaction(t, conn, tx)
	t.Logf("gas cost to call new_swap: %d", tx.Gas())

	require.Equal(t, 1, len(receipt.Logs))
	id, err := GetIDFromLog(receipt.Logs[0])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[0])
	require.NoError(t, err)

	swap := SwapFactorySwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  [32]byte{},
		PubKeyRefund: cmt,
		Timeout0:     t0,
		Timeout1:     t1,
		Value:        big.NewInt(0),
		Nonce:        nonce,
	}

	// now let's try to refund
	var s [32]byte
	secret := proof.Secret()
	copy(s[:], common.Reverse(secret[:]))
	tx, err = contract.Refund(auth, swap, s)
	require.NoError(t, err)
	tests.MineTransaction(t, conn, tx)
	t.Logf("gas cost to call Refund: %d", tx.Gas())

	callOpts := &bind.CallOpts{
		From:    crypto.PubkeyToAddress(*pub),
		Context: context.Background(),
	}

	stage, err := contract.Swaps(callOpts, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}

func TestSwap_Refund_afterT1(t *testing.T) {
	// generate refund secret and public key
	dleq := &dleq.CGODLEq{}
	proof, err := dleq.Prove()
	require.NoError(t, err)
	res, err := dleq.Verify(proof)
	require.NoError(t, err)

	// hash public key
	cmt := res.Secp256k1PublicKey().Keccak256()

	// deploy swap contract with refund key hash
	auth, conn, pkA := setupXMRTakerAuth(t)
	defer conn.Close()
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	_, tx, contract, err := DeploySwapFactory(auth, conn)
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapFactory.sol: %d", tx.Gas())
	_, err = bind.WaitDeployed(context.Background(), conn, tx)
	require.NoError(t, err)

	nonce := big.NewInt(0)
	tx, err = contract.NewSwap(auth, [32]byte{}, cmt, addr, defaultTimeoutDuration, nonce)
	require.NoError(t, err)
	receipt := tests.MineTransaction(t, conn, tx)
	t.Logf("gas cost to call new_swap: %d", tx.Gas())

	require.Equal(t, 1, len(receipt.Logs))
	id, err := GetIDFromLog(receipt.Logs[0])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[0])
	require.NoError(t, err)

	swap := SwapFactorySwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  [32]byte{},
		PubKeyRefund: cmt,
		Timeout0:     t0,
		Timeout1:     t1,
		Value:        big.NewInt(0),
		Nonce:        nonce,
	}

	// fast forward past t1
	rpcClient, err := rpc.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	var result string
	err = rpcClient.Call(&result, "evm_snapshot")
	require.NoError(t, err)

	err = rpcClient.Call(nil, "evm_increaseTime", defaultTimeoutDuration.Int64()*2+60)
	require.NoError(t, err)

	defer func() {
		var ok bool
		err = rpcClient.Call(&ok, "evm_revert", result)
		require.NoError(t, err)
	}()

	// now let's try to refund
	var s [32]byte
	secret := proof.Secret()
	copy(s[:], common.Reverse(secret[:]))
	tx, err = contract.Refund(auth, swap, s)
	require.NoError(t, err)
	tests.MineTransaction(t, conn, tx)
	t.Logf("gas cost to call Refund: %d", tx.Gas())

	callOpts := &bind.CallOpts{
		From:    crypto.PubkeyToAddress(*pub),
		Context: context.Background(),
	}

	stage, err := contract.Swaps(callOpts, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}

func TestSwap_MultipleSwaps(t *testing.T) {
	// test case where contract has multiple swaps happening at once
	auth, conn, pkA := setupXMRTakerAuth(t)
	defer conn.Close()
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	_, tx, contract, err := DeploySwapFactory(auth, conn)
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapFactory.sol: %d", tx.Gas())
	_, err = bind.WaitDeployed(context.Background(), conn, tx)
	require.NoError(t, err)

	numSwaps := 16
	type swapCase struct {
		id     [32]byte
		secret [32]byte
		swap   SwapFactorySwap
	}

	// setup all swap instances in contract
	swapCases := []*swapCase{}
	for i := 0; i < numSwaps; i++ {
		sc := &swapCase{}

		// generate claim secret and public key
		dleq := &dleq.CGODLEq{}
		proof, err := dleq.Prove() //nolint:govet
		require.NoError(t, err)
		res, err := dleq.Verify(proof)
		require.NoError(t, err)

		// hash public key
		cmt := res.Secp256k1PublicKey().Keccak256()
		secret := proof.Secret()
		copy(sc.secret[:], common.Reverse(secret[:]))

		nonce := big.NewInt(int64(i))
		tx, err = contract.NewSwap(auth, cmt, [32]byte{}, addr, defaultTimeoutDuration, nonce)
		require.NoError(t, err)
		receipt := tests.MineTransaction(t, conn, tx)
		t.Logf("gas cost to call new_swap: %d", tx.Gas())

		require.Equal(t, 1, len(receipt.Logs))
		sc.id, err = GetIDFromLog(receipt.Logs[0])
		require.NoError(t, err)

		t0, t1, err := GetTimeoutsFromLog(receipt.Logs[0])
		require.NoError(t, err)

		sc.swap = SwapFactorySwap{
			Owner:        addr,
			Claimer:      addr,
			PubKeyClaim:  cmt,
			PubKeyRefund: [32]byte{},
			Timeout0:     t0,
			Timeout1:     t1,
			Value:        big.NewInt(0),
			Nonce:        nonce,
		}

		swapCases = append(swapCases, sc)
	}

	for _, sc := range swapCases {
		// set contract to Ready
		tx, err = contract.SetReady(auth, sc.swap)
		require.NoError(t, err)
		tests.MineTransaction(t, conn, tx)
		t.Logf("gas cost to call SetReady: %d", tx.Gas())

		// now let's try to claim
		tx, err = contract.Claim(auth, sc.swap, sc.secret)
		require.NoError(t, err)
		tests.MineTransaction(t, conn, tx)
		t.Logf("gas cost to call Claim: %d", tx.Gas())

		callOpts := &bind.CallOpts{
			From:    crypto.PubkeyToAddress(*pub),
			Context: context.Background(),
		}

		stage, err := contract.Swaps(callOpts, sc.id)
		require.NoError(t, err)
		require.Equal(t, StageCompleted, stage)
	}
}
