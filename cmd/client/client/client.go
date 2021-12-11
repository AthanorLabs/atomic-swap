package client

// Client represents a swap client, used to interact with a swap daemon via JSON-RPC calls.
type Client struct {
	endpoint string
}

// NewClient ...
func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
	}
}
