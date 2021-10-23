package monero

import (
	"fmt"
)

type Client interface {
	GetAddress(idx uint) (*getAddressResponse, error)
	GetBalance(idx uint) (*getBalanceResponse, error)
	Transfer(to Address, accountIdx, amount uint) error
	GenerateFromKeys(kp *PrivateKeyPair, filename, password string) error
}

type client struct {
	endpoint string
}

func NewClient(endpoint string) *client {
	return &client{
		endpoint: endpoint,
	}
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

func (c *client) GenerateFromKeys(kp *PrivateKeyPair, filename, password string) error {
	address := kp.Address()
	return c.callGenerateFromKeys(kp.sk, kp.vk, address, filename, password)
}

func (c *client) GetAddress(idx uint) (*getAddressResponse, error) {
	return c.callGetAddress(idx)
}
