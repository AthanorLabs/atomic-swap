// Package monero provides client libraries for working with wallet files and interacting
// with a monero node. Management of monero-wallet-rpc daemon instances is fully
// encapsulated by these libraries.
package monero

import (
	"bufio"
	"context"
	"errors"
	"fmt"
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
)

// WalletClient represents a monero-wallet-rpc client.
type WalletClient interface {
	GetAccounts() (*wallet.GetAccountsResponse, error)
	GetAddress(idx uint64) (*wallet.GetAddressResponse, error)
	PrimaryAddress() mcrypto.Address
	GetBalance(idx uint64) (*wallet.GetBalanceResponse, error)
	Transfer(to mcrypto.Address, accountIdx uint64, amount *coins.PiconeroAmount) (*wallet.TransferResponse, error)
	SweepAll(to mcrypto.Address, accountIdx uint64) (*wallet.SweepAllResponse, error)
	WaitForReceipt(req *WaitForReceiptRequest) (*wallet.Transfer, error)
	CreateABWalletConf() *WalletClientConf
	WalletName() string
	GetHeight() (uint64, error)
	GetChainHeight() (uint64, error)
	Refresh() error
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

// WaitForReceiptRequest wraps the input parameters for WaitForReceipt
type WaitForReceiptRequest struct {
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
	if conf.WalletFilePath == "" {
		panic("WalletFilePath is a required conf field") // should have been caught before we were invoked
	}

	if path.Dir(conf.WalletFilePath) == "." {
		return nil, errors.New("wallet file cannot be in the current working directory")
	}

	walletExists, err := common.FileExists(conf.WalletFilePath)
	if err != nil {
		return nil, err
	}
	isNewWallet := !walletExists

	if conf.MoneroWalletRPCPath == "" {
		conf.MoneroWalletRPCPath, err = getMoneroWalletRPCBin()
		if err != nil {
			return nil, err
		}
	}

	if len(conf.MonerodNodes) == 0 {
		conf.MonerodNodes = common.ConfigDefaultsForEnv(conf.Env).MoneroNodes
	}
	validatedNode, err := findWorkingNode(conf.Env, conf.MonerodNodes)
	if err != nil {
		return nil, err
	}
	conf.MonerodNodes = []*common.MoneroNode{validatedNode}

	if conf.LogPath == "" {
		// default to the folder above the wallet
		conf.LogPath = path.Join(path.Dir(path.Dir(conf.WalletFilePath)), "monero-wallet-rpc.log")
	}

	if conf.WalletPort == 0 {
		conf.WalletPort, err = getFreeTCPPort()
		if err != nil {
			return nil, err
		}
	}

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

	// TODO: Break off code from here into a separate function
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
	return c.wRPC.GetBalance(&wallet.GetBalanceRequest{
		AccountIndex: idx,
	})
}

// WaitForReceipt waits for the passed monero transaction ID to receive numConfirmations
// and returns the transfer information. While this function will always wait for the
// transaction to leave the mem-pool even if zero confirmations are requested, it is the
// caller's responsibility to request enough confirmations that the returned transfer
// information will not be invalidated by a block reorg.
func (c *walletClient) WaitForReceipt(req *WaitForReceiptRequest) (*wallet.Transfer, error) {
	height, err := c.GetHeight()
	if err != nil {
		return nil, err
	}

	var transfer *wallet.Transfer

	for {
		if err = c.Refresh(); err != nil {
			return nil, err
		}
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
	to mcrypto.Address,
	accountIdx uint64,
	amount *coins.PiconeroAmount,
) (*wallet.TransferResponse, error) {
	amt, err := amount.Uint64()
	if err != nil {
		return nil, err
	}
	return c.wRPC.Transfer(&wallet.TransferRequest{
		Destinations: []wallet.Destination{{
			Amount:  amt,
			Address: string(to),
		}},
		AccountIndex: accountIdx,
	})
}

func (c *walletClient) SweepAll(to mcrypto.Address, accountIdx uint64) (*wallet.SweepAllResponse, error) {
	return c.wRPC.SweepAll(&wallet.SweepAllRequest{
		AccountIndex: accountIdx,
		Address:      string(to),
	})
}

func (c *walletClient) CreateABWalletConf() *WalletClientConf {
	walletName := fmt.Sprintf("ab-swap-wallet-%s", time.Now().Format(common.TimeFmtNSecs))
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
		panic("addresses do not match")
	}

	return c, err
}

// CreateSpendWalletFromKeys creates a new monero-wallet-rpc process and wallet from a given wallet address,
// view key, and optional spend key
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

// CreateViewOnlyWalletFromKeys creates a view-only wallet from a given view key and address
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

func (c *walletClient) Refresh() error {
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
	res, err := c.wRPC.GetHeight()
	if err != nil {
		return 0, err
	}
	return res.Height, nil
}

// GetChainHeight gets the blockchain height directly from the monero daemon instead
// of the wallet height. Unlike the wallet method GetHeight, this method does not
// require a wallet to be open and is safe to call without grabbing the client mutex.
func (c *walletClient) GetChainHeight() (uint64, error) {
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
	if c.rpcProcess != nil {
		p := c.rpcProcess
		err := p.Kill()
		if err == nil {
			_, _ = p.Wait()
		}
	}
}

// CloseAndRemoveWallet kills the monero-wallet-rpc process and removes the wallet files. This
// should never be called on the user's primary wallet. It is for temporary swap wallets only.
// Call this function at most once from a single go process.
func (c *walletClient) CloseAndRemoveWallet() {
	c.Close()
	walletFiles := []string{
		c.conf.WalletFilePath,
		c.conf.WalletFilePath + ".keys",
		c.conf.WalletFilePath + ".address.txt",
	}
	// Just log any file removal errors, as there is nothing useful the caller can do
	// with the errors
	for _, file := range walletFiles {
		if err := os.Remove(file); err != nil {
			log.Errorf("Failed to remove wallet file %q: %s", file, err)
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
	defer func() { _ = ln.Close() }()
	if err != nil {
		return 0, err
	}
	return uint(ln.Addr().(*net.TCPAddr).Port), nil
}
