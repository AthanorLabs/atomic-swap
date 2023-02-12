// Package monero provides client libraries for working with wallet files and interacting
// with a monero node. Management of monero-wallet-rpc daemon instances is fully
// encapsulated by these libraries.
package monero

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/MarinX/monerorpc"
	monerodaemon "github.com/MarinX/monerorpc/daemon"
	"github.com/MarinX/monerorpc/wallet"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
)

const (
	moneroWalletRPCLogPrefix = "[monero-wallet-rpc]: "

	// MinSpendConfirmations is the number of confirmations required on transaction
	// outputs before they can be spent again.
	MinSpendConfirmations = 10

	// SweepToSelfConfirmations is the number of confirmations that we wait for when
	// sweeping funds from an A+B wallet to our primary wallet.
	SweepToSelfConfirmations = 2
)

// WalletClient represents a monero-wallet-rpc client.
type WalletClient interface {
	GetAccounts() (*wallet.GetAccountsResponse, error)
	GetAddress(idx uint64) (*wallet.GetAddressResponse, error)
	PrimaryAddress() mcrypto.Address
	GetBalance(idx uint64) (*wallet.GetBalanceResponse, error)
	Transfer(
		ctx context.Context,
		to mcrypto.Address,
		accountIdx uint64,
		amount *coins.PiconeroAmount,
		numConfirmations uint64,
	) (*wallet.Transfer, error)
	SweepAll(
		ctx context.Context,
		to mcrypto.Address,
		accountIdx uint64,
		numConfirmations uint64,
	) ([]*wallet.Transfer, error)
	CreateWalletConf(walletNamePrefix string) *WalletClientConf
	WalletName() string
	GetHeight() (uint64, error)
	Endpoint() string // URL on which the wallet is accepting RPC requests
	Close()           // Close closes the client itself, including any open wallet
	CloseAndRemoveWallet()
}

// WalletClientConf wraps the configuration fields needed to call NewWalletClient
type WalletClientConf struct {
	Env                 common.Environment   // Required
	WalletFilePath      string               // Required, wallet created if it does not exist
	WalletPassword      string               // Optional, password used to open wallet or when creating a new wallet
	WalletPort          uint                 // Optional, zero means OS picks a random port
	MonerodNodes        []*common.MoneroNode // Optional, defaulted from environment if nil
	MoneroWalletRPCPath string               // optional, path to monero-rpc-binary
	LogPath             string               // optional, default is dir(WalletFilePath)/../monero-wallet-rpc.log
}

// Fill fills in the optional configuration values (Port, MonerodNodes, MoneroWalletRPCPath,
// and LogPath) if they are not set.
// Note: MonerodNodes is set to the first validated node.
func (conf *WalletClientConf) Fill() error {
	if conf.WalletFilePath == "" {
		panic("WalletFilePath is a required conf field") // should have been caught before we were invoked
	}

	var err error
	if conf.MoneroWalletRPCPath == "" {
		conf.MoneroWalletRPCPath, err = getMoneroWalletRPCBin()
		if err != nil {
			return err
		}
	}

	if len(conf.MonerodNodes) == 0 {
		conf.MonerodNodes = common.ConfigDefaultsForEnv(conf.Env).MoneroNodes
	}

	validatedNode, err := findWorkingNode(conf.Env, conf.MonerodNodes)
	if err != nil {
		return err
	}
	conf.MonerodNodes = []*common.MoneroNode{validatedNode}

	if conf.LogPath == "" {
		// default to the folder above the wallet
		conf.LogPath = path.Join(path.Dir(path.Dir(conf.WalletFilePath)), "monero-wallet-rpc.log")
	}

	if conf.WalletPort == 0 {
		conf.WalletPort, err = getFreeTCPPort()
		if err != nil {
			return err
		}
	}

	return nil
}

// waitForReceiptRequest wraps the input parameters for waitForReceipt
type waitForReceiptRequest struct {
	Ctx              context.Context
	TxID             string
	NumConfirmations uint64
	AccountIdx       uint64
}

