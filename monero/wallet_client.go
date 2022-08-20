package monero

import (
	"fmt"
	"sync"

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
}

type walletClient struct {
	mu  sync.Mutex
	rpc wallet.Wallet // full API with slightly different method signatures
}

// NewWalletClient returns a new monero-wallet-rpc walletClient.
func NewWalletClient(endpoint string) *walletClient {
	return &walletClient{
		rpc: monerorpc.New(endpoint, nil).Wallet,
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
