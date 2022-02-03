package monero

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"

	"github.com/stretchr/testify/require"
)

func TestClient_Transfer(t *testing.T) {
	if testing.Short() {
		t.Skip() // TODO: this fails on CI with a "No wallet file" error at line 76
	}

	const amount = 2800000000
	cBob := NewClient(common.DefaultBobMoneroEndpoint)

	err := cBob.OpenWallet("test-wallet", "")
	require.NoError(t, err)

	bobAddr, err := cBob.callGetAddress(0)
	require.NoError(t, err)

	daemon := NewClient(common.DefaultMoneroDaemonEndpoint)
	_ = daemon.callGenerateBlocks(bobAddr.Address, 181)

	time.Sleep(time.Second * 10)

	balance, err := cBob.GetBalance(0)
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

	r, err := rand.Int(rand.Reader, big.NewInt(10000))
	require.NoError(t, err)

	cAlice := NewClient(common.DefaultAliceMoneroEndpoint)

	// generate view-only account for A+B
	walletFP := fmt.Sprintf("test-wallet-%d", r)
	err = cAlice.callGenerateFromKeys(nil, vkABPriv, kpABPub.Address(common.Mainnet), walletFP, "")
	require.NoError(t, err)
	err = cAlice.OpenWallet(walletFP, "")
	require.NoError(t, err)

	// transfer to account A+B
	_, err = cBob.Transfer(kpABPub.Address(common.Mainnet), 0, amount)
	require.NoError(t, err)
	err = daemon.callGenerateBlocks(bobAddr.Address, 1)
	require.NoError(t, err)

	for {
		t.Log("checking balance...")
		balance, err = cAlice.GetBalance(0)
		require.NoError(t, err)

		if balance.Balance > 0 {
			t.Log("balance of AB: ", balance.Balance)
			t.Log("unlocked balance of AB: ", balance.UnlockedBalance)
			break
		}

		_ = daemon.callGenerateBlocks(bobAddr.Address, 1)
		time.Sleep(time.Second)
	}

	_ = daemon.callGenerateBlocks(bobAddr.Address, 16)

	// generate spend account for A+B
	skAKPriv := mcrypto.SumPrivateSpendKeys(kpA.SpendKey(), kpB.SpendKey())
	// ignore the error for now, as it can error with "Wallet already exists."
	_ = cAlice.callGenerateFromKeys(skAKPriv, vkABPriv, kpABPub.Address(common.Mainnet),
		fmt.Sprintf("test-wallet-%d", r), "")

	err = cAlice.refresh()
	require.NoError(t, err)

	balance, err = cAlice.GetBalance(0)
	require.NoError(t, err)
	require.NotEqual(t, 0, balance.Balance)

	// transfer from account A+B back to Bob's address
	_, err = cAlice.Transfer(mcrypto.Address(bobAddr.Address), 0, 1)
	require.NoError(t, err)
}

func TestClient_CloseWallet(t *testing.T) {
	c := NewClient(common.DefaultBobMoneroEndpoint)
	err := c.OpenWallet("test-wallet", "")
	require.NoError(t, err)

	err = c.CloseWallet()
	require.NoError(t, err)

	err = c.OpenWallet("test-wallet", "")
	require.NoError(t, err)
}

func TestClient_GetAccounts(t *testing.T) {
	c := NewClient(common.DefaultBobMoneroEndpoint)
	resp, err := c.GetAccounts()
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.SubaddressAccounts))
}

func TestClient_GetHeight(t *testing.T) {
	c := NewClient(common.DefaultBobMoneroEndpoint)
	resp, err := c.GetHeight()
	require.NoError(t, err)
	require.NotEqual(t, 0, resp)
}