type walletClient struct {
	wRPC       wallet.Wallet       // full monero-wallet-rpc API (larger than the WalletClient interface)
	dRPC       monerodaemon.Daemon // full monerod RPC API
	endpoint   string
	walletAddr mcrypto.Address
	conf       *WalletClientConf
	rpcProcess *os.Process // monero-wallet-rpc process that we create
}

// NewWalletClient returns a WalletClient for a newly created monero-wallet-rpc process.
func NewWalletClient(conf *WalletClientConf) (WalletClient, error) {
	if path.Dir(conf.WalletFilePath) == "." {
		return nil, errors.New("wallet file cannot be in the current working directory")
	}

	err := conf.Fill()
	if err != nil {
		return nil, err
	}

	walletExists, err := common.FileExists(conf.WalletFilePath)
	if err != nil {
		return nil, err
	}

	isNewWallet := !walletExists
	validatedNode := conf.MonerodNodes[0]

	proc, err := createWalletRPCService(
		conf.Env,
		conf.MoneroWalletRPCPath,
		conf.WalletPort,
		path.Dir(conf.WalletFilePath),
		conf.LogPath,
		validatedNode)
	if err != nil {
		return nil, err
	}

	c := NewThinWalletClient(validatedNode.Host, validatedNode.Port, conf.WalletPort).(*walletClient)
	c.rpcProcess = proc

	walletName := path.Base(conf.WalletFilePath)
	if isNewWallet {
		if err = c.CreateWallet(walletName, conf.WalletPassword); err != nil {
			c.Close()
			return nil, err
		}
		log.Infof("New Monero wallet %s created", conf.WalletFilePath)
	} else {
		err = c.wRPC.OpenWallet(&wallet.OpenWalletRequest{
			Filename: walletName,
			Password: conf.WalletPassword,
		})
		if err != nil {
			c.Close()
			return nil, err
		}
	}
	acctResp, err := c.GetAddress(0)
	if err != nil {
		c.Close()
		return nil, err
	}
	c.walletAddr = mcrypto.Address(acctResp.Address)
	c.conf = conf
	return c, nil
}

// NewThinWalletClient returns a WalletClient for an existing monero-wallet-rpc process.
func NewThinWalletClient(monerodHost string, monerodPort uint, walletPort uint) WalletClient {
	monerodEndpoint := fmt.Sprintf("http://%s:%d/json_rpc", monerodHost, monerodPort)
	walletEndpoint := fmt.Sprintf("http://127.0.0.1:%d/json_rpc", walletPort)
	return &walletClient{
		dRPC:     monerorpc.New(monerodEndpoint, nil).Daemon,
		wRPC:     monerorpc.New(walletEndpoint, nil).Wallet,
		endpoint: walletEndpoint,
	}
}

func (c *walletClient) WalletName() string {
	return path.Base(c.conf.WalletFilePath)
}

func (c *walletClient) GetAccounts() (*wallet.GetAccountsResponse, error) {
	return c.wRPC.GetAccounts(&wallet.GetAccountsRequest{})
}

func (c *walletClient) GetBalance(idx uint64) (*wallet.GetBalanceResponse, error) {
	if err := c.refresh(); err != nil {
		return nil, err
	}
	return c.wRPC.GetBalance(&wallet.GetBalanceRequest{
		AccountIndex: idx,
	})
}

// waitForReceipt waits for the passed monero transaction ID to receive numConfirmations
// and returns the transfer information. While this function will always wait for the
// transaction to leave the mem-pool even if zero confirmations are requested, it is the
// caller's responsibility to request enough confirmations that the returned transfer
// information will not be invalidated by a block reorg.
func (c *walletClient) waitForReceipt(req *waitForReceiptRequest) (*wallet.Transfer, error) {
	height, err := c.GetHeight()
	if err != nil {
		return nil, err
	}

	var transfer *wallet.Transfer

	for {
		// Wallet is already refreshed here, due to GetHeight above and WaitForBlocks below
		transferResp, err := c.wRPC.GetTransferByTxid(&wallet.GetTransferByTxidRequest{
			TxID:         req.TxID,
			AccountIndex: req.AccountIdx,
		})
		if err != nil {
			return nil, err
		}

		transfer = &transferResp.Transfer
		log.Infof("Received %d of %d confirmations of XMR TXID=%s (height=%d)",
			transfer.Confirmations,
			req.NumConfirmations,
			req.TxID,
			height)
		// wait for transaction be mined (height set) even if 0 confirmations requested
		if transfer.Height > 0 && transfer.Confirmations >= req.NumConfirmations {
			break
		}

		height, err = WaitForBlocks(req.Ctx, c, 1)
		if err != nil {
			return nil, err
		}
	}

	return transfer, nil
}

