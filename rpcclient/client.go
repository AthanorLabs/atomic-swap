package rpcclient

import (
	"context"
)

// Client represents a swap RPC client, used to interact with a swap daemon via JSON-RPC calls.
type Client struct {
	ctx      context.Context
	endpoint string
}

// NewClient ...
func NewClient(ctx context.Context, endpoint string) *Client {
	return &Client{
		ctx:      ctx,
		endpoint: endpoint,
	}
}
