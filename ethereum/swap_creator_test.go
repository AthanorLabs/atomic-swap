// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/crypto/secp256k1"
	"github.com/athanorlabs/atomic-swap/dleq"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/tests"
)

var (
	defaultTimeoutDuration = big.NewInt(60) // 60 seconds
	ethAssetAddress        = ethcommon.Address(types.EthAssetETH)
	defaultSwapValue       = big.NewInt(100)
)

func setupXMRTakerAuth(t *testing.T) (*bind.TransactOpts, *ethclient.Client, *ecdsa.PrivateKey) {
	conn, _ := tests.NewEthClient(t)
	pkA := tests.GetTakerTestKey(t)
	chainID, err := conn.ChainID(context.Background())
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(pkA, chainID)
	require.NoError(t, err)
	return auth, conn, pkA
}

func approveERC20(t *testing.T,
	auth *bind.TransactOpts,
	conn *ethclient.Client,
	erc20Contract *ERC20Mock,
	swapCreatorAddress ethcommon.Address,
	value *big.Int,
) {
	require.NotNil(t, erc20Contract)
	tx, err := erc20Contract.Approve(auth, swapCreatorAddress, value)
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call Approve: %d", receipt.GasUsed)
}

func testNewSwap(t *testing.T, asset ethcommon.Address, erc20Contract *ERC20Mock) {
	auth, conn, _ := setupXMRTakerAuth(t)
	address, tx, contract, err := DeploySwapCreator(auth, conn, ethcommon.Address{})
	require.NoError(t, err)
	require.NotEqual(t, ethcommon.Address{}, address)
	require.NotNil(t, tx)
	require.NotNil(t, contract)

	receipt, err := block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapCreator.sol: %d", receipt.GasUsed)

	owner := auth.From
	claimer := common.EthereumPrivateKeyToAddress(tests.GetMakerTestKey(t))

	var pubKeyClaim, pubKeyRefund [32]byte
	_, err = rand.Read(pubKeyClaim[:])
	require.NoError(t, err)
	_, err = rand.Read(pubKeyRefund[:])
	require.NoError(t, err)

	nonce, err := rand.Prime(rand.Reader, 256)
	require.NoError(t, err)

	value := defaultSwapValue
	isEthAsset := asset == ethAssetAddress

	if isEthAsset {
		auth.Value = value
	} else {
		approveERC20(t, auth, conn, erc20Contract, address, value)
	}

	tx, err = contract.NewSwap(
		auth,
		pubKeyClaim,
		pubKeyRefund,
		claimer,
		defaultTimeoutDuration,
		defaultTimeoutDuration,
		asset,
		value,
		nonce,
	)
	require.NoError(t, err)

	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call new_swap: %d", receipt.GasUsed)

	newSwapLogIndex := 0
	if !isEthAsset {
		newSwapLogIndex = 2
	}
	require.Equal(t, newSwapLogIndex+1, len(receipt.Logs))

	swapID, err := GetIDFromLog(receipt.Logs[newSwapLogIndex])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[newSwapLogIndex])
	require.NoError(t, err)

	// validate that off-chain swapID calculation matches the on-chain value
	swap := SwapCreatorSwap{
		Owner:        owner,
		Claimer:      claimer,
		PubKeyClaim:  pubKeyClaim,
		PubKeyRefund: pubKeyRefund,
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        asset,
		Value:        value,
		Nonce:        nonce,
	}

	// validate our off-net calculation of the SwapID
	require.Equal(t, types.Hash(swapID).Hex(), swap.SwapID().Hex())
}

func TestSwapCreator_NewSwap(t *testing.T) {
	testNewSwap(t, ethAssetAddress, nil)
}

