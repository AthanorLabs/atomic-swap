// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_roundToDecimalPlace(t *testing.T) {
	// Round half down
	amt := StrToDecimal("33.4999999999999999999999999999999999")
	err := roundToDecimalPlace(amt, amt, 0)
	require.NoError(t, err)
	assert.Equal(t, "33", amt.String())

	// Round half up
	amt = StrToDecimal("33.5")
	err = roundToDecimalPlace(amt, amt, 0)
	require.NoError(t, err)
	assert.Equal(t, "34", amt.String())

	// Round at Decimal position
	amt = StrToDecimal("0.00009")
	res := new(apd.Decimal) // use a separate result variable this time
	err = roundToDecimalPlace(res, amt, 4)
	require.NoError(t, err)
	assert.Equal(t, "0.0001", res.String())
	assert.Equal(t, "0.00009", amt.String()) // input value unchanged
}
