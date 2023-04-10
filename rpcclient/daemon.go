package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/rpc"
)

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
