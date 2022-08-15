package monero

import (
	"sync"

	"github.com/MarinX/monerorpc"
	"github.com/MarinX/monerorpc/wallet"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
)

// Client represents a monero-wallet-rpc client.
type Client interface {
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

type client struct {
	sync.Mutex
	rpc *monerorpc.MoneroRPC
}

// NewClient returns a new monero-wallet-rpc client.
func NewClient(endpoint string) *client {
	return &client{
		rpc: monerorpc.New(endpoint, nil),
	}
}

func (c *client) LockClient() {
	c.Lock()
}

func (c *client) UnlockClient() {
	c.Unlock()
}

func (c *client) GetAccounts() (*wallet.GetAccountsResponse, error) {
	return c.callGetAccounts()
}

func (c *client) GetBalance(idx uint64) (*wallet.GetBalanceResponse, error) {
	return c.callGetBalance(idx)
}

func (c *client) Transfer(to mcrypto.Address, accountIdx, amount uint64) (*wallet.TransferResponse, error) {
	destination := wallet.Destination{
		Amount:  amount,
		Address: string(to),
	}

	return c.callTransfer([]wallet.Destination{destination}, accountIdx)
}

func (c *client) SweepAll(to mcrypto.Address, accountIdx uint64) (*wallet.SweepAllResponse, error) {
	return c.callSweepAll(string(to), accountIdx)
}

func (c *client) GenerateFromKeys(kp *mcrypto.PrivateKeyPair, filename, password string, env common.Environment) error {
	return c.callGenerateFromKeys(kp.SpendKey(), kp.ViewKey(), kp.Address(env), filename, password)
}

func (c *client) GenerateViewOnlyWalletFromKeys(vk *mcrypto.PrivateViewKey, address mcrypto.Address,
	filename, password string) error {
	return c.callGenerateFromKeys(nil, vk, address, filename, password)
}

func (c *client) GetAddress(idx uint64) (*wallet.GetAddressResponse, error) {
	return c.callGetAddress(idx)
}

func (c *client) Refresh() error {
	return c.refresh()
}

func (c *client) refresh() error {
	_, err := c.rpc.Wallet.Refresh(&wallet.RefreshRequest{})
	return err
}

func (c *client) CreateWallet(filename, password string) error {
	return c.callCreateWallet(filename, password)
}

func (c *client) OpenWallet(filename, password string) error {
	return c.callOpenWallet(filename, password)
}

func (c *client) CloseWallet() error {
	return c.rpc.Wallet.CloseWallet()
}

func (c *client) GetHeight() (uint64, error) {
	return c.callGetHeight()
}
