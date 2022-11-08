package xmrmaker

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	rcommon "github.com/athanorlabs/go-relayer/common"
	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	"github.com/athanorlabs/go-relayer/relayer"
	rrpc "github.com/athanorlabs/go-relayer/rpc"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/dleq"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/tests"
)

var (
	defaultTestTimeoutDuration = big.NewInt(60 * 5)
	relayerCommission          = float64(0.01)
)

// runRelayer starts the relayer and returns the endpoint URL
func runRelayer(
	t *testing.T,
	ctx context.Context, //nolint:revive // context should be 1st param to function
	ec *ethclient.Client,
	forwarderAddress ethcommon.Address,
	sk *ecdsa.PrivateKey,
	chainID *big.Int,
) string {
	iforwarder, err := gsnforwarder.NewIForwarder(forwarderAddress, ec)
	require.NoError(t, err)
	fw := gsnforwarder.NewIForwarderWrapped(iforwarder)

	key := rcommon.NewKeyFromPrivateKey(sk)

	cfg := &relayer.Config{
		Ctx:                   ctx,
		EthClient:             ec,
		Forwarder:             fw,
		Key:                   key,
		ChainID:               chainID,
		NewForwardRequestFunc: gsnforwarder.NewIForwarderForwardRequest,
	}

	r, err := relayer.NewRelayer(cfg)
	require.NoError(t, err)

	server, err := rrpc.NewServer(&rrpc.Config{
		Ctx:     ctx,
		Address: "127.0.0.1:0",
		Relayer: r,
	})
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.Start()
		require.ErrorIs(t, err, context.Canceled)
	}()

	t.Cleanup(func() {
		// We could call server.Shutdown() but the context cancellation would beat us to
		// it. Just make sure the test hangs if the server didn't exit.
		wg.Wait()
	})
	return server.HttpURL()
}

func TestSwapState_ClaimRelayer_ERC20(t *testing.T) {
	initialBalance := big.NewInt(100000000000)

	sk := tests.GetMakerTestKey(t)
	conn, chainID := tests.NewEthClient(t)

	pub := sk.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	txOpts, err := bind.NewKeyedTransactorWithChainID(sk, chainID)
	require.NoError(t, err)

	_, tx, _, err := contracts.DeployERC20Mock(
		txOpts,
		conn,
		"Mock",
		"MOCK",
		addr,
		initialBalance,
	)
	require.NoError(t, err)
	contractAddr, err := bind.WaitDeployed(context.Background(), conn, tx)
	require.NoError(t, err)

	testSwapStateClaimRelayer(t, sk, types.EthAsset(contractAddr))
}

func TestSwapState_ClaimRelayer_ETH(t *testing.T) {
	sk := tests.GetMakerTestKey(t)
	testSwapStateClaimRelayer(t, sk, types.EthAssetETH)
}

func testSwapStateClaimRelayer(t *testing.T, sk *ecdsa.PrivateKey, asset types.EthAsset) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	relayerSk := tests.GetTestKeyByIndex(t, 1)
	require.NotEqual(t, sk, relayerSk)
	conn, chainID := tests.NewEthClient(t)

	txOpts, err := bind.NewKeyedTransactorWithChainID(sk, chainID)
	require.NoError(t, err)

	// generate claim secret and public key
	dleq := &dleq.DefaultDLEq{}
	proof, err := dleq.Prove()
	require.NoError(t, err)
	res, err := dleq.Verify(proof)
	require.NoError(t, err)

	// hash public key of claim secret
	cmt := res.Secp256k1PublicKey().Keccak256()

	pub := sk.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	// deploy forwarder
	forwarderAddress, tx, forwarderContract, err := gsnforwarder.DeployForwarder(txOpts, conn)
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(ctx, conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy Forwarder.sol: %d", receipt.GasUsed)

	tx, err = forwarderContract.RegisterDomainSeparator(txOpts, gsnforwarder.DefaultName, gsnforwarder.DefaultVersion)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(ctx, conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call RegisterDomainSeparator: %d", receipt.GasUsed)

	// start relayer
	relayerEndpoint := runRelayer(t, ctx, conn, forwarderAddress, relayerSk, chainID)

	// deploy swap contract with claim key hash
	contractAddr, tx, contract, err := contracts.DeploySwapFactory(txOpts, conn, forwarderAddress)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(ctx, conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapFactory.sol: %d", receipt.GasUsed)

	if asset != types.EthAssetETH {
		token, err := contracts.NewIERC20(asset.Address(), conn) //nolint:govet
		require.NoError(t, err)

		balance, err := token.BalanceOf(&bind.CallOpts{}, addr)
		require.NoError(t, err)

		tx, err = token.Approve(txOpts, contractAddr, balance)
		require.NoError(t, err)

		_, err = block.WaitForReceipt(ctx, conn, tx.Hash())
		require.NoError(t, err)
	}

	value := big.NewInt(100000000000)
	nonce := big.NewInt(0)
	txOpts.Value = value

	tx, err = contract.NewSwap(txOpts, cmt, [32]byte{}, addr,
		defaultTestTimeoutDuration, asset.Address(), value, nonce)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(ctx, conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call new_swap: %d", receipt.GasUsed)
	txOpts.Value = big.NewInt(0)

	logIndex := 0
	if asset != types.EthAssetETH {
		logIndex = 2
	}

	require.Equal(t, logIndex+1, len(receipt.Logs))
	id, err := contracts.GetIDFromLog(receipt.Logs[logIndex])
	require.NoError(t, err)

	t0, t1, err := contracts.GetTimeoutsFromLog(receipt.Logs[logIndex])
	require.NoError(t, err)

	swap := contracts.SwapFactorySwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  cmt,
		PubKeyRefund: [32]byte{},
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        asset.Address(),
		Value:        value,
		Nonce:        nonce,
	}

	// set contract to Ready
	tx, err = contract.SetReady(txOpts, swap)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(ctx, conn, tx.Hash())
	t.Logf("gas cost to call SetReady: %d", receipt.GasUsed)
	require.NoError(t, err)

	// now let's try to claim
	var s [32]byte
	secret := proof.Secret()
	copy(s[:], common.Reverse(secret[:]))

	txHash, err := claimRelayer(
		context.Background(),
		sk,
		contract,
		contractAddr,
		conn,
		relayerEndpoint,
		relayerCommission,
		&swap,
		s,
	)
	require.NoError(t, err)

	receipt, err = block.WaitForReceipt(context.Background(), conn, txHash)
	require.NoError(t, err)
	t.Logf("gas cost to call Claim via relayer: %d", receipt.GasUsed)

	if asset != types.EthAssetETH {
		require.Equal(t, 3, len(receipt.Logs))
	} else {
		// expected 1 Claimed log
		require.Equal(t, 1, len(receipt.Logs))
	}

	stage, err := contract.Swaps(nil, id)
	require.NoError(t, err)
	require.Equal(t, contracts.StageCompleted, stage)
}

func TestCalculateRelayerCommissionValue(t *testing.T) {
	swapValueF := big.NewFloat(0).Mul(big.NewFloat(4.567), numEtherUnitsFloat)
	swapValue, _ := swapValueF.Int(nil)

	relayerCommission := float64(0.01398)

	expectedF := big.NewFloat(0).Mul(big.NewFloat(0.06384666), numEtherUnitsFloat)
	expected, _ := expectedF.Int(nil)

	val, err := calculateRelayerCommissionValue(swapValue, relayerCommission)
	require.NoError(t, err)
	require.Equal(t, expected, val)
}
