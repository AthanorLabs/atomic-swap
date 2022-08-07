package monero

import (
	"fmt"
	"testing"
	"time"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func TestClient_Transfer(t *testing.T) {
	if testing.Short() {
		t.Skip() // TODO: this fails on CI with a "No wallet file" error at line 76
	}

	const amount = 2800000000
	cXMRMaker := NewClient(tests.CreateWalletRPCService(t))

	err := cXMRMaker.CreateWallet("test-wallet", "")
	require.NoError(t, err)

	xmrmakerAddr, err := cXMRMaker.callGetAddress(0)
	require.NoError(t, err)

	daemon := NewClient(common.DefaultMoneroDaemonEndpoint)
	_ = daemon.callGenerateBlocks(xmrmakerAddr.Address, 512)

	time.Sleep(time.Second * 10)

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

	cXMRTaker := NewClient(tests.CreateWalletRPCService(t))

	// generate view-only account for A+B
	walletFP := fmt.Sprintf("test-wallet-%s", time.Now().Format(common.TimeFmtNSecs))
	err = cXMRTaker.callGenerateFromKeys(nil, vkABPriv, kpABPub.Address(common.Mainnet), walletFP, "")
	require.NoError(t, err)
	err = cXMRTaker.OpenWallet(walletFP, "")
	require.NoError(t, err)

	// transfer to account A+B
	_, err = cXMRMaker.Transfer(kpABPub.Address(common.Mainnet), 0, amount)
	require.NoError(t, err)
	err = daemon.callGenerateBlocks(xmrmakerAddr.Address, 1)
	require.NoError(t, err)

	for {
		t.Log("checking balance...")
		balance, err = cXMRTaker.GetBalance(0)
		require.NoError(t, err)

		if balance.Balance > 0 {
			t.Log("balance of AB: ", balance.Balance)
			t.Log("unlocked balance of AB: ", balance.UnlockedBalance)
			break
		}

		_ = daemon.callGenerateBlocks(xmrmakerAddr.Address, 1)
		time.Sleep(time.Second)
	}

	err = daemon.callGenerateBlocks(xmrmakerAddr.Address, 16)
	require.NoError(t, err)

	// generate spend account for A+B
	skAKPriv := mcrypto.SumPrivateSpendKeys(kpA.SpendKey(), kpB.SpendKey())
	// ignore the error for now, as it can error with "Wallet already exists."
	_ = cXMRTaker.callGenerateFromKeys(skAKPriv, vkABPriv, kpABPub.Address(common.Mainnet), walletFP, "")

	err = cXMRTaker.refresh()
	require.NoError(t, err)

	balance, err = cXMRTaker.GetBalance(0)
	require.NoError(t, err)
	require.Greater(t, balance.Balance, float64(0))

	// transfer from account A+B back to XMRMaker's address
	_, err = cXMRTaker.Transfer(mcrypto.Address(xmrmakerAddr.Address), 0, 1)
	require.NoError(t, err)
}

func TestClient_CloseWallet(t *testing.T) {
	c := NewClient(tests.CreateWalletRPCService(t))
	err := c.CreateWallet("test-wallet", "")
	require.NoError(t, err)

	err = c.CloseWallet()
	require.NoError(t, err)

	err = c.OpenWallet("test-wallet", "")
	require.NoError(t, err)
}

func TestClient_GetAccounts(t *testing.T) {
	c := NewClient(tests.CreateWalletRPCService(t))
	err := c.CreateWallet("test-wallet", "")
	require.NoError(t, err)
	resp, err := c.GetAccounts()
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.SubaddressAccounts))
}

func TestClient_GetHeight(t *testing.T) {
	c := NewClient(tests.CreateWalletRPCService(t))
	err := c.CreateWallet("test-wallet", "")
	require.NoError(t, err)
	resp, err := c.GetHeight()
	require.NoError(t, err)
	require.NotEqual(t, 0, resp)
}
