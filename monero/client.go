package monero

import (
	"fmt"

	"github.com/noot/atomic-swap/common"
)

type Client interface {
	GetAccounts() (*getAccountsResponse, error)
	GetAddress(idx uint) (*getAddressResponse, error)
	GetBalance(idx uint) (*getBalanceResponse, error)
	Transfer(to Address, accountIdx, amount uint) error
	GenerateFromKeys(kp *PrivateKeyPair, filename, password string, env common.Environment) error
	GenerateViewOnlyWalletFromKeys(vk *PrivateViewKey, address Address, filename, password string) error
	Refresh() error
	OpenWallet(filename, password string) error
	CloseWallet() error
}

type client struct {
	endpoint string
}

func NewClient(endpoint string) *client {
	return &client{
		endpoint: endpoint,
	}
}

func (c *client) GetAccounts() (*getAccountsResponse, error) {
	return c.callGetAccounts()
}

func (c *client) GetBalance(idx uint) (*getBalanceResponse, error) {
	return c.callGetBalance(idx)
}

func (c *client) Transfer(to Address, accountIdx, amount uint) error {
	destination := Destination{
		Amount:  amount,
		Address: string(to),
	}

	txhash, err := c.callTransfer([]Destination{destination}, accountIdx)
	fmt.Printf("transfer: txhash=%s\n", txhash)
	return err
}

func (c *client) GenerateFromKeys(kp *PrivateKeyPair, filename, password string, env common.Environment) error {
	return c.callGenerateFromKeys(kp.sk, kp.vk, kp.Address(env), filename, password)
}

func (c *client) GenerateViewOnlyWalletFromKeys(vk *PrivateViewKey, address Address, filename, password string) error {
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

	resp, err := postRPC(c.endpoint, method, "{}")
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

	resp, err := postRPC(c.endpoint, method, "{}")
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return resp.Error
	}

	return nil
}
