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
	// `./monero-wallet-rpc  --stagenet --rpc-bind-port 18080 --password "" --disable-rpc-login --wallet-file stagenet-wallet`
	const amount = 3000000000000
	cA := NewClient(defaultEndpoint)

	aliceAddress, err := cA.callGetAddress(0)
	require.NoError(t, err)
	t.Log("aliceAddress", aliceAddress)

	daemon := NewClient(defaultDaemonEndpoint)

	balance, err := cA.GetBalance(0)
	require.NoError(t, err)
	t.Log("balance: ", balance.Balance)
	t.Log("unlocked balance: ", balance.UnlockedBalance)
	t.Log("blocks to unlock: ", balance.BlocksToUnlock)

	if balance.BlocksToUnlock > 0 && balance.UnlockedBalance == 0 {
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
	// `./monero-wallet-rpc  --stagenet --rpc-bind-port 18082 --password "" --disable-rpc-login --wallet-dir .`
	// TODO: it seems the wallet CLI fails to generate from keys when wallet-dir is not set,
	// but it fails to load the wallet if wallet-file is not set (and these two flags cannot be used together)
	cB := NewClient("http://127.0.0.1:18084/json_rpc")

	// generate view-only account for A+B
	err = cB.callGenerateFromKeys(nil, vkABPriv, kpABPub.Address(), fmt.Sprintf("test-wallet-%d", r), "")
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

		err = daemon.callGenerateBlocks(aliceAddress.Address, 1)
		require.NoError(t, err)
		time.Sleep(time.Second)
	}

	err = daemon.callGenerateBlocks(aliceAddress.Address, 16)
	require.NoError(t, err)

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
