package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/rpc"
)

// Shutdown swapd
func (c *Client) Shutdown() error {
	const (
		method = "daemon_shutdown"
	)
	if err := c.post(method, nil, nil); err != nil {
		return err
	}
	return nil
}

// Version returns version & misc info about swapd and its dependencies
func (c *Client) Version() (*rpc.VersionResponse, error) {
	const (
		method = "daemon_version"
	)
	resp := &rpc.VersionResponse{}
	if err := c.post(method, nil, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
