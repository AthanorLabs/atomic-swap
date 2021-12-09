package rpc

import (
	"fmt"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
)

// Codec ...
type Codec struct{}

// NewCodec ...
func NewCodec() *Codec {
	return new(Codec)
}

// NewRequest ...
func (c *Codec) NewRequest(req *http.Request) rpc.CodecRequest {
	outer := &CodecRequest{}
	inner := json2.NewCodec().NewRequest(req)
	outer.CodecRequest = inner.(*json2.CodecRequest)
	return outer
}

// CodecRequest ...
type CodecRequest struct {
	*json2.CodecRequest
}

// Method ...
func (cr *CodecRequest) Method() (string, error) {
	method, err := cr.CodecRequest.Method()
	if err != nil {
		return "", err
	}

	parts := strings.Split(method, "_")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid method %s", method)
	}

	service, method := parts[0], parts[1]
	r, n := utf8.DecodeRuneInString(method)
	if unicode.IsLower(r) {
		return fmt.Sprintf("%s.%s%s", service, string(unicode.ToUpper(r)), method[n:]), nil
	}

	return fmt.Sprintf("%s.%s", service, method), nil
}
