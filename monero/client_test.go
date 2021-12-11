package monero

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero/crypto"

	"github.com/stretchr/testify/require"
)

func TestClient_Transfer(t *testing.T) {
	const amount = 2800000000
	cA := NewClient(common.DefaultBobMoneroEndpoint)

	err := cA.OpenWallet("test-wallet", "")
	require.NoError(t, err)

	aliceAddress, err := cA.callGetAddress(0)
	require.NoError(t, err)
	t.Log("aliceAddress", aliceAddress)

	daemon := NewClient(common.DefaultMoneroDaemonEndpoint)
	_ = daemon.callGenerateBlocks(aliceAddress.Address, 181)

	time.Sleep(time.Second * 10)

	balance, err := cA.GetBalance(0)
	require.NoError(t, err)
	t.Log("balance: ", balance.Balance)
	t.Log("unlocked balance: ", balance.UnlockedBalance)
	t.Log("blocks to unlock: ", balance.BlocksToUnlock)

	if balance.UnlockedBalance < amount {
		t.Fatal("need to wait for balance to unlock")
	}

	kpA, err := crypto.GenerateKeys()
	require.NoError(t, err)

	kpB, err := crypto.GenerateKeys()
	require.NoError(t, err)

	kpABPub := crypto.SumSpendAndViewKeys(kpA.PublicKeyPair(), kpB.PublicKeyPair())

	vkABPriv := crypto.SumPrivateViewKeys(kpA.ViewKey(), kpB.ViewKey())

	r, err := rand.Int(rand.Reader, big.NewInt(10000))
	require.NoError(t, err)

	cB := NewClient(common.DefaultAliceMoneroEndpoint)

	// generate view-only account for A+B
	walletFP := fmt.Sprintf("test-wallet-%d", r)
	err = cB.callGenerateFromKeys(nil, vkABPriv, kpABPub.Address(common.Mainnet), walletFP, "")
	require.NoError(t, err)

	// transfer to account A+B
	_, err = cA.Transfer(kpABPub.Address(common.Mainnet), 0, amount)
	require.NoError(t, err)
	err = daemon.callGenerateBlocks(aliceAddress.Address, 1)
	require.NoError(t, err)

	for {
		t.Log("checking balance...")
		balance, err = cB.GetBalance(0)
		require.NoError(t, err)

		if balance.Balance > 0 {
			t.Log("balance of AB: ", balance.Balance)
			t.Log("unlocked balance of AB: ", balance.UnlockedBalance)
			break
		}

		_ = daemon.callGenerateBlocks(aliceAddress.Address, 1)
		time.Sleep(time.Second)
	}

	_ = daemon.callGenerateBlocks(aliceAddress.Address, 16)

	// generate spend account for A+B
	skAKPriv := crypto.SumPrivateSpendKeys(kpA.SpendKey(), kpB.SpendKey())
	// ignore the error for now, as it can error with "Wallet already exists."
	_ = cB.callGenerateFromKeys(skAKPriv, vkABPriv, kpABPub.Address(common.Mainnet),
		fmt.Sprintf("test-wallet-%d", r), "")

	err = cB.refresh()
	require.NoError(t, err)

	balance, err = cB.GetBalance(0)
	require.NoError(t, err)
	if balance.Balance == 0 {
		t.Fatal("no balance in account 0")
	}

	// transfer from account A+B back to Alice's address
	_, err = cB.Transfer(crypto.Address(aliceAddress.Address), 0, 1)
	require.NoError(t, err)
}
