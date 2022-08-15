package monero

import (
	"github.com/MarinX/monerorpc"
)

// DaemonClient represents a monerod client.
type DaemonClient interface {
	GenerateBlocks(address string, amount uint) error
}

// NewDaemonClient returns a new monerod client.
func NewDaemonClient(endpoint string) *client {
	return &client{
		rpc: monerorpc.New(endpoint, nil),
	}
}

type generateBlocksRequest struct {
	Address        string `json:"wallet_address"`
	AmountOfBlocks uint   `json:"amount_of_blocks"`
}

type generateBlocksResponse struct {
	Blocks []string `json:"blocks"`
	Height int      `json:"height"`
}

func (c *client) GenerateBlocks(address string, amount uint) error {
	return c.callGenerateBlocks(address, amount)
}

func (c *client) callGenerateBlocks(address string, amount uint) error {
	const method = "generateblocks"
	req := &generateBlocksRequest{
		Address:        address,
		AmountOfBlocks: amount,
	}
	resp := &generateBlocksResponse{}
	return c.rpc.Do(method, req, resp)
}