func TestSwapCreator_Claim_vec(t *testing.T) {
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
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	_, tx, contract, err := DeploySwapCreator(auth, conn, ethcommon.Address{})
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapCreator.sol: %d", receipt.GasUsed)

	nonce := big.NewInt(0)
	auth.Value = defaultSwapValue
	tx, err = contract.NewSwap(auth, cmt, [32]byte{}, addr, defaultTimeoutDuration,
		defaultTimeoutDuration, ethcommon.Address(types.EthAssetETH), defaultSwapValue, nonce)
	require.NoError(t, err)
	auth.Value = nil

	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call new_swap: %d", receipt.GasUsed)

	require.Equal(t, 1, len(receipt.Logs))
	id, err := GetIDFromLog(receipt.Logs[0])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[0])
	require.NoError(t, err)

	swap := SwapCreatorSwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  cmt,
		PubKeyRefund: [32]byte{},
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        ethcommon.Address(types.EthAssetETH),
		Value:        defaultSwapValue,
		Nonce:        nonce,
	}

	// set contract to Ready
	tx, err = contract.SetReady(auth, swap)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call set_ready: %d", receipt.GasUsed)

	// now let's try to claim
	tx, err = contract.Claim(auth, swap, s)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call claim: %d", receipt.GasUsed)

	stage, err := contract.Swaps(nil, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}

func testClaim(t *testing.T, asset ethcommon.Address, newLogIndex int, value *big.Int, erc20Contract *ERC20Mock) {
	// generate claim secret and public key
	dleq := &dleq.DefaultDLEq{}
	proof, err := dleq.Prove()
	require.NoError(t, err)
	res, err := dleq.Verify(proof)
	require.NoError(t, err)

	// hash public key
	cmt := res.Secp256k1PublicKey().Keccak256()

	// deploy swap contract with claim key hash
	authOrig, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	// TODO: Rewrite this code to avoid the awkward use of txOpts. Code was using
	//       same TxOpts for multiple transactions and we needed a quick fix to get
	//       CI working.
	txOpts := *authOrig
	swapCreatorAddr, tx, contract, err := DeploySwapCreator(&txOpts, conn, ethcommon.Address{})
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapCreator.sol: %d", receipt.GasUsed)

	if asset != ethAssetAddress {
		approveERC20(t, authOrig, conn, erc20Contract, swapCreatorAddr, value)
	}

	nonce := big.NewInt(0)
	txOpts = *authOrig
	if asset == ethAssetAddress {
		txOpts.Value = value
	}

	tx, err = contract.NewSwap(&txOpts, cmt, [32]byte{}, addr,
		defaultTimeoutDuration, defaultTimeoutDuration, asset, value, nonce)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call new_swap: %d", receipt.GasUsed)

	require.Equal(t, newLogIndex+1, len(receipt.Logs))
	id, err := GetIDFromLog(receipt.Logs[newLogIndex])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[newLogIndex])
	require.NoError(t, err)

	swap := SwapCreatorSwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  cmt,
		PubKeyRefund: [32]byte{},
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        asset,
		Value:        value,
		Nonce:        nonce,
	}

	// set contract to Ready
	txOpts = *authOrig
	tx, err = contract.SetReady(&txOpts, swap)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	t.Logf("gas cost to call SetReady: %d", receipt.GasUsed)
	require.NoError(t, err)

	// now let's try to claim
	var s [32]byte
	secret := proof.Secret()
	copy(s[:], secret[:])
	txOpts = *authOrig
	tx, err = contract.Claim(&txOpts, swap, s)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call Claim: %d", receipt.GasUsed)

	stage, err := contract.Swaps(nil, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}

func TestSwapCreator_Claim_random(t *testing.T) {
	testClaim(t, ethAssetAddress, 0, defaultSwapValue, nil)
}

func testRefundBeforeT0(t *testing.T, asset ethcommon.Address, erc20Contract *ERC20Mock, newLogIndex int) {
	// generate refund secret and public key
	dleq := &dleq.DefaultDLEq{}
	proof, err := dleq.Prove()
	require.NoError(t, err)
	res, err := dleq.Verify(proof)
	require.NoError(t, err)

	// hash public key
	cmt := res.Secp256k1PublicKey().Keccak256()

	// deploy swap contract with refund key hash
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	address, tx, contract, err := DeploySwapCreator(auth, conn, ethcommon.Address{})
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapCreator.sol: %d", receipt.GasUsed)

	if asset != ethAssetAddress {
		approveERC20(t, auth, conn, erc20Contract, address, defaultSwapValue)
	}

	nonce := big.NewInt(0)
	auth.Value = defaultSwapValue
	tx, err = contract.NewSwap(auth, [32]byte{}, cmt, addr, defaultTimeoutDuration, defaultTimeoutDuration,
		asset, defaultSwapValue, nonce)
	require.NoError(t, err)
	auth.Value = nil

	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call new_swap: %d", receipt.GasUsed)

	require.Equal(t, newLogIndex+1, len(receipt.Logs))
	id, err := GetIDFromLog(receipt.Logs[newLogIndex])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[newLogIndex])
	require.NoError(t, err)

	swap := SwapCreatorSwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  [32]byte{},
		PubKeyRefund: cmt,
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        asset,
		Value:        defaultSwapValue,
		Nonce:        nonce,
	}

	// now let's try to refund
	var s [32]byte
	secret := proof.Secret()
	copy(s[:], secret[:])
	tx, err = contract.Refund(auth, swap, s)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call Refund: %d", receipt.GasUsed)

	stage, err := contract.Swaps(nil, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}

