// Copyright 2023 The AthanorLabs/atomic-swap Authors
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
	ethtypes "github.com/ethereum/go-ethereum/core/types"
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
	defaultSwapValue       = big.NewInt(100)
	dummySwapKey           = [32]byte{1} // dummy non-zero value for claim/refund key
)

func getAuth(t *testing.T, pk *ecdsa.PrivateKey) *bind.TransactOpts {
	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(common.GanacheChainID))
	require.NoError(t, err)
	return txOpts
}

func getReceipt(t *testing.T, ec *ethclient.Client, tx *ethtypes.Transaction) *ethtypes.Receipt {
	receipt, err := block.WaitForReceipt(context.Background(), ec, tx.Hash())
	require.NoError(t, err)
	return receipt
}

func approveERC20(t *testing.T,
	ec *ethclient.Client,
	pk *ecdsa.PrivateKey,
	erc20Contract *TestERC20,
	swapCreatorAddress ethcommon.Address,
	value *big.Int,
) {
	require.NotNil(t, erc20Contract)

	tx, err := erc20Contract.Approve(getAuth(t, pk), swapCreatorAddress, value)
	require.NoError(t, err)
	receipt := getReceipt(t, ec, tx)

	t.Logf("gas cost to call Approve %d (delta %d)", receipt.GasUsed, MaxTokenApproveGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, MaxTokenApproveGas, int(receipt.GasUsed), "Token Approve")
}

func deploySwapCreator(t *testing.T, ec *ethclient.Client, pk *ecdsa.PrivateKey) (ethcommon.Address, *SwapCreator) {
	swapCreatorAddr, tx, swapCreator, err := DeploySwapCreator(getAuth(t, pk), ec)
	require.NoError(t, err)
	receipt := getReceipt(t, ec, tx)

	t.Logf("gas cost to deploy SwapCreator.sol: %d (delta %d)",
		receipt.GasUsed, maxSwapCreatorDeployGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, maxSwapCreatorDeployGas, int(receipt.GasUsed), "deploy SwapCreator")

	return swapCreatorAddr, swapCreator
}

func testNewSwap(t *testing.T, asset types.EthAsset, erc20Contract *TestERC20) {
	pk := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)

	swapCreatorAddr, swapCreator := deploySwapCreator(t, ec, pk)

	owner := crypto.PubkeyToAddress(pk.PublicKey)
	claimer := common.EthereumPrivateKeyToAddress(tests.GetMakerTestKey(t))

	var pubKeyClaim, pubKeyRefund [32]byte
	_, err := rand.Read(pubKeyClaim[:])
	require.NoError(t, err)
	_, err = rand.Read(pubKeyRefund[:])
	require.NoError(t, err)

	nonce, err := rand.Prime(rand.Reader, 256)
	require.NoError(t, err)

	txOpts := getAuth(t, pk)
	value := defaultSwapValue
	if asset.IsETH() {
		txOpts.Value = value
	} else {
		approveERC20(t, ec, pk, erc20Contract, swapCreatorAddr, value)
	}

	tx, err := swapCreator.NewSwap(
		txOpts,
		pubKeyClaim,
		pubKeyRefund,
		claimer,
		defaultTimeoutDuration,
		defaultTimeoutDuration,
		asset.Address(),
		value,
		nonce,
	)
	require.NoError(t, err)

	receipt := getReceipt(t, ec, tx)
	if asset.IsETH() {
		t.Logf("gas cost to call ETH NewSwap: %d (delta %d)",
			receipt.GasUsed, MaxNewSwapETHGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxNewSwapETHGas, int(receipt.GasUsed), "ETH NewSwap")
	} else {
		t.Logf("gas cost to call token NewSwap: %d (delta %d)",
			receipt.GasUsed, MaxNewSwapTokenGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxNewSwapTokenGas, int(receipt.GasUsed), "Token NewSwap")
	}

	newSwapLogIndex := 0
	if asset.IsToken() {
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
		Asset:        asset.Address(),
		Value:        value,
		Nonce:        nonce,
	}

	// validate our off-net calculation of the SwapID
	require.Equal(t, types.Hash(swapID).Hex(), swap.SwapID().Hex())
}

