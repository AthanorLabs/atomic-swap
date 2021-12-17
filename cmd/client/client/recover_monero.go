package client

import (
	"encoding/json"

	mcrypto "github.com/noot/atomic-swap/monero/crypto"
	"github.com/noot/atomic-swap/rpc"
	"github.com/noot/atomic-swap/rpcclient"
)

// RecoverMonero calls the RPC function recover_monero
func (c *Client) RecoverMonero(as, bs, contractAddr string) (mcrypto.Address, error) {
	const (
		method = "recover_monero"
	)

	req := &rpc.RecoverMoneroRequest{
		AliceSecret:     as,
		BobSecret:       bs,
		ContractAddress: contractAddr,
	}

	params, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := rpcclient.PostRPC(c.endpoint, method, string(params))
	if err != nil {
		return "", err
	}

	if resp.Error != nil {
		return "", resp.Error
	}

	var res *rpc.RecoverMoneroResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return "", err
	}

	return mcrypto.Address(res.Address), nil
}
