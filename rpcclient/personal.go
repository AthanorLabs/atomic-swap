// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/rpc"
)

// SetSwapTimeout calls personal_setSwapTimeout.
func (c *Client) SetSwapTimeout(timeoutSeconds uint64) error {
	const (
		method = "personal_setSwapTimeout"
	)

	req := &rpc.SetSwapTimeoutRequest{
		Timeout: timeoutSeconds,
	}

	if err := c.post(method, req, nil); err != nil {
		return err
	}

	return nil
}

// GetSwapTimeout calls personal_getSwapTimeout.
func (c *Client) GetSwapTimeout() (*rpc.GetSwapTimeoutResponse, error) {
	const (
		method = "personal_getSwapTimeout"
	)

	swapTimeout := &rpc.GetSwapTimeoutResponse{}
	if err := c.post(method, nil, swapTimeout); err != nil {
		return nil, err
	}

	return swapTimeout, nil
}

// TokenInfo calls personal_tokenInfo
func (c *Client) TokenInfo(tokenAddr ethcommon.Address) (*coins.ERC20TokenInfo, error) {
	const (
		method = "personal_tokenInfo"
	)

	// Note: coins.ERC20TokenInfo and rpctypes.TokenInfoRequest are aliases
	request := &rpctypes.TokenInfoRequest{TokenAddr: tokenAddr}
	tokenInfo := new(rpctypes.TokenInfoResponse)

	if err := c.post(method, request, tokenInfo); err != nil {
		return nil, err
	}

	return tokenInfo, nil
}

// Balances calls personal_balances.
func (c *Client) Balances(request *rpctypes.BalancesRequest) (*rpctypes.BalancesResponse, error) {
	const (
		method = "personal_balances"
	)

	balances := &rpctypes.BalancesResponse{}
	if err := c.post(method, request, balances); err != nil {
		return nil, err
	}

	return balances, nil
}

// TransferXMR calls personal_transferXMR
func (c *Client) TransferXMR(request *rpc.TransferXMRRequest) (*rpc.TransferXMRResponse, error) {
	const (
		method = "personal_transferXMR"
	)

	resp := new(rpc.TransferXMRResponse)
	if err := c.post(method, request, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// SweepXMR calls personal_sweepXMR
func (c *Client) SweepXMR(request *rpc.SweepXMRRequest) (*rpc.SweepXMRResponse, error) {
	const (
		method = "personal_sweepXMR"
	)

	resp := new(rpc.SweepXMRResponse)
	if err := c.post(method, request, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// TransferETH calls personal_transferETH
func (c *Client) TransferETH(request *rpc.TransferETHRequest) (*rpc.TransferETHResponse, error) {
	const (
		method = "personal_transferETH"
	)

	resp := new(rpc.TransferETHResponse)
	if err := c.post(method, request, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