func TestSwapCreator_NewSwap(t *testing.T) {
	testNewSwap(t, types.EthAssetETH, nil)
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
	pkA := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	addr := crypto.PubkeyToAddress(pkA.PublicKey)

	_, swapCreator := deploySwapCreator(t, ec, pkA)

	txOpts := getAuth(t, pkA)
	txOpts.Value = defaultSwapValue

	nonce := big.NewInt(0)
	tx, err := swapCreator.NewSwap(txOpts, cmt, dummySwapKey, addr, defaultTimeoutDuration,
		defaultTimeoutDuration, ethcommon.Address(types.EthAssetETH), defaultSwapValue, nonce)
	require.NoError(t, err)

	receipt := getReceipt(t, ec, tx)
	t.Logf("gas cost to call ETH NewSwap: %d (delta %d)", receipt.GasUsed, MaxNewSwapETHGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, MaxNewSwapETHGas, int(receipt.GasUsed), "ETH NewSwap")

	require.Equal(t, 1, len(receipt.Logs))
	id, err := GetIDFromLog(receipt.Logs[0])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[0])
	require.NoError(t, err)

	swap := SwapCreatorSwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  cmt,
		PubKeyRefund: dummySwapKey,
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        ethcommon.Address(types.EthAssetETH),
		Value:        defaultSwapValue,
		Nonce:        nonce,
	}

	// set contract to Ready
	tx, err = swapCreator.SetReady(getAuth(t, pkA), swap)
	require.NoError(t, err)
	receipt = getReceipt(t, ec, tx)

	t.Logf("gas cost to call SetReady: %d (delta %d)", receipt.GasUsed, MaxSetReadyGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, MaxSetReadyGas, int(receipt.GasUsed), "SetReady")

	// now let's try to claim
	tx, err = swapCreator.Claim(getAuth(t, pkA), swap, s)
	require.NoError(t, err)
	receipt = getReceipt(t, ec, tx)
	t.Logf("gas cost to call ETH Claim: %d (delta %d)", receipt.GasUsed, MaxClaimETHGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, MaxClaimETHGas, int(receipt.GasUsed), "ETH Claim")

	stage, err := swapCreator.Swaps(nil, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}

func testClaim(t *testing.T, asset types.EthAsset, newLogIndex int, value *big.Int, erc20Contract *TestERC20) {
	// generate claim secret and public key
	dleq := &dleq.DefaultDLEq{}
	proof, err := dleq.Prove()
	require.NoError(t, err)
	res, err := dleq.Verify(proof)
	require.NoError(t, err)

	// hash public key
	cmt := res.Secp256k1PublicKey().Keccak256()

	// deploy swap contract with claim key hash
	pkA := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	addr := crypto.PubkeyToAddress(pkA.PublicKey)

	swapCreatorAddr, swapCreator := deploySwapCreator(t, ec, pkA)

	if asset.IsToken() {
		approveERC20(t, ec, pkA, erc20Contract, swapCreatorAddr, value)
	}

	txOpts := getAuth(t, pkA)
	require.NoError(t, err)
	if asset.IsETH() {
		txOpts.Value = value
	}

	nonce := GenerateNewSwapNonce()
	tx, err := swapCreator.NewSwap(txOpts, cmt, dummySwapKey, addr,
		defaultTimeoutDuration, defaultTimeoutDuration, asset.Address(), value, nonce)
	require.NoError(t, err)

	receipt := getReceipt(t, ec, tx)
	if asset.IsETH() {
		t.Logf("gas cost to call ETH NewSwap: %d (delta %d)",
			receipt.GasUsed, MaxNewSwapETHGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxNewSwapETHGas, int(receipt.GasUsed), "ETH NewSwap")
	} else {
		t.Logf("gas cost to call token NewSwap: %d (delta %d)",
			receipt.GasUsed, MaxNewSwapTokenGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxNewSwapTokenGas, int(receipt.GasUsed), "Token NewSwap")
	}

	require.Equal(t, newLogIndex+1, len(receipt.Logs))
	id, err := GetIDFromLog(receipt.Logs[newLogIndex])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[newLogIndex])
	require.NoError(t, err)

	swap := SwapCreatorSwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  cmt,
		PubKeyRefund: dummySwapKey,
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        asset.Address(),
		Value:        value,
		Nonce:        nonce,
	}

	// ensure we can't claim before setting contract to Ready
	_, err = swapCreator.Claim(getAuth(t, pkA), swap, proof.Secret())
	require.ErrorContains(t, err, "VM Exception while processing transaction: revert")

	// set contract to Ready
	tx, err = swapCreator.SetReady(getAuth(t, pkA), swap)
	require.NoError(t, err)
	receipt = getReceipt(t, ec, tx)
	t.Logf("gas cost to call SetReady: %d (delta %d)", receipt.GasUsed, MaxSetReadyGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, MaxSetReadyGas, int(receipt.GasUsed), "SetReady")

	// now let's try to claim
	tx, err = swapCreator.Claim(getAuth(t, pkA), swap, proof.Secret())
	require.NoError(t, err)
	receipt = getReceipt(t, ec, tx)

	if asset.IsETH() {
		t.Logf("gas cost to call ETH Claim: %d (delta %d)", receipt.GasUsed, MaxClaimETHGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxClaimETHGas, int(receipt.GasUsed), "ETH Claim")
	} else {
		t.Logf("gas cost to call token Claim: %d (delta %d)", receipt.GasUsed, MaxClaimTokenGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxClaimTokenGas, int(receipt.GasUsed), "Token Claim")
	}

	stage, err := swapCreator.Swaps(nil, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}

