package coins

import (
	"strings"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePositive_pass(t *testing.T) {
	const fieldName = "testValue"
	xmrValue := StrToDecimal("200.123456789012")
	err := ValidatePositive(fieldName, NumMoneroDecimals, xmrValue)
	require.NoError(t, err)

	ethValue := StrToDecimal("0.123456789012345678")
	err = ValidatePositive(fieldName, NumEtherDecimals, ethValue)
	require.NoError(t, err)
}

func TestValidatePositive_errors(t *testing.T) {
	const fieldName = "testValue"
	type entry struct {
		numDecPlaces uint8
		value        *apd.Decimal
		errContains  string
	}
	testEntries := []entry{
		{
			numDecPlaces: 1,
			value:        nil,
			errContains:  `"testValue" is not set`,
		},
		{
			numDecPlaces: 1,
			value:        new(apd.Decimal),
			errContains:  `"testValue" must be non-zero`,
		},
		{
			numDecPlaces: 1,
			value:        StrToDecimal("-1"),
			errContains:  `"testValue" cannot be negative`,
		},
		{
			numDecPlaces: 1,
			value:        StrToDecimal(strings.Repeat("6", MaxCoinPrecision+1)),
			errContains:  `"testValue" has too many digits`,
		},
		{
			numDecPlaces: 0,
			value:        StrToDecimal("1.1"), // 1 decimal place, zero allowed
			errContains:  `"testValue" has too many decimal points; found=1 max=0`,
		},
		{
			numDecPlaces: NumMoneroDecimals,
			value:        StrToDecimal("1.1234567890123"),
			errContains:  `"testValue" has too many decimal points; found=13 max=12`,
		},
		{
			numDecPlaces: NumEtherDecimals,
			value:        StrToDecimal("1.12345678901234567890000000000000000000"), // zeros at end ignored
			errContains:  `"testValue" has too many decimal points; found=19 max=18`,
		},
	}
	for _, entry := range testEntries {
		err := ValidatePositive(fieldName, entry.numDecPlaces, entry.value)
		assert.ErrorContains(t, err, entry.errContains)
	}
}
