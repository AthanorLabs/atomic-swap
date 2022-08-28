package swapfactory

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/dleq"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
)

func TestSwapFactory_NewSwap_ERC20(t *testing.T) {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	// deploy ERC20Mock
	erc20Addr, erc20Tx, _, err := DeployERC20Mock(auth, conn, "ERC20Mock", "MOCK", addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy ERC20Mock.sol: %d", receipt.GasUsed)

	testNewSwap(t, erc20Addr)
}

func TestSwapFactory_Claim_ERC20(t *testing.T) {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	// deploy SwapFactory
	address, tx, contract, err := DeploySwapFactory(auth, conn)
	require.NoError(t, err)
	require.NotEqual(t, ethcommon.Address{}, address)
	require.NotNil(t, tx)
	require.NotNil(t, contract)
	receipt, err := block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy SwapFactory.sol: %d", receipt.GasUsed)

	// deploy ERC20Mock
	erc20Addr, erc20Tx, _, err := DeployERC20Mock(auth, conn, "ERC20Mock", "MOCK", addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy ERC20Mock.sol: %d", receipt.GasUsed)

	// generate claim secret and public key
	dleq := &dleq.CGODLEq{}
	proof, err := dleq.Prove()
	require.NoError(t, err)
	res, err := dleq.Verify(proof)
	require.NoError(t, err)

	// hash public key
	cmt := res.Secp256k1PublicKey().Keccak256()

	nonce := big.NewInt(0)
	tx, err = contract.NewSwap(auth, cmt, [32]byte{}, addr,
		defaultTimeoutDuration, erc20Addr, big.NewInt(0), nonce)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call new_swap: %d", receipt.GasUsed)

	// 3 logs:
	// Approval
	// Transfer
	// New
	//
	// The ERC20Mock contract auto-approves on a call to `transferFrom`,
	// obviously real ERC20s shouldn't have this behaviour
	require.Equal(t, 3, len(receipt.Logs))

	id, err := GetIDFromLog(receipt.Logs[2])
	require.NoError(t, err)

	t0, t1, err := GetTimeoutsFromLog(receipt.Logs[2])
	require.NoError(t, err)

	swap := SwapFactorySwap{
		Owner:        addr,
		Claimer:      addr,
		PubKeyClaim:  cmt,
		PubKeyRefund: [32]byte{},
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        erc20Addr,
		Value:        big.NewInt(0),
		Nonce:        nonce,
	}

	// set contract to Ready
	tx, err = contract.SetReady(auth, swap)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	t.Logf("gas cost to call SetReady: %d", receipt.GasUsed)
	require.NoError(t, err)

	// now let's try to claim
	var s [32]byte
	secret := proof.Secret()
	copy(s[:], common.Reverse(secret[:]))
	tx, err = contract.Claim(auth, swap, s)
	require.NoError(t, err)
	receipt, err = block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to call Claim: %d", receipt.GasUsed)

	stage, err := contract.Swaps(nil, id)
	require.NoError(t, err)
	require.Equal(t, StageCompleted, stage)
}
