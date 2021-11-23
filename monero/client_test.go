package monero

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_Transfer(t *testing.T) {
	// start RPC server with wallet w/ balance:
	//
	// `./monero-wallet-rpc --rpc-bind-port 18083 --password "" --disable-rpc-login --wallet-file test-wallet`
	const amount = 2800000000
	cA := NewClient(defaultEndpointWalletFile)

	aliceAddress, err := cA.callGetAddress(0)
	require.NoError(t, err)
	t.Log("aliceAddress", aliceAddress)

	daemon := NewClient(defaultDaemonEndpoint)
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

	kpA, err := GenerateKeys()
	require.NoError(t, err)

	kpB, err := GenerateKeys()
	require.NoError(t, err)

	kpABPub := SumSpendAndViewKeys(kpA.PublicKeyPair(), kpB.PublicKeyPair())

	vkABPriv := SumPrivateViewKeys(kpA.vk, kpB.vk)

	r, err := rand.Int(rand.Reader, big.NewInt(10000))
	require.NoError(t, err)

	// start RPC server with wallet-dir
	// `./monero-wallet-rpc --rpc-bind-port 18084 --password "" --disable-rpc-login --wallet-dir .`
	// TODO: it seems the wallet CLI fails to generate from keys when wallet-dir is not set,
	// but it fails to load the wallet if wallet-file is not set (and these two flags cannot be used together)
	cB := NewClient(defaultEndpointWalletDir)

	// generate view-only account for A+B
	walletFP := fmt.Sprintf("test-wallet-%d", r)
	err = cB.callGenerateFromKeys(nil, vkABPriv, kpABPub.Address(), walletFP, "")
	require.NoError(t, err)

	// transfer to account A+B
	err = cA.Transfer(kpABPub.Address(), 0, amount)
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
	skAKPriv := SumPrivateSpendKeys(kpA.sk, kpB.sk)
	err = cB.callGenerateFromKeys(skAKPriv, vkABPriv, kpABPub.Address(), fmt.Sprintf("test-wallet-spaghet%d", r), "")
	require.NoError(t, err)

	err = cB.refresh()
	require.NoError(t, err)

	balance, err = cB.GetBalance(0)
	require.NoError(t, err)
	if balance.Balance == 0 {
		t.Fatal("no balance in account 0")
	}

	// transfer from account A+B back to Alice's address
	err = cB.Transfer(Address(aliceAddress.Address), 0, 1)
	require.NoError(t, err)
}
