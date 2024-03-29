// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package rpcclient provides client libraries for interacting with a local swapd instance using
// the JSON-RPC remote procedure call protocol and websockets.
package rpcclient

import (
	"bytes"
	"context"
	"errors"
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

// Client primarily exists to be a JSON-RPC client to swapd instances, but it can be used
// to POST JSON-RPC requests to any JSON-RPC server. Its current use case assumes swapd is
// running on the local host of a single use system. TLS and authentication are not
// currently supported.
type Client struct {
	ctx        context.Context
	endpoint   string
	wsEndpoint string
}

// NewClient creates a new JSON-RPC client for the specified endpoint. The passed context
// is used for the full lifetime of the client.
func NewClient(ctx context.Context, port uint16) *Client {
	return &Client{
		ctx:        ctx,
		endpoint:   fmt.Sprintf("http://127.0.0.1:%d", port),
		wsEndpoint: fmt.Sprintf("ws://127.0.0.1:%d/ws", port),
	}
}

// post makes a JSON-RPC call to the client's endpoint, serializing any passed request
// object and deserializing any passed response object from the POST response body. Nil
// can be passed as the request or response when no data needs to be serialized or
// deserialized respectively.
func (c *Client) post(method string, request any, response any) error {
	data, err := json2.EncodeClientRequest(method, request)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewReader(data))
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

	// Even if the response is nil, we still need to parse the outer JSON-RPC
	// shell to get any error that the server may have returned.
	err = json2.DecodeClientResponse(httpResp.Body, response)
	if response == nil && errors.Is(err, json2.ErrNullResult) {
		err = nil
	}
	if err != nil {
		return fmt.Errorf("%q failed: %w", method, err)
	}

	return nil
}