func (c *walletClient) Transfer(
	ctx context.Context,
	to mcrypto.Address,
	accountIdx uint64,
	amount *coins.PiconeroAmount,
	numConfirmations uint64,
) (*wallet.Transfer, error) {
	amt, err := amount.Uint64()
	if err != nil {
		return nil, err
	}
	amountStr := amount.AsMoneroString()
	log.Infof("Transferring %s XMR to %s", amountStr, to)
	reqResp, err := c.wRPC.Transfer(&wallet.TransferRequest{
		Destinations: []wallet.Destination{{
			Amount:  amt,
			Address: string(to),
		}},
		AccountIndex: accountIdx,
	})
	if err != nil {
		log.Warnf("Transfer of %s XMR failed: %s", amountStr, err)
		return nil, fmt.Errorf("transfer failed: %w", err)
	}
	log.Infof("Transfer of %s XMR initiated, TXID=%s", amountStr, reqResp.TxHash)
	transfer, err := c.waitForReceipt(&waitForReceiptRequest{
		Ctx:              ctx,
		TxID:             reqResp.TxHash,
		NumConfirmations: numConfirmations,
		AccountIdx:       accountIdx,
	})
	if err != nil {
		return nil, fmt.Errorf("monero TXID=%s receipt failure: %w", reqResp.TxHash, err)
	}
	log.Infof("Transfer TXID=%s succeeded with %d confirmations and fee %s XMR",
		transfer.TxID,
		transfer.Confirmations,
		coins.FmtPiconeroAmtAsXMR(transfer.Fee),
	)
	return transfer, nil
}

func (c *walletClient) SweepAll(
	ctx context.Context,
	to mcrypto.Address,
	accountIdx uint64,
	numConfirmations uint64,
) ([]*wallet.Transfer, error) {
	addrResp, err := c.GetAddress(accountIdx)
	if err != nil {
		return nil, fmt.Errorf("sweep operation failed to get address: %w", err)
	}
	from := addrResp.Address

	balance, err := c.GetBalance(accountIdx)
	if err != nil {
		return nil, fmt.Errorf("sweep operation failed to get balance: %w", err)
	}
	log.Infof("Starting sweep of %s XMR from %s to %s", coins.FmtPiconeroAmtAsXMR(balance.Balance), from, to)
	if balance.Balance == 0 {
		return nil, fmt.Errorf("sweep from %s failed, no balance to sweep", from)
	}
	if balance.BlocksToUnlock > 0 {
		log.Infof("Sweep operation waiting %d blocks for balance to fully unlock", balance.BlocksToUnlock)
		if _, err = WaitForBlocks(ctx, c, int(balance.BlocksToUnlock)); err != nil {
			return nil, fmt.Errorf("sweep operation failed waiting to unlock balance: %w", err)
		}
	}

	reqResp, err := c.wRPC.SweepAll(&wallet.SweepAllRequest{
		AccountIndex: accountIdx,
		Address:      string(to),
	})
	if err != nil {
		return nil, fmt.Errorf("sweep_all from %s failed: %w", from, err)
	}
	log.Infof("Sweep transaction started, TX IDs: %s", strings.Join(reqResp.TxHashList, ", "))

	var transfers []*wallet.Transfer
	for _, txID := range reqResp.TxHashList {
		receipt, err := c.waitForReceipt(&waitForReceiptRequest{
			Ctx:              ctx,
			TxID:             txID,
			NumConfirmations: numConfirmations,
			AccountIdx:       accountIdx,
		})
		if err != nil {
			return nil, fmt.Errorf("sweep of TXID=%s failed waiting for receipt: %w", txID, err)
		}
		log.Infof("Sweep transfer ID=%s of %s XMR (%s XMR fees) completed at height %d",
			txID,
			coins.FmtPiconeroAmtAsXMR(receipt.Amount),
			coins.FmtPiconeroAmtAsXMR(receipt.Fee),
			receipt.Height,
		)
		transfers = append(transfers, receipt)
	}

	return transfers, nil
}