func TestSwapCreator_Claim_random(t *testing.T) {
	testClaim(t, types.EthAssetETH, 0, defaultSwapValue, nil)
}

func testRefundBeforeT0(t *testing.T, asset types.EthAsset, erc20Contract *TestERC20, newLogIndex int) {
	// generate refund secret and public key
	dleq := &dleq.DefaultDLEq{}
	proof, err := dleq.Prove()
	require.NoError(t, err)
	res, err := dleq.Verify(proof)
	require.NoError(t, err)

	// hash public key
	cmt := res.Secp256k1PublicKey().Keccak256()

	// deploy swap contract with refund key hash
	pkA := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	addr := crypto.PubkeyToAddress(pkA.PublicKey)

	swapCreatorAddr, swapCreator := deploySwapCreator(t, ec, pkA)

	if asset.IsToken() {
		approveERC20(t, ec, pkA, erc20Contract, swapCreatorAddr, defaultSwapValue)
	}

	txOpts := getAuth(t, pkA)
	txOpts.Value = defaultSwapValue

	nonce := GenerateNewSwapNonce()
	tx, err := swapCreator.NewSwap(txOpts, dummySwapKey, cmt, addr, defaultTimeoutDuration, defaultTimeoutDuration,
		asset.Address(), defaultSwapValue, nonce)
	require.NoError(t, err)

	receipt := getReceipt(t, ec, tx)
	if asset.IsETH() {
		t.Logf("gas cost to call ETH NewSwap: %d (delta %d)",
			receipt.GasUsed, MaxNewSwapETHGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxNewSwapETHGas, int(receipt.GasUsed), "ETH NewSwap")
	} else {
		t.Logf("gas cost to call token NewSwap: %d (delta %d)",
			receipt.GasUsed, MaxNewSwapTokenGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxNewSwapTokenGas, int(receipt.GasUsed), "ETH NewSwap")
	}

	require.Equal(t, newLogIndex+1, len(receipt.Logs))
	id, err := GetIDFromLog(receipt.Logs[newLogIndex])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[newLogIndex])
	require.NoError(t, err)

	swap := SwapCreatorSwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  dummySwapKey,
		PubKeyRefund: cmt,
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        asset.Address(),
		Value:        defaultSwapValue,
		Nonce:        nonce,
	}

	// now let's try to refund
	tx, err = swapCreator.Refund(getAuth(t, pkA), swap, proof.Secret())
	require.NoError(t, err)
	receipt = getReceipt(t, ec, tx)

	if asset.IsETH() {
		t.Logf("gas cost to call ETH Refund: %d (delta %d)",
			receipt.GasUsed, MaxRefundETHGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxRefundETHGas, int(receipt.GasUsed), "ETH Refund")
	} else {
		t.Logf("gas cost to call token Refund: %d (delta %d)",
			receipt.GasUsed, MaxRefundTokenGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxRefundTokenGas, int(receipt.GasUsed), "Token Refund")
	}

	stage, err := swapCreator.Swaps(nil, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}

