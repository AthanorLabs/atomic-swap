package monero

import (
	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero/crypto"
	"github.com/noot/atomic-swap/rpcclient"
)

// Client represents a monero-wallet-rpc client.
type Client interface {
	GetAccounts() (*getAccountsResponse, error)
	GetAddress(idx uint) (*getAddressResponse, error)
	GetBalance(idx uint) (*GetBalanceResponse, error)
	Transfer(to crypto.Address, accountIdx, amount uint) (*TransferResponse, error)
	GenerateFromKeys(kp *crypto.PrivateKeyPair, filename, password string, env common.Environment) error
	GenerateViewOnlyWalletFromKeys(vk *crypto.PrivateViewKey, address crypto.Address, filename, password string) error
	GetHeight() (uint, error)
	Refresh() error
	OpenWallet(filename, password string) error
	CloseWallet() error
}

type client struct {
	endpoint string
}

// NewClient returns a new monero-wallet-rpc client.
func NewClient(endpoint string) *client { //nolint:revive
	return &client{
		endpoint: endpoint,
	}
}

func (c *client) GetAccounts() (*getAccountsResponse, error) {
	return c.callGetAccounts()
}

func (c *client) GetBalance(idx uint) (*GetBalanceResponse, error) {
	return c.callGetBalance(idx)
}

func (c *client) Transfer(to crypto.Address, accountIdx, amount uint) (*TransferResponse, error) {
	destination := Destination{
		Amount:  amount,
		Address: string(to),
	}

	return c.callTransfer([]Destination{destination}, accountIdx)
}

func (c *client) GenerateFromKeys(kp *crypto.PrivateKeyPair, filename, password string, env common.Environment) error {
	return c.callGenerateFromKeys(kp.SpendKey(), kp.ViewKey(), kp.Address(env), filename, password)
}

func (c *client) GenerateViewOnlyWalletFromKeys(vk *crypto.PrivateViewKey, address crypto.Address,
	filename, password string) error {
	return c.callGenerateFromKeys(nil, vk, address, filename, password)
}

func (c *client) GetAddress(idx uint) (*getAddressResponse, error) {
	return c.callGetAddress(idx)
}

func (c *client) Refresh() error {
	return c.refresh()
}

func (c *client) refresh() error {
	const method = "refresh"

	resp, err := rpcclient.PostRPC(c.endpoint, method, "{}")
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

func (c *client) OpenWallet(filename, password string) error {
	return c.callOpenWallet(filename, password)
}

func (c *client) CloseWallet() error {
	const method = "close_wallet"

	resp, err := rpcclient.PostRPC(c.endpoint, method, "{}")
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

func (c *client) GetHeight() (uint, error) {
	return c.callGetHeight()
}