func (c *walletClient) CreateWalletConf(walletNamePrefix string) *WalletClientConf {
	walletName := fmt.Sprintf("%s-%s", walletNamePrefix, time.Now().Format(common.TimeFmtNSecs))
	walletPath := path.Join(path.Dir(c.conf.WalletFilePath), walletName)
	conf := &WalletClientConf{
		Env:                 c.conf.Env,
		WalletFilePath:      walletPath,
		WalletPassword:      c.conf.WalletPassword,
		WalletPort:          0,
		MonerodNodes:        c.conf.MonerodNodes,
		MoneroWalletRPCPath: c.conf.MoneroWalletRPCPath,
		LogPath:             c.conf.LogPath,
	}
	return conf
}

func createWalletFromKeys(
	conf *WalletClientConf,
	walletRestoreHeight uint64,
	privateSpendKey *mcrypto.PrivateSpendKey, // nil for a view-only wallet
	privateViewKey *mcrypto.PrivateViewKey,
	address mcrypto.Address,
) (WalletClient, error) {
	if conf.WalletPort == 0 { // swap wallets need randomized ports, so we expect this to be zero
		var err error
		conf.WalletPort, err = getFreeTCPPort()
		if err != nil {
			return nil, err
		}
	}
	// should be a one item list, we use the same node that the primary wallet is using
	monerodNode := conf.MonerodNodes[0]

	proc, err := createWalletRPCService(
		conf.Env,
		conf.MoneroWalletRPCPath,
		conf.WalletPort,
		path.Dir(conf.WalletFilePath),
		conf.LogPath,
		monerodNode,
	)
	if err != nil {
		return nil, err
	}

	c := NewThinWalletClient(monerodNode.Host, monerodNode.Port, conf.WalletPort).(*walletClient)
	c.rpcProcess = proc
	c.conf = conf
	err = c.generateFromKeys(
		privateSpendKey, // nil for a view-only wallet
		privateViewKey,
		address,
		walletRestoreHeight,
		path.Base(conf.WalletFilePath),
		conf.WalletPassword,
	)
	if err != nil {
		c.Close()
		return nil, err
	}

	acctResp, err := c.GetAddress(0)
	if err != nil {
		c.Close()
		return nil, err
	}
	c.walletAddr = mcrypto.Address(acctResp.Address)
	if c.walletAddr != address {
		c.Close()
		return nil, fmt.Errorf("provided address %s does not match monero-wallet-rpc computed address %s",
			address, c.walletAddr)
	}

	bal, err := c.GetBalance(0)
	if err != nil {
		c.Close()
		return nil, err
	}

	log.Infof("Created wallet %s, balance is %s XMR (%d blocks to unlock), address is %s",
		c.WalletName(),
		coins.FmtPiconeroAmtAsXMR(bal.Balance),
		bal.BlocksToUnlock,
		c.PrimaryAddress(),
	)
	return c, nil
}

// CreateSpendWalletFromKeys creates a new monero-wallet-rpc process, wallet client and
// spend wallet for the passed private key pair (view key and spend key).
func CreateSpendWalletFromKeys(
	conf *WalletClientConf,
	privateKeyPair *mcrypto.PrivateKeyPair,
	restoreHeight uint64,
) (WalletClient, error) {
	privateViewKey := privateKeyPair.ViewKey()
	privateSpendKey := privateKeyPair.SpendKey()
	address := privateKeyPair.Address(conf.Env)
	return createWalletFromKeys(conf, restoreHeight, privateSpendKey, privateViewKey, address)
}

// CreateSpendWalletFromKeysAndAddress creates a new monero-wallet-rpc process, wallet client and
// spend wallet for the passed private key pair (view key and spend key).
func CreateSpendWalletFromKeysAndAddress(
	conf *WalletClientConf,
	privateKeyPair *mcrypto.PrivateKeyPair,
	address mcrypto.Address,
	restoreHeight uint64,
) (WalletClient, error) {
	privateViewKey := privateKeyPair.ViewKey()
	privateSpendKey := privateKeyPair.SpendKey()
	return createWalletFromKeys(conf, restoreHeight, privateSpendKey, privateViewKey, address)
}

