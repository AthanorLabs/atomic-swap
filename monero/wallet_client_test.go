package monero

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func TestClient_Transfer(t *testing.T) {
	const amount = 2800000000
	cXMRMaker := NewWalletClient(tests.CreateWalletRPCService(t))

	err := cXMRMaker.CreateWallet("test-wallet", "")
	require.NoError(t, err)

	xmrmakerAddr, err := cXMRMaker.GetAddress(0)
	require.NoError(t, err)

	daemon := NewDaemonClient(common.DefaultMoneroDaemonEndpoint)
	err = daemon.GenerateBlocks(xmrmakerAddr.Address, 512)
	require.NoError(t, err)
	require.NoError(t, cXMRMaker.Refresh())

	balance, err := cXMRMaker.GetBalance(0)
	require.NoError(t, err)
	t.Log("balance: ", balance.Balance)
	t.Log("unlocked balance: ", balance.UnlockedBalance)
	t.Log("blocks to unlock: ", balance.BlocksToUnlock)

	if balance.UnlockedBalance < amount {
		t.Fatal("need to wait for balance to unlock")
	}

	kpA, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	kpB, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	kpABPub := mcrypto.SumSpendAndViewKeys(kpA.PublicKeyPair(), kpB.PublicKeyPair())
	vkABPriv := mcrypto.SumPrivateViewKeys(kpA.ViewKey(), kpB.ViewKey())

	cXMRTaker := NewWalletClient(tests.CreateWalletRPCService(t))

	// generate view-only account for A+B
	walletFP := fmt.Sprintf("test-wallet-%s", time.Now().Format(common.TimeFmtNSecs))
	err = cXMRTaker.generateFromKeys(nil, vkABPriv, kpABPub.Address(common.Mainnet), walletFP, "")
	require.NoError(t, err)
	err = cXMRTaker.OpenWallet(walletFP, "")
	require.NoError(t, err)

	// transfer to account A+B
	resp, err := cXMRMaker.Transfer(kpABPub.Address(common.Mainnet), 0, amount)
	require.NoError(t, err)
	t.Logf("Transfer resp: %#v", resp)
	err = daemon.GenerateBlocks(xmrmakerAddr.Address, 1)
	require.NoError(t, err)

	for {
		t.Log("checking balance...")
		require.NoError(t, cXMRTaker.Refresh())
		balance, err = cXMRTaker.GetBalance(0)
		require.NoError(t, err)

		if balance.Balance > 0 {
			t.Log("balance of AB: ", balance.Balance)
			t.Log("unlocked balance of AB: ", balance.UnlockedBalance)
			break
		}

		err = daemon.GenerateBlocks(xmrmakerAddr.Address, 1)
		require.NoError(t, err)
		require.NoError(t, cXMRMaker.Refresh())
	}

	err = daemon.GenerateBlocks(xmrmakerAddr.Address, 16)
	require.NoError(t, err)
	require.NoError(t, cXMRMaker.Refresh())

	// generate spend account for A+B
	skAKPriv := mcrypto.SumPrivateSpendKeys(kpA.SpendKey(), kpB.SpendKey())
	// ignore the error for now, as it can error with "Wallet already exists."
	_ = cXMRTaker.generateFromKeys(skAKPriv, vkABPriv, kpABPub.Address(common.Mainnet), walletFP, "")

	require.NoError(t, cXMRTaker.Refresh())

	balance, err = cXMRTaker.GetBalance(0)
	require.NoError(t, err)
	require.Greater(t, balance.Balance, uint64(0))

	// transfer from account A+B back to XMRMaker's address
	_, err = cXMRTaker.Transfer(mcrypto.Address(xmrmakerAddr.Address), 0, 1)
	require.NoError(t, err)
}

func TestClient_CloseWallet(t *testing.T) {
	c := NewWalletClient(tests.CreateWalletRPCService(t))
	err := c.CreateWallet("test-wallet", "")
	require.NoError(t, err)

	err = c.CloseWallet()
	require.NoError(t, err)

	err = c.OpenWallet("test-wallet", "")
	require.NoError(t, err)
}

func TestClient_GetAccounts(t *testing.T) {
	c := NewWalletClient(tests.CreateWalletRPCService(t))
	err := c.CreateWallet("test-wallet", "")
	require.NoError(t, err)
	resp, err := c.GetAccounts()
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.SubaddressAccounts))
}

func TestClient_GetHeight(t *testing.T) {
	c := NewWalletClient(tests.CreateWalletRPCService(t))
	err := c.CreateWallet("test-wallet", "")
	require.NoError(t, err)
	resp, err := c.GetHeight()
	require.NoError(t, err)
	require.NotEqual(t, 0, resp)
}

func TestCallGenerateFromKeys(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	r, err := rand.Int(rand.Reader, big.NewInt(999))
	require.NoError(t, err)

	c := NewWalletClient(tests.CreateWalletRPCService(t))
	err = c.generateFromKeys(kp.SpendKey(), kp.ViewKey(), kp.Address(common.Mainnet),
		fmt.Sprintf("test-wallet-%d", r), "")
	require.NoError(t, err)
}