func TestSwapCreator_Refund_beforeT0(t *testing.T) {
	testRefundBeforeT0(t, ethAssetAddress, nil, 0)
}

func testRefundAfterT1(t *testing.T, asset ethcommon.Address, erc20Contract *ERC20Mock, newLogIndex int) {
	// generate refund secret and public key
	dleq := &dleq.DefaultDLEq{}
	proof, err := dleq.Prove()
	require.NoError(t, err)
	res, err := dleq.Verify(proof)
	require.NoError(t, err)

	// hash public key
	cmt := res.Secp256k1PublicKey().Keccak256()

	// deploy swap contract with refund key hash
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	address, tx, contract, err := DeploySwapCreator(auth, conn, ethcommon.Address{})
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapCreator.sol: %d", receipt.GasUsed)

	if asset != ethAssetAddress {
		approveERC20(t, auth, conn, erc20Contract, address, defaultSwapValue)
	}

	nonce := big.NewInt(0)
	timeout := big.NewInt(1) // T1 expires before we get the receipt for new_swap TX
	auth.Value = defaultSwapValue
	tx, err = contract.NewSwap(auth, [32]byte{}, cmt, addr, timeout, timeout,
		asset, defaultSwapValue, nonce)
	require.NoError(t, err)
	auth.Value = nil

	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call new_swap: %d", receipt.GasUsed)

	require.Equal(t, newLogIndex+1, len(receipt.Logs))
	id, err := GetIDFromLog(receipt.Logs[newLogIndex])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[newLogIndex])
	require.NoError(t, err)

	<-time.After(time.Until(time.Unix(t1.Int64()+1, 0)))

	swap := SwapCreatorSwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  [32]byte{},
		PubKeyRefund: cmt,
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        asset,
		Value:        defaultSwapValue,
		Nonce:        nonce,
	}

	// now let's try to refund
	var s [32]byte
	secret := proof.Secret()
	copy(s[:], secret[:])
	tx, err = contract.Refund(auth, swap, s)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call Refund: %d", receipt.GasUsed)

	callOpts := &bind.CallOpts{
		From:    crypto.PubkeyToAddress(*pub),
		Context: context.Background(),
	}

	stage, err := contract.Swaps(callOpts, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}

func TestSwapCreator_Refund_afterT1(t *testing.T) {
	testRefundAfterT1(t, ethAssetAddress, nil, 0)
}

