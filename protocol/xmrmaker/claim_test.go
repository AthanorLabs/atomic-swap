package xmrmaker

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/dleq"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/relayer"
	"github.com/athanorlabs/atomic-swap/tests"
)

var (
	defaultTestTimeoutDuration = big.NewInt(60 * 5)
)

func TestSwapState_ClaimRelayer_ERC20(t *testing.T) {
	initialBalance := big.NewInt(90000000000000000)

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

	ec, chainID := tests.NewEthClient(t)
	extendedEC, err := extethclient.NewEthClient(ctx, common.Development, ec, sk)
	require.NoError(t, err)

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
	forwarderAddress, tx, forwarderContract, err := gsnforwarder.DeployForwarder(txOpts, ec)
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(ctx, ec, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy Forwarder.sol: %d", receipt.GasUsed)

	tx, err = forwarderContract.RegisterDomainSeparator(txOpts, gsnforwarder.DefaultName, gsnforwarder.DefaultVersion)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(ctx, ec, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call RegisterDomainSeparator: %d", receipt.GasUsed)

	// deploy swap contract with claim key hash
	contractAddr, tx, contract, err := contracts.DeploySwapFactory(txOpts, ec, forwarderAddress)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(ctx, ec, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapFactory.sol: %d", receipt.GasUsed)

	if asset != types.EthAssetETH {
		token, err := contracts.NewIERC20(asset.Address(), ec) //nolint:govet
		require.NoError(t, err)

		balance, err := token.BalanceOf(&bind.CallOpts{}, addr)
		require.NoError(t, err)

		tx, err = token.Approve(txOpts, contractAddr, balance)
		require.NoError(t, err)

		_, err = block.WaitForReceipt(ctx, ec, tx.Hash())
		require.NoError(t, err)
	}

	value := big.NewInt(90000000000000000)
	nonce := big.NewInt(0)
	txOpts.Value = value

	tx, err = contract.NewSwap(txOpts, cmt, [32]byte{}, addr,
		defaultTestTimeoutDuration, asset.Address(), value, nonce)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(ctx, ec, tx.Hash())
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

	swap := &contracts.SwapFactorySwap{
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
	tx, err = contract.SetReady(txOpts, *swap)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(ctx, ec, tx.Hash())
	t.Logf("gas cost to call SetReady: %d", receipt.GasUsed)
	require.NoError(t, err)

	secret := proof.Secret()

	// now let's try to claim
	req, err := relayer.CreateRelayClaimRequest(
		ctx,
		sk,
		ec,
		relayer.DefaultRelayerFee,
		contractAddr,
		forwarderAddress,
		swap,
		&secret,
	)
	require.NoError(t, err)

	resp, err := relayer.ValidateAndSendTransaction(ctx, req, extendedEC, forwarderAddress)
	require.NoError(t, err)

	receipt, err = block.WaitForReceipt(ctx, ec, resp.TxHash)
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
