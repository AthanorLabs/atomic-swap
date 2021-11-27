package rpcclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var (
	contentTypeJSON   = "application/json"
	dialTimeout       = 60 * time.Second
	httpClientTimeout = 30 * time.Minute
	callTimeout       = 30 * time.Minute

	transport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: dialTimeout,
		}).Dial,
	}
	httpClient = &http.Client{
		Transport: transport,
		Timeout:   httpClientTimeout,
	}
)

// ServerResponse is the JSON format of a response
type ServerResponse struct {
	// JSON-RPC Version
	Version string `json:"jsonrpc"`
	// Resulting values
	Result json.RawMessage `json:"result"`
	// Any generated errors
	Error *Error `json:"error"`
	// Request id
	ID *json.RawMessage `json:"id"`
}

// ErrCode is a int type used for the rpc error codes
type ErrCode int

// Error is a struct that holds the error message and the error code for a error
type Error struct {
	Message   string                 `json:"message"`
	ErrorCode ErrCode                `json:"code"`
	Data      map[string]interface{} `json:"data"`
}

// Error ...
func (e *Error) Error() string {
	return fmt.Sprintf("message=%s; code=%d; data=%v", e.Message, e.ErrorCode, e.Data)
}

// PostRPC posts a JSON-RPC call to the given endpoint.
func PostRPC(endpoint, method, params string) (*ServerResponse, error) {
	data := []byte(`{"jsonrpc":"2.0","method":"` + method + `","params":` + params + `,"id":0}`)
	buf := &bytes.Buffer{}
	_, err := buf.Write(data)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("POST", endpoint, buf)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", contentTypeJSON)

	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()
	r = r.WithContext(ctx)

	resp, err := httpClient.Do(r)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var sv *ServerResponse
	if err = json.Unmarshal(body, &sv); err != nil {
		return nil, err
	}

	return sv, nil
}