func TestSwapCreator_Refund_beforeT0(t *testing.T) {
	testRefundBeforeT0(t, types.EthAssetETH, nil, 0)
}

func testRefundAfterT1(t *testing.T, asset types.EthAsset, erc20Contract *TestERC20, newLogIndex int) {
	ctx := context.Background()

	// generate refund secret and public key
	dleq := &dleq.DefaultDLEq{}
	proof, err := dleq.Prove()
	require.NoError(t, err)
	res, err := dleq.Verify(proof)
	require.NoError(t, err)

	// hash public key
	cmt := res.Secp256k1PublicKey().Keccak256()

	// deploy swap contract with refund key hash
	pkA := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	addr := crypto.PubkeyToAddress(pkA.PublicKey)

	swapCreatorAddr, swapCreator := deploySwapCreator(t, ec, pkA)

	if asset.IsToken() {
		approveERC20(t, ec, pkA, erc20Contract, swapCreatorAddr, defaultSwapValue)
	}

	txOpts := getAuth(t, pkA)
	txOpts.Value = defaultSwapValue

	nonce := GenerateNewSwapNonce()
	timeout := big.NewInt(3)
	tx, err := swapCreator.NewSwap(txOpts, dummySwapKey, cmt, addr, timeout, timeout,
		asset.Address(), defaultSwapValue, nonce)
	require.NoError(t, err)
	receipt := getReceipt(t, ec, tx)

	if asset.IsETH() {
		t.Logf("gas cost to call ETH NewSwap: %d (delta %d)",
			receipt.GasUsed, MaxNewSwapETHGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxNewSwapETHGas, int(receipt.GasUsed), "ETH NewSwap")
	} else {
		t.Logf("gas cost to call token NewSwap: %d (delta %d)",
			receipt.GasUsed, MaxNewSwapTokenGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxNewSwapTokenGas, int(receipt.GasUsed), "Token NewSwap")
	}

	require.Equal(t, newLogIndex+1, len(receipt.Logs))
	id, err := GetIDFromLog(receipt.Logs[newLogIndex])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[newLogIndex])
	require.NoError(t, err)

	// ensure we can't refund between T0 and T1
	<-time.After(time.Until(time.Unix(t0.Int64()+1, 0)))
	swap := SwapCreatorSwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  dummySwapKey,
		PubKeyRefund: cmt,
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        asset.Address(),
		Value:        defaultSwapValue,
		Nonce:        nonce,
	}

	secret := proof.Secret()
	tx, err = swapCreator.Refund(getAuth(t, pkA), swap, secret)
	require.NoError(t, err)
	_, err = block.WaitForReceipt(ctx, ec, tx.Hash())
	require.ErrorContains(t, err, "VM Exception while processing transaction: revert")

	<-time.After(time.Until(time.Unix(t1.Int64()+1, 0)))

	// now let's try to refund
	tx, err = swapCreator.Refund(getAuth(t, pkA), swap, secret)
	require.NoError(t, err)
	receipt = getReceipt(t, ec, tx)

	if asset.IsETH() {
		t.Logf("gas cost to call ETH Refund: %d (delta %d)",
			receipt.GasUsed, MaxRefundETHGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxRefundETHGas, int(receipt.GasUsed), "ETH Refund")
	} else {
		t.Logf("gas cost to call token Refund: %d (delta %d)",
			receipt.GasUsed, MaxRefundTokenGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, MaxRefundTokenGas, int(receipt.GasUsed), "Token Refund")
	}

	callOpts := &bind.CallOpts{Context: ctx}
	stage, err := swapCreator.Swaps(callOpts, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}

