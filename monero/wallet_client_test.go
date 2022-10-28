package monero

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
)

var moneroWalletRPCPath = path.Join("..", "monero-bin", "monero-wallet-rpc")

func TestClient_Transfer(t *testing.T) {
	amount := common.MoneroToPiconero(10) // 1k monero

	cXMRMaker := CreateWalletClient(t)
	MineMinXMRBalance(t, cXMRMaker, amount+common.MoneroToPiconero(0.1)) // add a little extra for fees

	balance := GetBalance(t, cXMRMaker)
	t.Logf("Bob's initial balance: bal=%d unlocked=%d blocks-to-unlock=%d",
		balance.Balance, balance.UnlockedBalance, balance.BlocksToUnlock)
	require.Greater(t, balance.UnlockedBalance, uint64(amount))

	kpA, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	kpB, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	abAddress := mcrypto.SumSpendAndViewKeys(kpA.PublicKeyPair(), kpB.PublicKeyPair()).Address(common.Mainnet)
	vkABPriv := mcrypto.SumPrivateViewKeys(kpA.ViewKey(), kpB.ViewKey())

	// Transfer from Bob's account to the Alice+Bob swap account
	transResp, err := cXMRMaker.Transfer(abAddress, 0, uint64(amount))
	require.NoError(t, err)
	t.Logf("Bob sent %f (+fee %f) XMR to A+B address with TX ID %s",
		common.MoneroAmount(transResp.Amount).AsMonero(), common.MoneroAmount(transResp.Fee).AsMonero(),
		transResp.TxHash)
	require.NoError(t, err)
	transfer, err := cXMRMaker.WaitForTransReceipt(&WaitForReceiptRequest{
		Ctx:              context.Background(),
		TxID:             transResp.TxHash,
		DestAddr:         abAddress,
		NumConfirmations: MinSpendConfirmations,
		AccountIdx:       0,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, transfer.Confirmations, uint64(MinSpendConfirmations))
	t.Logf("Bob's TX was mined at height %d with %d confirmations", transfer.Height, transfer.Confirmations)
	cXMRMaker.Close() // Done with bob, make sure no one uses him again

	// Establish Alice's primary wallet
	alicePrimaryWallet := "test-swap-wallet"
	cXMRTaker, err := NewWalletClient(&WalletClientConf{
		Env:                 common.Development,
		WalletFilePath:      path.Join(t.TempDir(), "wallet", alicePrimaryWallet),
		MoneroWalletRPCPath: moneroWalletRPCPath,
	})
	require.NoError(t, err)
	addrResp, err := cXMRTaker.GetAddress(0)
	require.NoError(t, err)
	alicePrimaryAddr := mcrypto.Address(addrResp.Address)

	// Alice generates a view-only wallet for A+B to confirm that Bob sent the funds
	viewWalletName := "test-view-wallet"
	err = cXMRTaker.GenerateViewOnlyWalletFromKeys(vkABPriv, abAddress, transfer.Height, viewWalletName, "")
	require.NoError(t, err)

	// Verify that generateFromKeys closed Alice's primary wallet and opened the new A+B
	// view wallet by checking the address of the current wallet
	addrResp, err = cXMRTaker.GetAddress(0)
	require.NoError(t, err)
	require.Equal(t, abAddress, mcrypto.Address(addrResp.Address))

	balance = GetBalance(t, cXMRTaker)
	height, err := cXMRTaker.GetHeight()
	require.NoError(t, err)
	t.Logf("A+B View-Only wallet balance: bal=%d unlocked=%d blocks-to-unlock=%d, cur-height=%d",
		balance.Balance, balance.UnlockedBalance, balance.BlocksToUnlock, height)
	require.Zero(t, balance.BlocksToUnlock)
	require.Equal(t, balance.UnlockedBalance, balance.Balance)
	require.Equal(t, balance.UnlockedBalance, uint64(amount))

	// At this point Alice has received the key from Bob to create an A+B spend wallet.
	// She'll now sweep the funds from the A+B spend wallet into her primary wallet.
	spendWalletName := "test-spend-wallet"
	// TODO: Can we convert View-only wallet into spend wallet if it is the same wallet?
	skAKPriv := mcrypto.SumPrivateSpendKeys(kpA.SpendKey(), kpB.SpendKey())
	err = cXMRTaker.(*walletClient).generateFromKeys(skAKPriv, vkABPriv, abAddress, transfer.Height, spendWalletName, "")
	require.NoError(t, err)

	balance = GetBalance(t, cXMRTaker)
	// Verify that the spend wallet, like the view-only wallet, has the exact amount expected in it
	require.Equal(t, balance.UnlockedBalance, uint64(amount))

	// Alice transfers from A+B spend wallet to her primary wallet's address
	sweepResp, err := cXMRTaker.SweepAll(alicePrimaryAddr, 0)
	require.NoError(t, err)
	t.Logf("%#v", sweepResp)
	require.Len(t, sweepResp.TxHashList, 1) // In our case, it should always be a single transaction
	sweepTxID := sweepResp.TxHashList[0]
	sweepAmount := sweepResp.AmountList[0]
	sweepFee := sweepResp.FeeList[0]

	t.Logf("Sweep of A+B wallet sent %d with fees %d to Alice's primary wallet",
		sweepAmount, sweepFee)
	require.Equal(t, uint64(amount), sweepAmount+sweepFee)

	transfer, err = cXMRTaker.WaitForTransReceipt(&WaitForReceiptRequest{
		Ctx:              context.Background(),
		TxID:             sweepTxID,
		DestAddr:         alicePrimaryAddr,
		NumConfirmations: 2,
		AccountIdx:       0,
	})
	require.NoError(t, err)
	require.Equal(t, sweepFee, transfer.Fee)
	require.Equal(t, sweepAmount, transfer.Amount)
	t.Logf("Alice's sweep transactions was mined at height %d with %d confirmations",
		transfer.Height, transfer.Confirmations)

	// Verify zero balance of A+B wallet after sweep
	balance = GetBalance(t, cXMRTaker)
	require.Equal(t, balance.Balance, uint64(0))

	// Switch Alice back to her primary wallet
	require.NoError(t, cXMRTaker.OpenWallet(alicePrimaryWallet, ""))

	balance = GetBalance(t, cXMRTaker)
	t.Logf("Alice's primary wallet after sweep: bal=%d unlocked=%d blocks-to-unlock=%d",
		balance.Balance, balance.UnlockedBalance, balance.BlocksToUnlock)
	require.Equal(t, balance.Balance, sweepAmount)
}

func Test_walletClient_SweepAll_nothingToSweepReturnsError(t *testing.T) {
	emptyWallet := CreateWalletClient(t)
	takerWallet := CreateWalletClient(t)

	addrResp, err := takerWallet.GetAddress(0)
	require.NoError(t, err)
	destAddr := mcrypto.Address(addrResp.Address)

	_, err = emptyWallet.SweepAll(destAddr, 0)
	require.ErrorContains(t, err, "No unlocked balance in the specified account")
}

func TestClient_CloseWallet(t *testing.T) {
	password := t.Name()
	c, err := NewWalletClient(&WalletClientConf{
		Env:                 common.Development,
		WalletFilePath:      path.Join(t.TempDir(), "wallet", "test-wallet"),
		WalletPassword:      password,
		MoneroWalletRPCPath: moneroWalletRPCPath,
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
		Env:                 common.Development,
		WalletFilePath:      path.Join(t.TempDir(), "wallet", "test-wallet"),
		MoneroWalletRPCPath: moneroWalletRPCPath,
	})
	require.NoError(t, err)
	defer c.Close()
	resp, err := c.GetAccounts()
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.SubaddressAccounts))
}

