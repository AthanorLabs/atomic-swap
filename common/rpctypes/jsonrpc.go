// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpctypes

import (
	"encoding/json"
	"fmt"
)

// DefaultJSONRPCVersion ...
const DefaultJSONRPCVersion = "2.0"

// Request represents a JSON-RPC request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      uint64          `json:"id"`
}

// Response is the JSON format of a response
type Response struct {
	Version string           `json:"jsonrpc"`
	Result  json.RawMessage  `json:"result"`
	Error   *Error           `json:"error"`
	ID      *json.RawMessage `json:"id"`
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
	if e.ErrorCode != 0 {
		return fmt.Sprintf("message=%s; code=%d; data=%v", e.Message, e.ErrorCode, e.Data)
	}
	return e.Message
}