func TestSwapCreator_Refund_afterT1(t *testing.T) {
	testRefundAfterT1(t, types.EthAssetETH, nil, 0)
}

// test case where contract has multiple swaps happening at once
func TestSwapCreator_MultipleSwaps(t *testing.T) {
	pkContractCreator := tests.GetTestKeyByIndex(t, 0)
	ec, _ := tests.NewEthClient(t)

	_, swapCreator := deploySwapCreator(t, ec, pkContractCreator)

	const numSwaps = 16
	type swapCase struct {
		index     int // index in the swap array
		walletKey *ecdsa.PrivateKey
		id        [32]byte
		secret    [32]byte
		swap      SwapCreatorSwap
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

		sc.secret = proof.Secret()
		sc.walletKey = tests.GetTestKeyByIndex(t, i)
		addrSwap := crypto.PubkeyToAddress(*sc.walletKey.Public().(*ecdsa.PublicKey))

		sc.swap = SwapCreatorSwap{
			Owner:        addrSwap,
			Claimer:      addrSwap,
			PubKeyClaim:  res.Secp256k1PublicKey().Keccak256(),
			PubKeyRefund: dummySwapKey, // no one calls refund in this test
			Timeout0:     nil,          // timeouts initialised when swap is created
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

			auth := getAuth(t, sc.walletKey)
			auth.Value = sc.swap.Value
			tx, err := swapCreator.NewSwap(
				auth,
				sc.swap.PubKeyClaim,
				sc.swap.PubKeyRefund,
				sc.swap.Claimer,
				defaultTimeoutDuration,
				defaultTimeoutDuration,
				types.EthAssetETH.Address(),
				sc.swap.Value,
				sc.swap.Nonce,
			)
			require.NoError(t, err)
			receipt := getReceipt(t, ec, tx)

			t.Logf("gas cost to call ETH NewSwap[%d]: %d (delta %d)",
				sc.index, receipt.GasUsed, MaxNewSwapETHGas-int(receipt.GasUsed))
			require.GreaterOrEqual(t, MaxNewSwapETHGas, int(receipt.GasUsed), "ETH NewSwap")

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
			tx, err := swapCreator.SetReady(getAuth(t, sc.walletKey), sc.swap)
			require.NoError(t, err)
			receipt := getReceipt(t, ec, tx)
			t.Logf("gas cost to call SetReady[%d]: %d (delta %d)",
				sc.index, receipt.GasUsed, MaxSetReadyGas-int(receipt.GasUsed))
			require.GreaterOrEqual(t, MaxSetReadyGas, int(receipt.GasUsed), "SetReady")
		}(&swapCases[i])
	}
	wg.Wait() // set_ready called on all swaps

	// call claim on all the swaps
	wg.Add(numSwaps)
	for i := 0; i < numSwaps; i++ {
		go func(sc *swapCase) {
			defer wg.Done()
			tx, err := swapCreator.Claim(getAuth(t, sc.walletKey), sc.swap, sc.secret)
			require.NoError(t, err)
			receipt := getReceipt(t, ec, tx)
			t.Logf("gas cost to call ETH Claim[%d]: %d (delta %d)",
				sc.index, receipt.GasUsed, MaxClaimETHGas-int(receipt.GasUsed))
			require.GreaterOrEqual(t, MaxClaimETHGas, int(receipt.GasUsed), "ETH Claim")
		}(&swapCases[i])
	}
	wg.Wait() // claim called on all swaps

	// ensure all swaps are completed
	wg.Add(numSwaps)
	for i := 0; i < numSwaps; i++ {
		go func(sc *swapCase) {
			defer wg.Done()
			stage, err := swapCreator.Swaps(nil, sc.id)
			require.NoError(t, err)
			require.Equal(t, StageToString(StageCompleted), StageToString(stage))
		}(&swapCases[i])
	}
	wg.Wait() // status of all swaps checked
}