func TestSwapCreator_MultipleSwaps(t *testing.T) {
	// test case where contract has multiple swaps happening at once
	conn, chainID := tests.NewEthClient(t)

	pkContractCreator := tests.GetTestKeyByIndex(t, 0)
	auth, err := bind.NewKeyedTransactorWithChainID(pkContractCreator, chainID)
	require.NoError(t, err)

	_, tx, contract, err := DeploySwapCreator(auth, conn, ethcommon.Address{})
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapCreator.sol: %d", receipt.GasUsed)

	const numSwaps = 16
	type swapCase struct {
		index     int // index in the swap array
		walletKey *ecdsa.PrivateKey
		id        [32]byte
		secret    [32]byte
		swap      SwapCreatorSwap
	}

	getAuth := func(sc *swapCase) *bind.TransactOpts {
		auth, err := bind.NewKeyedTransactorWithChainID(sc.walletKey, chainID)
		require.NoError(t, err)
		return auth
	}

	swapCases := [numSwaps]swapCase{}

	// setup all swap instances
	for i := 0; i < numSwaps; i++ {
		sc := &swapCases[i]
		sc.index = i

		// generate claim secret and public key
		dleq := &dleq.DefaultDLEq{}
		proof, err := dleq.Prove()
		require.NoError(t, err)
		res, err := dleq.Verify(proof)
		require.NoError(t, err)

		secret := proof.Secret()
		copy(sc.secret[:], secret[:])

		sc.walletKey = tests.GetTestKeyByIndex(t, i)
		addrSwap := crypto.PubkeyToAddress(*sc.walletKey.Public().(*ecdsa.PublicKey))

		sc.swap = SwapCreatorSwap{
			Owner:        addrSwap,
			Claimer:      addrSwap,
			PubKeyClaim:  res.Secp256k1PublicKey().Keccak256(),
			PubKeyRefund: [32]byte{}, // no one calls refund in this test
			Timeout0:     nil,        // timeouts initialised when swap is created
			Timeout1:     nil,
			Asset:        ethcommon.Address(types.EthAssetETH),
			Value:        defaultSwapValue,
			Nonce:        big.NewInt(int64(i)),
		}
	}

	// We create all transactions in parallel, so the transactions of each swap stage can get bundled up
	// into one or two blocks and greatly speed up the test.
	var wg sync.WaitGroup

	// create swap instances
	wg.Add(numSwaps)
	for i := 0; i < numSwaps; i++ {
		go func(sc *swapCase) {
			defer wg.Done()
			auth := getAuth(sc)
			auth.Value = sc.swap.Value
			tx, err := contract.NewSwap(
				auth,
				sc.swap.PubKeyClaim,
				sc.swap.PubKeyRefund,
				sc.swap.Claimer,
				defaultTimeoutDuration,
				defaultTimeoutDuration,
				ethcommon.Address(types.EthAssetETH),
				sc.swap.Value,
				sc.swap.Nonce,
			)
			require.NoError(t, err)
			auth.Value = nil

			receipt, err := block.WaitForReceipt(context.Background(), conn, tx.Hash())
			require.NoError(t, err)
			t.Logf("gas cost to call new_swap[%d]: %d", sc.index, receipt.GasUsed)

			require.Equal(t, 1, len(receipt.Logs))
			sc.id, err = GetIDFromLog(receipt.Logs[0])
			require.NoError(t, err)

			sc.swap.Timeout0, sc.swap.Timeout1, err = GetTimeoutsFromLog(receipt.Logs[0])
			require.NoError(t, err)
		}(&swapCases[i])
	}
	wg.Wait() // all swaps created

	// set all swaps to Ready
	wg.Add(numSwaps)
	for i := 0; i < numSwaps; i++ {
		go func(sc *swapCase) {
			defer wg.Done()
			tx, err := contract.SetReady(getAuth(sc), sc.swap)
			require.NoError(t, err)
			receipt, err := block.WaitForReceipt(context.Background(), conn, tx.Hash())
			require.NoError(t, err)
			t.Logf("gas cost to call SetReady[%d]: %d", sc.index, receipt.GasUsed)
		}(&swapCases[i])
	}
	wg.Wait() // set_ready called on all swaps

	// call claim on all the swaps
	wg.Add(numSwaps)
	for i := 0; i < numSwaps; i++ {
		go func(sc *swapCase) {
			defer wg.Done()
			tx, err := contract.Claim(getAuth(sc), sc.swap, sc.secret)
			require.NoError(t, err)
			receipt, err := block.WaitForReceipt(context.Background(), conn, tx.Hash())
			require.NoError(t, err)
			t.Logf("gas cost to call Claim[%d]: %d", sc.index, receipt.GasUsed)
		}(&swapCases[i])
	}
	wg.Wait() // claim called on all swaps

	// ensure all swaps are completed
	wg.Add(numSwaps)
	for i := 0; i < numSwaps; i++ {
		go func(sc *swapCase) {
			defer wg.Done()
			stage, err := contract.Swaps(nil, sc.id)
			require.NoError(t, err)
			require.Equal(t, StageToString(StageCompleted), StageToString(stage))
		}(&swapCases[i])
	}
	wg.Wait() // status of all swaps checked
}
