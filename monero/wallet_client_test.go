package monero

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/assert"
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	transferAmt := coins.MoneroToPiconero(coins.StrToDecimal("10"))
	transferAmtPlusFees := coins.MoneroToPiconero(coins.StrToDecimal("10.01"))

	cXMRMaker := CreateWalletClient(t)
	MineMinXMRBalance(t, cXMRMaker, transferAmtPlusFees)

	balanceBob := GetBalance(t, cXMRMaker)
	t.Logf("Bob's initial balance: bal=%s XMR, unlocked=%s XMR, blocks-to-unlock=%d",
		coins.FmtPiconeroAmtAsXMR(balanceBob.Balance),
		coins.FmtPiconeroAmtAsXMR(balanceBob.UnlockedBalance),
		balanceBob.BlocksToUnlock)

	transferAmtU64, err := transferAmt.Uint64()
	require.NoError(t, err)
	require.Greater(t, balanceBob.UnlockedBalance, transferAmtU64)

	kpA, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	kpB, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	abAddress := mcrypto.SumSpendAndViewKeys(kpA.PublicKeyPair(), kpB.PublicKeyPair()).Address(common.Development)
	vkABPriv := mcrypto.SumPrivateViewKeys(kpA.ViewKey(), kpB.ViewKey())

	// Transfer from Bob's account to the Alice+Bob swap account
	transfer, err := cXMRMaker.Transfer(ctx, abAddress, 0, transferAmt, MinSpendConfirmations)
	require.NoError(t, err)
	t.Logf("Bob sent %s (+fee %s) XMR to A+B address with TX ID %s",
		coins.FmtPiconeroAmtAsXMR(transfer.Amount),
		coins.FmtPiconeroAmtAsXMR(transfer.Fee),
		transfer.TxID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, transfer.Confirmations, uint64(MinSpendConfirmations))
	t.Logf("Bob's TX was mined at height %d with %d confirmations", transfer.Height, transfer.Confirmations)
	cXMRMaker.Close() // Done with bob, make sure no one uses him again
	cXMRMaker = nil

	// Establish Alice's primary wallet
	alicePrimaryWallet := "test-swap-wallet"
	cXMRTaker, err := NewWalletClient(&WalletClientConf{
		Env:                 common.Development,
		WalletFilePath:      path.Join(t.TempDir(), "wallet", alicePrimaryWallet),
		MoneroWalletRPCPath: moneroWalletRPCPath,
	})
	require.NoError(t, err)
	alicePrimaryAddr := cXMRTaker.PrimaryAddress()

	// Alice generates a view-only wallet for A+B to confirm that Bob sent the funds
	conf := cXMRTaker.CreateABWalletConf("alice-view-wallet-to-verify-funds")
	abViewCli, err := CreateViewOnlyWalletFromKeys(conf, vkABPriv, abAddress, transfer.Height)
	require.NoError(t, err)
	defer abViewCli.CloseAndRemoveWallet()

	balanceABWal := GetBalance(t, abViewCli)
	height, err := abViewCli.GetHeight()
	require.NoError(t, err)
	t.Logf("A+B View-Only wallet balance: bal=%s unlocked=%s blocks-to-unlock=%d, cur-height=%d",
		coins.FmtPiconeroAmtAsXMR(balanceABWal.Balance),
		coins.FmtPiconeroAmtAsXMR(balanceABWal.UnlockedBalance),
		balanceABWal.BlocksToUnlock, height)
	require.Zero(t, balanceABWal.BlocksToUnlock)
	require.Equal(t, balanceABWal.Balance, balanceABWal.UnlockedBalance)
	require.Equal(t, transferAmtU64, balanceABWal.UnlockedBalance)

	// At this point Alice has received the key from Bob to create an A+B spend wallet.
	// She'll now sweep the funds from the A+B spend wallet into her primary wallet.
	abWalletKeyPair := mcrypto.NewPrivateKeyPair(
		mcrypto.SumPrivateSpendKeys(kpA.SpendKey(), kpB.SpendKey()),
		mcrypto.SumPrivateViewKeys(kpA.ViewKey(), kpB.ViewKey()),
	)
	require.NoError(t, err)
	conf = abViewCli.CreateABWalletConf("alice-spend-wallet-to-claim")
	abSpendCli, err := CreateSpendWalletFromKeys(conf, abWalletKeyPair, transfer.Height)
	require.NoError(t, err)
	defer abSpendCli.CloseAndRemoveWallet()
	require.Equal(t, abSpendCli.PrimaryAddress(), abViewCli.PrimaryAddress())

	balanceABWal = GetBalance(t, abSpendCli)
	// Verify that the spend wallet, like the view-only wallet, has the exact amount expected in it
	require.Equal(t, transferAmtU64, balanceABWal.UnlockedBalance)

	// Alice transfers from A+B spend wallet to her primary wallet's address
	transfers, err := abSpendCli.SweepAll(ctx, alicePrimaryAddr, 0, 2)
	require.NoError(t, err)
	t.Logf("Alice swept AB wallet funds with %d transfers", len(transfers))
	require.Len(t, transfers, 1) // In our case, it should always be a single transaction
	sweepAmount := transfers[0].Amount
	sweepFee := transfers[0].Fee

	t.Logf("Sweep of A+B wallet sent %s XMR with fees %s XMR to Alice's primary wallet",
		coins.FmtPiconeroAmtAsXMR(sweepAmount), coins.FmtPiconeroAmtAsXMR(sweepFee))
	require.Equal(t, transferAmtU64, sweepAmount+sweepFee)
	t.Logf("Alice's sweep transactions was mined at height %d with %d confirmations",
		transfer.Height, transfer.Confirmations)

	// Verify zero balance of A+B wallet after sweep
	balanceABWal = GetBalance(t, abSpendCli)
	require.Equal(t, balanceABWal.Balance, uint64(0))

	balanceAlice := GetBalance(t, cXMRTaker)
	t.Logf("Alice's primary wallet after sweep: bal=%s XMR, unlocked=%s XMR, blocks-to-unlock=%d",
		coins.FmtPiconeroAmtAsXMR(balanceAlice.Balance),
		coins.FmtPiconeroAmtAsXMR(balanceAlice.UnlockedBalance),
		balanceAlice.BlocksToUnlock)
	require.Equal(t, balanceAlice.Balance, sweepAmount)
}

