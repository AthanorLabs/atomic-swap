// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"fmt"
)

// ProvidesCoin represents a coin that a swap participant can provide.
type ProvidesCoin string

var (
	ProvidesXMR ProvidesCoin = "XMR" //nolint
	ProvidesETH ProvidesCoin = "ETH" //nolint
)

// NewProvidesCoin converts a string to a ProvidesCoin.
func NewProvidesCoin(s string) (ProvidesCoin, error) {
	switch s {
	case "XMR", "xmr":
		return ProvidesXMR, nil
	case "ETH", "eth":
		return ProvidesETH, nil
	default:
		return "", ErrInvalidCoin
	}
}

func (c *ProvidesCoin) String() string {
	return string(*c)
}

// MarshalText hands off JSON encoding to apd.Decimal
func (c *ProvidesCoin) MarshalText() ([]byte, error) {
	switch *c {
	case ProvidesXMR, ProvidesETH:
		return []byte(*c), nil
	}
	return nil, fmt.Errorf("cannot marshal ProvidesCoin %q", *c)
}

// UnmarshalText hands off JSON decoding to apd.Decimal
func (c *ProvidesCoin) UnmarshalText(data []byte) error {
	c2, err := NewProvidesCoin(string(data))
	if err != nil {
		return err
	}
	*c = c2
	return nil
}
