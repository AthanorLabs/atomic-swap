package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/rpc"
)

func (c *Client) Shutdown() error {
	const (
		method = "daemon_shutdown"
	)
	c.Post(method, nil, nil); // Does not expect a response from swapd
	return nil
}


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

