package tests

import (
	"github.com/athanorlabs/atomic-swap/coins"
)

// Str2Decimal converts strings to big decimal for tests, panicing on error.
// This function is intended for use with string constants, where panic is
// an acceptable behavior.
var Str2Decimal = coins.Str2Decimal
