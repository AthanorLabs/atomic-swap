package monero

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// curl -X POST http://127.0.0.1:18081/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"generateblocks","params":{ "wallet_address":"44GBHzv...","amount_of_blocks":1000000}}' -H 'Content-Type: application/json'

func TestClient_Transfer(t *testing.T) {
	// start RPC server with wallet w/ balance:
	//
	// `./monero-wallet-rpc  --stagenet --rpc-bind-port 18080 --password "" --disable-rpc-login --wallet-file stagenet-wallet`
	const amount = 33
	cA := NewClient("http://127.0.0.1:18080/json_rpc")

	balance, err := cA.GetBalance(0)
	require.NoError(t, err)
	t.Log("balance: ", balance.Balance)
	t.Log("unlocked balance: ", balance.UnlockedBalance)
	t.Log("blocks to unlock: ", balance.BlocksToUnlock)

	if balance.BlocksToUnlock > 0 {
		t.Fatal("need to wait for balance to unlock")
	}

	kpA, err := GenerateKeys()
	require.NoError(t, err)

	kpB, err := GenerateKeys()
	require.NoError(t, err)

	kpABPub := SumSpendAndViewKeys(kpA.PublicKeyPair(), kpB.PublicKeyPair())

	vkABPriv := SumPrivateViewKeys(kpA.vk, kpB.vk)

	r, err := rand.Int(rand.Reader, big.NewInt(999))
	require.NoError(t, err)

	// start RPC server with wallet-dir
	// `./monero-wallet-rpc  --stagenet --rpc-bind-port 18082 --password "" --disable-rpc-login --wallet-dir .`
	// TODO: it seems the wallet CLI fails to generate from keys when wallet-dir is not set,
	// but it fails to load the wallet if wallet-file is not set (and these two flags cannot be used together)
	cB := NewClient(defaultEndpoint)

	// generate view-only account for A+B
	err = cB.callGenerateFromKeys(nil, vkABPriv, kpABPub.Address(), fmt.Sprintf("test-wallet-%d", r), "")
	require.NoError(t, err)

	// transfer to account A+B
	err = cA.Transfer(kpABPub.Address(), 0, amount)
	require.NoError(t, err)

	for {
		time.Sleep(time.Second * 10)
		t.Log("checking balance...")
		balance, err = cB.GetBalance(0)
		require.NoError(t, err)

		if balance.Balance > 0 {
			t.Log("balance of AB: ", balance.Balance)
			t.Log("unlocked balance of AB: ", balance.UnlockedBalance)
			break
		}
	}

	// generate spend account for A+B
	skAKPriv := SumPrivateSpendKeys(kpA.sk, kpB.sk)
	err = cB.callGenerateFromKeys(skAKPriv, vkABPriv, kpABPub.Address(), fmt.Sprintf("test-wallet-%d", r), "")
	require.NoError(t, err)

	// transfer from account A+B back to original address
	err = cB.Transfer(kpABPub.Address(), 1, amount)
	require.NoError(t, err)
}
