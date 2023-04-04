// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"github.com/cockroachdb/apd/v3"
)

//
// Notes on the Exponent field of apd.Decimal:
// * All external apd.Decimal input values come from JSON.
// * apd.Decimal's unmarshalling code will throw an error if the exponent
//   of a value exceeds apd.MaxExponent (100,000).
// * Since our software will never have an apd.Decimal value with an
//   exponent that can be overflowed/underflowed, panic error handling
//   is fine.
// * Overflow checking is done and centralized here so code auditors do
//   do not waste time.
//

func increaseExponent(n *apd.Decimal, delta uint8) {
	delta32 := int32(delta)
	e := n.Exponent
	n.Exponent += delta32
	if n.Exponent < e {
		panic("overflow")
	}
}

func decreaseExponent(n *apd.Decimal, delta uint8) {
	delta32 := int32(delta)
	e := n.Exponent
	n.Exponent -= delta32
	if n.Exponent > e {
		panic("underflow")
	}
}
