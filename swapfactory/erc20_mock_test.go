package swapfactory

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

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

	erc20Addr, erc20Tx, _, err := DeployERC20Mock(auth, conn, "ERC20Mock", "MOCK", addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy ERC20Mock.sol: %d", receipt.GasUsed)

	// 3 logs:
	// Approval
	// Transfer
	// New
	//
	// The ERC20Mock contract auto-approves on a call to `transferFrom`,
	// obviously real ERC20s shouldn't have this behaviour
	testClaim(t, erc20Addr, 2)
}
