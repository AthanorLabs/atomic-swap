package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/rpc"
)

// Shutdown swapd
func (c *Client) Shutdown() error {
	const (
		method = "daemon_shutdown"
	)
	if err := c.Post(method, nil, nil); err != nil {
		return nil // Does not expect a response from swapd
	}
	return nil
}

// Version returns version & misc info about swapd and its dependencies
func (c *Client) Version() (*rpc.VersionResponse, error) {
	const (
		method = "daemon_version"
	)
	resp := &rpc.VersionResponse{}
	if err := c.Post(method, nil, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