// CreateViewOnlyWalletFromKeys creates a new monero-wallet-rpc process, wallet client and
// view-only wallet for the passed private view key and address.
func CreateViewOnlyWalletFromKeys(
	conf *WalletClientConf,
	privateViewKey *mcrypto.PrivateViewKey,
	address mcrypto.Address,
	restoreHeight uint64,
) (WalletClient, error) {
	return createWalletFromKeys(conf, restoreHeight, nil, privateViewKey, address)
}

func (c *walletClient) generateFromKeys(
	sk *mcrypto.PrivateSpendKey,
	vk *mcrypto.PrivateViewKey,
	address mcrypto.Address,
	restoreHeight uint64,
	filename,
	password string,
) error {
	const (
		successMessage         = "Wallet has been generated successfully."
		viewOnlySuccessMessage = "Watch-only wallet has been generated successfully."
	)

	spendKey := "" // not used when only generating a view key
	if sk != nil {
		spendKey = sk.Hex()
	}

	res, err := c.wRPC.GenerateFromKeys(&wallet.GenerateFromKeysRequest{
		Filename:      filename,
		Address:       string(address),
		RestoreHeight: restoreHeight,
		Viewkey:       vk.Hex(),
		Spendkey:      spendKey,
		Password:      password,
	})
	if err != nil {
		return err
	}

	expectedMessage := successMessage
	if spendKey == "" {
		expectedMessage = viewOnlySuccessMessage
	}
	if res.Info != expectedMessage {
		return fmt.Errorf("got unexpected Info string: %s", res.Info)
	}

	return nil
}

func (c *walletClient) GetAddress(idx uint64) (*wallet.GetAddressResponse, error) {
	return c.wRPC.GetAddress(&wallet.GetAddressRequest{
		AccountIndex: idx,
	})
}

func (c *walletClient) refresh() error {
	_, err := c.wRPC.Refresh(&wallet.RefreshRequest{})
	return err
}

func (c *walletClient) CreateWallet(filename, password string) error {
	return c.wRPC.CreateWallet(&wallet.CreateWalletRequest{
		Filename: filename,
		Password: password,
		Language: "English",
	})
}

func (c *walletClient) PrimaryAddress() mcrypto.Address {
	if c.walletAddr == "" {
		// Initialised in constructor function, so this shouldn't ever happen
		panic("primary wallet address was not initialised")
	}
	return c.walletAddr
}

func (c *walletClient) GetHeight() (uint64, error) {
	if err := c.refresh(); err != nil {
		return 0, err
	}

	res, err := c.wRPC.GetHeight()
	if err != nil {
		return 0, err
	}

	return res.Height, nil
}

// getChainHeight gets the blockchain height directly from the monero daemon instead
// of the wallet height.
func (c *walletClient) getChainHeight() (uint64, error) {
	res, err := c.dRPC.GetBlockCount()
	if err != nil {
		return 0, err
	}

	return res.Count, nil
}

func (c *walletClient) Endpoint() string {
	return c.endpoint
}

// Close kills the monero-wallet-rpc process closing the wallet. It is designed to only be
// called a single time from a single go process.
func (c *walletClient) Close() {
	if c.rpcProcess == nil {
		return // no monero-wallet-rpc instance was created
	}
	p := c.rpcProcess
	err := c.wRPC.StopWallet()
	if err != nil {
		log.Warnf("StopWallet errored: %s", err)
		err = p.Kill() // uses, SIG-TERM, which monero-wallet-rpc has a handler for
	}
	// If err is nil at this point, the process existed, and we block until the child
	// process exits. (Note: kill does not error when signaling an exited, but non-reaped
	// child.)
	if err == nil {
		_, _ = p.Wait()
	}

}

