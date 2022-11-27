package rpctypes

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

// PostRPC posts a JSON-RPC call to the given endpoint.
func PostRPC(endpoint, method string, request any, response any) error {
	data, err := json2.EncodeClientRequest(method, request)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", contentTypeJSON)

	// TODO: This context should be passed in
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
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
