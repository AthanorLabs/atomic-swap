package monero

import (
	"fmt"
)

type Client interface {
	Transfer(to Address, amount uint) error
}

type client struct {
	endpoint string
}

func NewClient(endpoint string) *client {
	return &client{
		endpoint: endpoint,
	}
}

func (c *client) Transfer(to Address, amount uint) error {
	destination := Destination{
		Amount: amount,
		Address: string(to),
	}

	txhash, err := c.callTransfer([]Destination{destination})
	fmt.Println("Bob: locked XMR", txhash)
	return err
}