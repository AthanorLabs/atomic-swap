// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"math"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_increaseExponent(t *testing.T) {
	n := apd.New(2, 0) // 2.00
	increaseExponent(n, 2)
	assert.Equal(t, "200", n.Text('f'))
}

func Test_increaseExponent_overflowPanics(t *testing.T) {
	n := apd.New(2, math.MaxInt32)
	defer func() {
		require.NotNil(t, recover()) // if recover returns nil, we didn't panic
	}()
	increaseExponent(n, 1)
}

func Test_decreaseExponent(t *testing.T) {
	n := apd.New(300, 0) // 300
	decreaseExponent(n, 3)

	// Right now the string value would be 0.300, because trailing zeros in the
	// coefficient are printed. Next line will strip trailing zeros from the
	// coefficient and add them to the exponent making the string value 0.3.
	_, _ = n.Reduce(n)

	assert.Equal(t, "0.3", n.Text('f'))
}

func Test_decreaseExponent_underflowPanics(t *testing.T) {
	n := apd.New(2, math.MinInt32)
	defer func() {
		require.NotNil(t, recover()) // if recover returns nil, we didn't panic
	}()
	decreaseExponent(n, 1)
}