// CloseAndRemoveWallet kills the monero-wallet-rpc process and removes the wallet files. This
// should never be called on the user's primary wallet. It is for temporary swap wallets only.
// Call this function at most once from a single go process.
func (c *walletClient) CloseAndRemoveWallet() {
	c.Close()

	// Just log any file removal errors, as there is nothing useful the caller can do
	// with the errors
	if err := os.Remove(c.conf.WalletFilePath); err != nil {
		log.Errorf("Failed to remove wallet file %q: %s", c.conf.WalletFilePath, err)
	}
	if err := os.Remove(c.conf.WalletFilePath + ".keys"); err != nil {
		log.Errorf("Failed to remove wallet keys file %q: %s", c.conf.WalletFilePath, err)
	}
	if err := os.Remove(c.conf.WalletFilePath + ".address.txt"); err != nil {
		// .address.txt doesn't always exist, only log if it existed and we failed
		if !errors.Is(err, fs.ErrNotExist) {
			log.Errorf("Failed to remove wallet address file %q: %s", c.conf.WalletFilePath, err)
		}
	}

}

func findWorkingNode(env common.Environment, nodes []*common.MoneroNode) (*common.MoneroNode, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no monero nodes")
	}

	var err error
	for _, n := range nodes {
		err = validateMonerodNode(env, n)
		if err != nil {
			log.Warnf("Non-working node: %s", err)
			continue
		}
		return n, nil
	}

	// err is non-nil if we get here
	return nil, fmt.Errorf("failed to validate any monerod RPC node, last error: %w", err)
}

// validateMonerodNode validates the monerod node before we launch monero-wallet-rpc, as
// doing the pre-checks creates more obvious error messages and faster failure.
func validateMonerodNode(env common.Environment, node *common.MoneroNode) error {
	endpoint := fmt.Sprintf("http://%s:%d/json_rpc", node.Host, node.Port)
	daemonCli := monerorpc.New(endpoint, nil).Daemon

	info, err := daemonCli.GetInfo()
	if err != nil {
		return fmt.Errorf("could not validate monerod endpoint %s: %w", endpoint, err)
	}

	switch env {
	case common.Stagenet:
		if !info.Stagenet {
			return fmt.Errorf("monerod endpoint %s is not a stagenet node", endpoint)
		}
	case common.Mainnet:
		if !info.Mainnet {
			return fmt.Errorf("monerod endpoint %s is not a mainnet node", endpoint)
		}
	case common.Development:
		if info.NetType != "fakechain" {
			return fmt.Errorf("monerod endpoint %s should have a network type of \"fakechain\" in dev mode",
				endpoint)
		}
	default:
		panic("unhandled environment type")
	}

	if env != common.Development && info.Offline {
		return fmt.Errorf("monerod endpoint %s is offline", endpoint)
	}

	if !info.Synchronized {
		return fmt.Errorf("monerod endpoint %s is not synchronised", endpoint)
	}

	return nil
}

// createWalletRPCService starts a monero-wallet-rpc instance. Default values are assigned
// to the MonerodHost, MonerodPort, WalletPort and LogPath fields of the config if they
// are not already set.
func createWalletRPCService(
	env common.Environment,
	walletRPCBinPath string,
	walletPort uint,
	walletDir string,
	logFilePath string,
	moneroNode *common.MoneroNode,
) (*os.Process, error) {
	walletRPCBinArgs := getWalletRPCFlags(env, walletPort, walletDir, logFilePath, moneroNode)
	proc, err := launchMoneroWalletRPCChild(walletRPCBinPath, walletRPCBinArgs...)
	if err != nil {
		return nil, fmt.Errorf("%w, see %s for details", err, logFilePath)
	}

	return proc, nil
}

// getMoneroWalletRPCBin returns the monero-wallet-rpc binary. It first looks for
// "./monero-bin/monero-wallet-rpc". If not found, it then looks for "monero-wallet-rpc"
// in the user's path.
func getMoneroWalletRPCBin() (string, error) {
	execName := "monero-wallet-rpc"
	priorityPath := path.Join("monero-bin", execName)
	execPath, err := exec.LookPath(priorityPath)
	if err == nil {
		return execPath, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	// search for the executable in the user's PATH
	return exec.LookPath(execName)
}

// getSysProcAttr returns SysProcAttr values that will work on all platforms, but this
// function is overwritten on Linux and FreeBSD.
var getSysProcAttr = func() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}

