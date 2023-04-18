// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"testing"

	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/assert"
)

func init() {
	logging.SetLogLevel("coins", "debug")
}

func TestDecimalCtx(t *testing.T) {
	c := DecimalCtx()
	assert.Equal(t, c.Precision, uint32(MaxCoinPrecision))
	// verify that package variable decimalCtx is unmodified, because c is a copy
	c.Precision = 3
	assert.Equal(t, decimalCtx.Precision, uint32(MaxCoinPrecision))
}
