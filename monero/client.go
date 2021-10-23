package monero

import (
	"fmt"
)

type Client interface {
	Transfer(to Address, accountIdx, amount uint) error
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
