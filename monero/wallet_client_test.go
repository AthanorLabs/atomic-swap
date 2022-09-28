package monero

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
)

func init() {
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// Project root is one directory up and monero-bin is below it. The only external program
	// any test in this package calls is monero-wallet-rpc, so we don't need any other
	// directories in our path.
	os.Setenv("PATH", path.Join(path.Dir(curDir), "monero-bin"))
}

func TestClient_Transfer(t *testing.T) {
	const amount = 2800000000
	cXMRMaker := CreateWalletClient(t)
	MineMinXMRBalance(t, cXMRMaker, amount)

	balance := GetBalance(t, cXMRMaker)
	t.Log("balance: ", balance.Balance)
	t.Log("unlocked balance: ", balance.UnlockedBalance)
	t.Log("blocks to unlock: ", balance.BlocksToUnlock)
	require.Greater(t, balance.UnlockedBalance, uint64(amount))

	kpA, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	kpB, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	kpABPub := mcrypto.SumSpendAndViewKeys(kpA.PublicKeyPair(), kpB.PublicKeyPair())
	vkABPriv := mcrypto.SumPrivateViewKeys(kpA.ViewKey(), kpB.ViewKey())

	cXMRTaker, err := NewWalletClient(&WalletClientConf{
		Env:            common.Development,
		WalletFilePath: path.Join(t.TempDir(), "wallet", "not-used"),
	})
	require.NoError(t, err)
	require.NoError(t, cXMRTaker.CloseWallet())

	// generate view-only account for A+B
	viewWalletName := "test-view-wallet"
	err = cXMRTaker.(*walletClient).generateFromKeys(nil, vkABPriv, kpABPub.Address(common.Mainnet), viewWalletName, "")
	require.NoError(t, err)
	err = cXMRTaker.OpenWallet(viewWalletName, "")
	require.NoError(t, err)

	// transfer to account A+B
	resp, err := cXMRMaker.Transfer(kpABPub.Address(common.Mainnet), 0, amount)
	require.NoError(t, err)
	t.Logf("Transfer resp: %#v", resp)
	_, err = WaitForBlocks(cXMRMaker, 1)
	require.NoError(t, err)

	// Something strange is happening below. On the first loop iteration, we are seeing a positive
	// Balance, a zero UnlockedBalance, but BlocksToUnlock is also zero. :| One the second loop,
	// BlocksToUnlock is above zero.
	for {
		t.Log("checking XMR Taker balance:")
		balance = GetBalance(t, cXMRTaker)
		t.Log("\tbalance of AB: ", balance.Balance)
		t.Log("\tunlocked balance of AB: ", balance.UnlockedBalance)
		t.Log("\tblocks to unlock AB: ", balance.BlocksToUnlock)
		if balance.UnlockedBalance > 0 {
			require.NoError(t, cXMRTaker.CloseWallet())
			break
		}
		time.Sleep(backgroundMineInterval)
	}

	// generate spend account for A+B
	spendWalletName := "test-spend-wallet"
	skAKPriv := mcrypto.SumPrivateSpendKeys(kpA.SpendKey(), kpB.SpendKey())
	err = cXMRTaker.(*walletClient).generateFromKeys(skAKPriv, vkABPriv, kpABPub.Address(common.Mainnet), spendWalletName, "") //nolint:lll
	require.NoError(t, err)

	balance = GetBalance(t, cXMRTaker)
	require.Greater(t, balance.UnlockedBalance, uint64(0))

	// transfer from account A+B back to XMRMaker's address
	xmrmakerAddr, err := cXMRTaker.GetAddress(0)
	require.NoError(t, err)
	_, err = cXMRTaker.Transfer(mcrypto.Address(xmrmakerAddr.Address), 0, 1)
	require.NoError(t, err)
}

func TestClient_CloseWallet(t *testing.T) {
	password := t.Name()
	c, err := NewWalletClient(&WalletClientConf{
		Env:            common.Development,
		WalletFilePath: path.Join(t.TempDir(), "wallet", "test-wallet"),
		WalletPassword: password,
	})
	require.NoError(t, err)
	defer c.Close()

	err = c.CloseWallet()
	require.NoError(t, err)

	err = c.OpenWallet("test-wallet", password)
	require.NoError(t, err)
}

func TestClient_GetAccounts(t *testing.T) {
	c, err := NewWalletClient(&WalletClientConf{
		Env:            common.Development,
		WalletFilePath: path.Join(t.TempDir(), "wallet", "test-wallet"),
	})
	require.NoError(t, err)
	defer c.Close()
	resp, err := c.GetAccounts()
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.SubaddressAccounts))
}

func TestClient_GetHeight(t *testing.T) {
	c, err := NewWalletClient(&WalletClientConf{
		Env:            common.Development,
		WalletFilePath: path.Join(t.TempDir(), "wallet", "test-wallet"),
	})
	require.NoError(t, err)
	defer c.Close()
	resp, err := c.GetHeight()
	require.NoError(t, err)
	require.NotEqual(t, 0, resp)
}

func TestCallGenerateFromKeys(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	r, err := rand.Int(rand.Reader, big.NewInt(999))
	require.NoError(t, err)

	c, err := NewWalletClient(&WalletClientConf{
		Env:            common.Development,
		WalletFilePath: path.Join(t.TempDir(), "wallet", "not-used"),
	})
	require.NoError(t, err)
	defer c.Close()

	addr, err := c.GetAddress(0)
	require.NoError(t, err)
	t.Logf("Address %s", addr.Address)

	// initial wallet automatically closed when a new wallet is opened
	err = c.(*walletClient).generateFromKeys(kp.SpendKey(), kp.ViewKey(), kp.Address(common.Mainnet),
		fmt.Sprintf("test-wallet-%d", r), "")
	require.NoError(t, err)

	addr, err = c.GetAddress(0)
	require.NoError(t, err)
	t.Logf("Address %s", addr.Address)
}

func Test_getMoneroWalletRPCBin(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(wd)
	os.Chdir("..")
	walletRPCPath, err := getMoneroWalletRPCBin()
	require.NoError(t, err)
	t.Log(walletRPCPath)
}
