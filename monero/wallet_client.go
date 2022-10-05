package monero

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/MarinX/monerorpc"
	"github.com/MarinX/monerorpc/wallet"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
)

// WalletClient represents a monero-wallet-rpc client.
type WalletClient interface {
	LockClient() // can't use Lock/Unlock due to name conflict
	UnlockClient()
	GetAccounts() (*wallet.GetAccountsResponse, error)
	GetAddress(idx uint64) (*wallet.GetAddressResponse, error)
	GetBalance(idx uint64) (*wallet.GetBalanceResponse, error)
	Transfer(to mcrypto.Address, accountIdx, amount uint64) (*wallet.TransferResponse, error)
	SweepAll(to mcrypto.Address, accountIdx uint64) (*wallet.SweepAllResponse, error)
	GenerateFromKeys(kp *mcrypto.PrivateKeyPair, filename, password string, env common.Environment) error
	GenerateViewOnlyWalletFromKeys(vk *mcrypto.PrivateViewKey, address mcrypto.Address, filename, password string) error
	GetHeight() (uint64, error)
	Refresh() error
	CreateWallet(filename, password string) error
	OpenWallet(filename, password string) error
	CloseWallet() error
	Endpoint() string // URL on which the wallet is accepting RPC requests
	Close()           // Close closes the client itself, including any wallets that were open
}

// WalletClientConf wraps the configuration fields needed to call NewWalletClient
type WalletClientConf struct {
	Env                 common.Environment // Required
	WalletFilePath      string             // Required, wallet created if it does not exist
	WalletPassword      string             // Optional, password used to open wallet or when creating a new wallet
	WalletPort          uint               // Optional, zero means OS picks a random port
	MonerodPort         uint               // optional, defaulted from Env if not set
	MonerodHost         string             // optional, defaults to 127.0.0.1
	MoneroWalletRPCPath string             // optional, path to monero-rpc-binary
	LogPath             string             // optional, default is dir(WalletFilePath)/../monero-wallet-rpc.log
}

type walletClient struct {
	mu         sync.Mutex
	rpc        wallet.Wallet // full API with slightly different method signatures
	endpoint   string
	rpcProcess *os.Process // monero-wallet-rpc process that we create
}

