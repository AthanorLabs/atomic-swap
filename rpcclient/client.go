package rpcclient

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/rpc/v2/json2"
)

var (
	contentTypeJSON   = "application/json"
	dialTimeout       = 60 * time.Second
	httpClientTimeout = 30 * time.Minute
	callTimeout       = 30 * time.Minute

	transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: dialTimeout,
		}).DialContext,
	}
	httpClient = &http.Client{
		Transport: transport,
		Timeout:   httpClientTimeout,
	}
)

// Client represents a swap RP
//
//	"time"
//
//	"github.com/gorilla/rpc/v2/json2"C client, used to interact with a swap daemon via JSON-RPC calls.
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

// Post makes a JSON-RPC call to the client's endpoint, serialising any passed request
// object and deserializing any passed response object from the POST response body. Nil
// can be passed as the request or response when no data needs to be serialised or
// deserialised respectively.
func (c *Client) Post(method string, request any, response any) error {
	data, err := json2.EncodeClientRequest(method, request)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", c.endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", contentTypeJSON)

	ctx, cancel := context.WithTimeout(c.ctx, callTimeout)
	defer cancel()
	httpReq = httpReq.WithContext(ctx)

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to post %q request: %w", method, err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if response == nil {
		return nil
	}

	if err = json2.DecodeClientResponse(httpResp.Body, response); err != nil {
		return fmt.Errorf("failed to read %q response: %w", method, err)
	}

	return nil
}
