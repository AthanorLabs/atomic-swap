package monero

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/cockroachdb/apd/v3"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
)

var moneroWalletRPCPath = path.Join("..", "monero-bin", "monero-wallet-rpc")

func init() {
	_ = logging.SetLogLevel("monero", "debug")
}

func TestClient_Transfer(t *testing.T) {
	amount := coins.MoneroToPiconero(apd.New(10, 0))
	amountPlusFees := coins.MoneroToPiconero(coins.StrToDecimal("10.01"))

	cXMRMaker := CreateWalletClient(t)
	MineMinXMRBalance(t, cXMRMaker, amountPlusFees)

	balance := GetBalance(t, cXMRMaker)
	t.Logf("Bob's initial balance: bal=%d unlocked=%d blocks-to-unlock=%d",
		balance.Balance, balance.UnlockedBalance, balance.BlocksToUnlock)

	amountU64, err := amount.Uint64()
	require.NoError(t, err)
	require.Greater(t, balance.UnlockedBalance, amountU64)

	kpA, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	kpB, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	abAddress := mcrypto.SumSpendAndViewKeys(kpA.PublicKeyPair(), kpB.PublicKeyPair()).Address(common.Mainnet)
	vkABPriv := mcrypto.SumPrivateViewKeys(kpA.ViewKey(), kpB.ViewKey())

	// Transfer from Bob's account to the Alice+Bob swap account
	transResp, err := cXMRMaker.Transfer(abAddress, 0, amount)
	require.NoError(t, err)
	t.Logf("Bob sent %s (+fee %s) XMR to A+B address with TX ID %s",
		coins.NewPiconeroAmount(transResp.Amount).AsMonero(),
		coins.NewPiconeroAmount(transResp.Fee).AsMonero(),
		transResp.TxHash)
	require.NoError(t, err)
	transfer, err := cXMRMaker.WaitForReceipt(&WaitForReceiptRequest{
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
	require.Equal(t, balance.UnlockedBalance, amountU64)

	// At this point Alice has received the key from Bob to create an A+B spend wallet.
	// She'll now sweep the funds from the A+B spend wallet into her primary wallet.
	spendWalletName := "test-spend-wallet"
	// TODO: Can we convert View-only wallet into spend wallet if it is the same wallet?
	skAKPriv := mcrypto.SumPrivateSpendKeys(kpA.SpendKey(), kpB.SpendKey())
	err = cXMRTaker.(*walletClient).generateFromKeys(skAKPriv, vkABPriv, abAddress, transfer.Height, spendWalletName, "")
	require.NoError(t, err)

	balance = GetBalance(t, cXMRTaker)
	// Verify that the spend wallet, like the view-only wallet, has the exact amount expected in it
	require.Equal(t, balance.UnlockedBalance, amountU64)

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
	require.Equal(t, amountU64, sweepAmount+sweepFee)

	transfer, err = cXMRTaker.WaitForReceipt(&WaitForReceiptRequest{
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

	require.NoError(t, c.Refresh())
	walletHeight, err := c.GetHeight()
	require.NoError(t, err)
	chainHeight, err := c.GetChainHeight()
	require.NoError(t, err)
	require.GreaterOrEqual(t, chainHeight, walletHeight)
	require.LessOrEqual(t, chainHeight-walletHeight, uint64(2))
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

func Test_validateMonerodConfigs_dev(t *testing.T) {
	env := common.Development
	node, err := findWorkingNode(env, common.ConfigDefaultsForEnv(env).MoneroNodes)
	require.NoError(t, err)
	require.NotNil(t, node)
}

func Test_validateMonerodConfigs_stagenet(t *testing.T) {
	env := common.Stagenet
	node, err := findWorkingNode(env, common.ConfigDefaultsForEnv(env).MoneroNodes)
	require.NoError(t, err)
	require.NotNil(t, node)
}

func Test_validateMonerodConfigs_mainnet(t *testing.T) {
	env := common.Mainnet
	node, err := findWorkingNode(env, common.ConfigDefaultsForEnv(env).MoneroNodes)
	require.NoError(t, err)
	require.NotNil(t, node)
}

func Test_validateMonerodConfig_misMatchedEnv(t *testing.T) {
	node := &common.MoneroNode{
		Host: "127.0.0.1",
		Port: common.DefaultMoneroDaemonDevPort,
	}
	err := validateMonerodNode(common.Mainnet, node)
	require.Error(t, err)
	require.Contains(t, err.Error(), "is not a mainnet node")
}

func Test_validateMonerodConfig_invalidPort(t *testing.T) {
	nonUsedPort, err := getFreeTCPPort()
	require.NoError(t, err)
	node := &common.MoneroNode{
		Host: "127.0.0.1",
		Port: nonUsedPort,
	}
	err = validateMonerodNode(common.Development, node)
	require.Error(t, err)
	require.Contains(t, err.Error(), "connection refused")
}

func Test_walletClient_waitForConfirmations_contextCancelled(t *testing.T) {
	const amount = 10
	minBal := coins.MoneroToPiconero(coins.StrToDecimal("10.01")) // add a little extra for fees
	destAddr := mcrypto.Address(blockRewardAddress)

	c := CreateWalletClient(t)
	MineMinXMRBalance(t, c, minBal)

	transResp, err := c.Transfer(destAddr, 0, coins.NewPiconeroAmount(amount))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = c.WaitForReceipt(&WaitForReceiptRequest{
		Ctx:              ctx,
		TxID:             transResp.TxHash,
		DestAddr:         destAddr,
		NumConfirmations: 999999999, // wait for a number of confirmations that would take a long time
		AccountIdx:       0,
	})
	require.ErrorIs(t, err, context.Canceled)
}