// NewWalletClient returns a WalletClient for a newly created monero-wallet-rpc process.
func NewWalletClient(conf *WalletClientConf) (WalletClient, error) {
	if conf.WalletFilePath == "" {
		panic("WalletFile is a required conf field") // should have been caught before we were invoked
	}
	if conf.WalletPort == 0 {
		var err error
		conf.WalletPort, err = getFreePort()
		if err != nil {
			return nil, err
		}
	}

	if path.Dir(conf.WalletFilePath) == "." {
		return nil, errors.New("wallet file can not be in the current working directory")
	}

	walletExists, err := common.FileExists(conf.WalletFilePath)
	if err != nil {
		return nil, err
	}
	isNewWallet := !walletExists

	endpoint, proc, err := createWalletRPCService(conf)
	if err != nil {
		return nil, err
	}

	c := NewThinWalletClient(endpoint).(*walletClient)
	c.rpcProcess = proc

	walletName := path.Base(conf.WalletFilePath)
	if isNewWallet {
		if err := c.CreateWallet(walletName, conf.WalletPassword); err != nil {
			c.Close()
			return nil, err
		}
		log.Infof("New Monero wallet %s created", conf.WalletFilePath)
	}
	if err := c.OpenWallet(walletName, conf.WalletPassword); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

// NewThinWalletClient returns a WalletClient for an existing monero-wallet-rpc process.
func NewThinWalletClient(endpoint string) WalletClient {
	return &walletClient{
		rpc:      monerorpc.New(endpoint, nil).Wallet,
		endpoint: endpoint,
	}
}

func (c *walletClient) LockClient() {
	c.mu.Lock()
}

func (c *walletClient) UnlockClient() {
	c.mu.Unlock()
}

func (c *walletClient) GetAccounts() (*wallet.GetAccountsResponse, error) {
	return c.rpc.GetAccounts(&wallet.GetAccountsRequest{})
}

func (c *walletClient) GetBalance(idx uint64) (*wallet.GetBalanceResponse, error) {
	return c.rpc.GetBalance(&wallet.GetBalanceRequest{
		AccountIndex: idx,
	})
}

func (c *walletClient) Transfer(to mcrypto.Address, accountIdx, amount uint64) (*wallet.TransferResponse, error) {
	return c.rpc.Transfer(&wallet.TransferRequest{
		Destinations: []wallet.Destination{{
			Amount:  amount,
			Address: string(to),
		}},
		AccountIndex: accountIdx,
		Priority:     0,
	})
}

func (c *walletClient) SweepAll(to mcrypto.Address, accountIdx uint64) (*wallet.SweepAllResponse, error) {
	return c.rpc.SweepAll(&wallet.SweepAllRequest{
		AccountIndex: accountIdx,
		Address:      string(to),
	})
}

// GenerateFromKeys creates a wallet from a given wallet address, view key, and optional spend key
func (c *walletClient) GenerateFromKeys(
	kp *mcrypto.PrivateKeyPair,
	filename, password string,
	env common.Environment,
) error {
	return c.generateFromKeys(kp.SpendKey(), kp.ViewKey(), kp.Address(env), filename, password)
}

// GenerateViewOnlyWalletFromKeys creates a view-only wallet from a given view key and address
func (c *walletClient) GenerateViewOnlyWalletFromKeys(
	vk *mcrypto.PrivateViewKey,
	address mcrypto.Address,
	filename,
	password string,
) error {
	return c.generateFromKeys(nil, vk, address, filename, password)
}

func (c *walletClient) generateFromKeys(
	sk *mcrypto.PrivateSpendKey,
	vk *mcrypto.PrivateViewKey,
	address mcrypto.Address,
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

	res, err := c.rpc.GenerateFromKeys(&wallet.GenerateFromKeysRequest{
		Filename: filename,
		Address:  string(address),
		Viewkey:  vk.Hex(),
		Spendkey: spendKey,
		Password: password,
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
	return c.rpc.GetAddress(&wallet.GetAddressRequest{
		AccountIndex: idx,
	})
}

func (c *walletClient) Refresh() error {
	_, err := c.rpc.Refresh(&wallet.RefreshRequest{})
	return err
}

func (c *walletClient) CreateWallet(filename, password string) error {
	return c.rpc.CreateWallet(&wallet.CreateWalletRequest{
		Filename: filename,
		Password: password,
		Language: "English",
	})
}

func (c *walletClient) OpenWallet(filename, password string) error {
	return c.rpc.OpenWallet(&wallet.OpenWalletRequest{
		Filename: filename,
		Password: password,
	})
}

func (c *walletClient) CloseWallet() error {
	return c.rpc.CloseWallet()
}

func (c *walletClient) GetHeight() (uint64, error) {
	res, err := c.rpc.GetHeight()
	if err != nil {
		return 0, err
	}
	return res.Height, nil
}

func (c *walletClient) Endpoint() string {
	return c.endpoint
}

func (c *walletClient) Close() {
	c.LockClient()
	defer c.UnlockClient()
	if c.rpcProcess != nil {
		p := c.rpcProcess
		c.rpcProcess = nil
		err := p.Kill()
		if err != nil {
			_, _ = p.Wait()
		}
	}
}

// validateMonerodConfig validates the monerod node before we launch monero-wallet-rpc, as
// doing the pre-checks creates more obvious error messages and faster failure.
func validateMonerodConfig(env common.Environment, monerodHost string, monerodPort uint) error {
	endpoint := fmt.Sprintf("http://%s:%d/json_rpc", monerodHost, monerodPort)
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

// createWalletRPCService starts a monero-wallet-rpc listening on a random port for tests. The json_rpc
// URL of the started service is returned.
func createWalletRPCService(conf *WalletClientConf) (string, *os.Process, error) {
	walletRPCBin := conf.MoneroWalletRPCPath
	if walletRPCBin == "" {
		var err error
		walletRPCBin, err = getMoneroWalletRPCBin()
		if err != nil {
			return "", nil, err
		}
	}

	if conf.MonerodHost == "" {
		conf.MonerodHost = "127.0.0.1"
	}
	if conf.MonerodPort == 0 {
		switch conf.Env {
		case common.Mainnet, common.Development:
			conf.MonerodPort = common.DefaultMoneroDaemonMainnetPort
		case common.Stagenet:
			conf.MonerodPort = common.DefaultMoneroDaemonStagenetPort
		default:
			panic("unhandled environment value")
		}
	}

	if err := validateMonerodConfig(conf.Env, conf.MonerodHost, conf.MonerodPort); err != nil {
		return "", nil, err
	}

	if conf.LogPath == "" {
		// default to the folder above the wallet
		conf.LogPath = path.Join(path.Dir(path.Dir(conf.WalletFilePath)), "monero-wallet-rpc.log")
	}

	walletRPCBinArgs := getWalletRPCFlags(conf)
	proc, err := launchMoneroWalletRPCChild(walletRPCBin, walletRPCBinArgs...)
	if err != nil {
		return "", nil, fmt.Errorf("%w, see %s for details", err, conf.LogPath)
	}

	return fmt.Sprintf("http://127.0.0.1:%d/json_rpc", conf.WalletPort), proc, nil
}

// getMoneroWalletRPCBin returns the monero-wallet-rpc binary. It first looks for
// "./monero-bin/monero-wallet-rpc". If not found, it then looks for "monero-wallet-rpc"
// in the user's path.
func getMoneroWalletRPCBin() (string, error) {
	execName := "monero-wallet-rpc"
	priorityPath := path.Join("monero-bin", execName)
	path, err := exec.LookPath(priorityPath)
	if err == nil {
		return path, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	// search for the executable in the user's PATH
	return exec.LookPath(execName)
}

func launchMoneroWalletRPCChild(walletRPCBin string, walletRPCBinArgs ...string) (*os.Process, error) {
	cmd := exec.Command(walletRPCBin, walletRPCBinArgs...)

	pRead, pWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	cmd.Stdout = pWrite
	cmd.Stderr = pWrite

	// Last entry wins if LANG/LC_ALL were already set. It would be nice to not
	// include the user's environment, but we'd have to narrow down the minimal
	// set. A directory named ".shared-ringdb" is created in the current working
	// directory if we don't add os.Environ().
	cmd.Env = append(os.Environ(), "LANG=C", "LC_ALL=C")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}

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

	scanner := bufio.NewScanner(pRead)
	started := false
	// Loop terminates when the child process exits or when we get the message that the RPC server started
	for scanner.Scan() {
		line := scanner.Text()
		// Skips the first 3 lines of boilerplate output, but logs the version after it
		if line != "This is the RPC monero wallet. It needs to connect to a monero" &&
			line != "daemon to work correctly." &&
			line != "" {
			log.Infof("monero-wallet-rpc: %s", line)
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

	// drain any additional output
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			log.Infof("monero-wallet-rpc: %s", line)
		}
	}()
	return cmd.Process, nil
}

// getWalletRPCFlags returns the flags used when launching monero-wallet-rpc
func getWalletRPCFlags(conf *WalletClientConf) []string {
	args := []string{
		"--rpc-bind-ip=127.0.0.1",
		fmt.Sprintf("--rpc-bind-port=%d", conf.WalletPort),
		"--disable-rpc-login", // TODO: Enable this?
		fmt.Sprintf("--wallet-dir=%s", path.Dir(conf.WalletFilePath)),
		// monero-wallet-rpc doesn't allow "--password=" syntax for empty passwords, so we use 2 args
		"--password", conf.WalletPassword,
		fmt.Sprintf("--log-file=%s", conf.LogPath),
		"--log-level=0",
	}

	switch conf.Env {
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
	// monero-wallet-rpc defaults --daemon-host to 127.0.0.1 if not set
	if conf.MonerodHost != "" {
		args = append(args, fmt.Sprintf("--daemon-host=%s", conf.MonerodHost))
	}
	// monero-wallet-rpc defaults --daemon-port to 18081 (38081 if --stagenet is passed)
	if conf.MonerodPort != 0 {
		args = append(args, fmt.Sprintf("--daemon-port=%d", conf.MonerodPort))
	}

	return args
}

// getFreePort returns an OS allocated and immediately freed port. There is nothing preventing
// something else on the system from using the port before the caller has a chance, but OS
// allocated ports are randomised to make the risk negligible.
func getFreePort() (uint, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	defer func() { _ = ln.Close() }()
	if err != nil {
		return 0, err
	}
	return uint(ln.Addr().(*net.TCPAddr).Port), nil
}