func launchMoneroWalletRPCChild(walletRPCBin string, walletRPCBinArgs ...string) (*os.Process, error) {
	cmd := exec.Command(walletRPCBin, walletRPCBinArgs...)

	pRead, pWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	cmd.Stdout = pWrite
	cmd.Stderr = pWrite

	// Last entry wins if an environment variable is in the list multiple times.
	// We parse some output, so we want to force English. NO_COLOR=1 failed to
	// remove ansi colour escapes, but setting TERM=dumb succeeded.
	cmd.Env = append(os.Environ(), "LANG=C", "LC_ALL=C", "TERM=dumb")

	cmd.SysProcAttr = getSysProcAttr()

	err = cmd.Start()
	// The writing side of the pipe will remain open in the child process after we close it
	// here, and the reading side will get an EOF after the last writing side closes. We
	// need to close the parent writing side after starting the child, in order for the child
	// to inherit the pipe's file descriptor for Stdout/Stderr. We can't close it in a defer
	// statement, because we need the scanner below to get EOF if the child process exits
	// on error.
	_ = pWrite.Close()

	// Handle err from cmd.Start() above
	if err != nil {
		return nil, err
	}

	log.Debugf("Started monero-wallet-rpc with PID=%d", cmd.Process.Pid)

	scanner := bufio.NewScanner(pRead)
	started := false
	// Loop terminates when the child process exits or when we get the message that the RPC server started
	for scanner.Scan() {
		line := scanner.Text()
		// Skips the first 3 lines of boilerplate output, but logs the version after it
		if line != "This is the RPC monero wallet. It needs to connect to a monero" &&
			line != "daemon to work correctly." &&
			line != "" {
			log.Info(moneroWalletRPCLogPrefix, line)
		}
		if strings.HasSuffix(line, "Starting wallet RPC server") {
			started = true
			break
		}
	}

	if !started {
		_, _ = cmd.Process.Wait() // shouldn't block, process already exited
		return nil, errors.New("failed to start monero-wallet-rpc")
	}

	time.Sleep(200 * time.Millisecond) // additional start time

	// Drain additional output. We are not detaching monero-wallet-rpc so it will
	// die when we exit. This has the downside that logs are sent both to the
	// monero-wallet-rpc.log file and to standard output.
	go func() {
		for scanner.Scan() {
			// We could log here, but it's noisy and we have a separate log file with
			// full logs. A future version could parse the log messages and send some
			// filtered subset to swapd's logs.
		}
		log.Warnf("monero-wallet-rpc pid=%d exited", cmd.Process.Pid)
	}()

	return cmd.Process, nil
}

// getWalletRPCFlags returns the flags used when launching monero-wallet-rpc
func getWalletRPCFlags(
	env common.Environment,
	walletPort uint,
	walletDir string,
	logFilePath string,
	moneroNode *common.MoneroNode,
) []string {
	args := []string{
		"--rpc-bind-ip=127.0.0.1",
		fmt.Sprintf("--rpc-bind-port=%d", walletPort),
		"--disable-rpc-login", // TODO: Enable this?
		fmt.Sprintf("--wallet-dir=%s", walletDir),
		fmt.Sprintf("--log-file=%s", logFilePath),
		"--log-level=0",
		fmt.Sprintf("--daemon-host=%s", moneroNode.Host),
		fmt.Sprintf("--daemon-port=%d", moneroNode.Port),
	}

	switch env {
	case common.Development:
		// See https://github.com/monero-project/monero/issues/8600
		args = append(args, "--allow-mismatched-daemon-version")
	case common.Mainnet:
		// do nothing
	case common.Stagenet:
		args = append(args, "--stagenet")
	default:
		panic("unhandled monero environment type")
	}

	return args
}

// getFreeTCPPort returns an OS allocated and immediately freed port. There is nothing preventing
// something else on the system from using the port before the caller has a chance, but OS
// allocated ports are randomised to make the risk negligible.
func getFreeTCPPort() (uint, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer func() { _ = ln.Close() }()

	return uint(ln.Addr().(*net.TCPAddr).Port), nil
}
