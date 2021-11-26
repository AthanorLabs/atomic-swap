package monero

import (
	"encoding/json"

	"github.com/noot/atomic-swap/rpcclient"
)

// DaemonClient represents a monerod client.
type DaemonClient interface {
	GenerateBlocks(address string, amount uint) error
}

type generateBlocksRequest struct {
	Address        string `json:"wallet_address"`
	AmountOfBlocks uint   `json:"amount_of_blocks"`
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

	params, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := rpcclient.PostRPC(c.endpoint, method, string(params))
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return resp.Error
	}

	return nil
}