func Test_walletClient_SweepAll_nothingToSweepReturnsError(t *testing.T) {
	emptyWallet := CreateWalletClient(t)
	takerWallet := CreateWalletClient(t)

	addrResp, err := takerWallet.GetAddress(0)
	require.NoError(t, err)
	destAddr := mcrypto.Address(addrResp.Address)

	_, err = emptyWallet.SweepAll(context.Background(), destAddr, 0, 1)
	require.ErrorContains(t, err, "no balance to sweep")
}

func TestClient_CloseAndRemoveWallet(t *testing.T) {
	password := t.Name()
	walletPath := path.Join(t.TempDir(), "wallet", "test-wallet")
	c, err := NewWalletClient(&WalletClientConf{
		Env:                 common.Development,
		WalletFilePath:      walletPath,
		WalletPassword:      password,
		MoneroWalletRPCPath: moneroWalletRPCPath,
	})
	require.NoError(t, err)
	info, err := os.Stat(walletPath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
	c.CloseAndRemoveWallet()
	_, err = os.Stat(walletPath)
	require.ErrorIs(t, err, os.ErrNotExist)
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
	const numConfirmations = 999999999 // won't be achieved before our context is cancelled

	minBal := coins.MoneroToPiconero(coins.StrToDecimal("10.01")) // add a little extra for fees
	destAddr := mcrypto.Address(blockRewardAddress)

	c := CreateWalletClient(t)
	MineMinXMRBalance(t, c, minBal)

	// Cancel the context after the transfer has started
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.Transfer(ctx, destAddr, 0, coins.NewPiconeroAmount(amount), numConfirmations)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestCreateWalletFromKeys(t *testing.T) {
	// First wallet is just to get the height and generate the config for the 2nd
	primaryCli, err := NewWalletClient(&WalletClientConf{
		Env:                 common.Development,
		WalletFilePath:      path.Join(t.TempDir(), "wallet", "not-used"),
		MoneroWalletRPCPath: moneroWalletRPCPath,
	})
	require.NoError(t, err)
	defer primaryCli.Close()

	height, err := primaryCli.GetHeight()
	require.NoError(t, err)

	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	conf := primaryCli.CreateABWalletConf("ab-wallet")
	abCli, err := CreateSpendWalletFromKeys(conf, kp, height)
	require.NoError(t, err)
	defer abCli.CloseAndRemoveWallet()
	require.Equal(t, kp.Address(common.Development), abCli.PrimaryAddress())
}
