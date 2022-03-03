package rpcclient

import (
	"encoding/json"
	"fmt"
)

// Request represents a JSON-RPC request
type Request struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      uint64                 `json:"id"`
}

// Response is the JSON format of a response
type Response struct {
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

// SubscribeSwapStatusResponse ...
type SubscribeSwapStatusResponse struct {
	Stage string `json:"stage"`
}