func TestClient_GetHeight(t *testing.T) {
	c, err := NewWalletClient(&WalletClientConf{
		Env:                 common.Development,
		WalletFilePath:      path.Join(t.TempDir(), "wallet", "test-wallet"),
		MoneroWalletRPCPath: moneroWalletRPCPath,
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

	c, err := NewWalletClient(&WalletClientConf{
		Env:                 common.Development,
		WalletFilePath:      path.Join(t.TempDir(), "wallet", "not-used"),
		MoneroWalletRPCPath: moneroWalletRPCPath,
	})
	require.NoError(t, err)
	defer c.Close()

	height, err := c.GetHeight()
	require.NoError(t, err)

	addr, err := c.GetAddress(0)
	require.NoError(t, err)
	t.Logf("Address %s", addr.Address)

	// initial wallet automatically closed when a new wallet is opened
	err = c.(*walletClient).generateFromKeys(
		kp.SpendKey(),
		kp.ViewKey(),
		kp.Address(common.Mainnet),
		height,
		"swap-deposit-wallet",
		"",
	)
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
	// monero-bin/monero-wallet-rpc should take precedence over any monero-wallet-rpc in
	// the user's path if the relative path to the binary exists.
	require.Equal(t, "monero-bin/monero-wallet-rpc", walletRPCPath)
}

func Test_validateMonerodConfig_devSuccess(t *testing.T) {
	err := validateMonerodConfig(common.Development, "127.0.0.1", common.DefaultMoneroDaemonDevPort)
	require.NoError(t, err)
}

func Test_validateMonerodConfig_stagenetSuccess(t *testing.T) {
	host := "node.sethforprivacy.com"
	err := validateMonerodConfig(common.Stagenet, host, 38089)
	require.NoError(t, err)
}

func Test_validateMonerodConfig_mainnetSuccess(t *testing.T) {
	host := "node.sethforprivacy.com"
	err := validateMonerodConfig(common.Mainnet, host, 18089)
	require.NoError(t, err)
}

func Test_validateMonerodConfig_misMatchedEnv(t *testing.T) {
	err := validateMonerodConfig(common.Mainnet, "127.0.0.1", common.DefaultMoneroDaemonDevPort)
	require.Error(t, err)
	require.Contains(t, err.Error(), "is not a mainnet node")
}

func Test_validateMonerodConfig_invalidPort(t *testing.T) {
	nonUsedPort, err := getFreePort()
	require.NoError(t, err)
	err = validateMonerodConfig(common.Development, "127.0.0.1", nonUsedPort)
	require.Error(t, err)
	require.Contains(t, err.Error(), "connection refused")
}

func Test_walletClient_waitForConfirmations_contextCancelled(t *testing.T) {
	amount := common.MoneroToPiconero(10) // 1k monero
	destAddr := mcrypto.Address(blockRewardAddress)

	c := CreateWalletClient(t)
	MineMinXMRBalance(t, c, amount+common.MoneroToPiconero(0.1)) // add a little extra for fees

	transResp, err := c.Transfer(destAddr, 0, uint64(amount))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = c.WaitForTransReceipt(&WaitForReceiptRequest{
		Ctx:              ctx,
		TxID:             transResp.TxHash,
		DestAddr:         destAddr,
		NumConfirmations: 999999999, // wait for a number of confirmations that would take a long time
		AccountIdx:       0,
	})
	require.ErrorIs(t, err, context.Canceled)
}
