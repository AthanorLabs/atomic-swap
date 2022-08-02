package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReverse(t *testing.T) {
	in := []byte{0xa, 0xb, 0xc}
	expected := []byte{0xc, 0xb, 0xa}
	require.Equal(t, expected, Reverse(in))
	require.Equal(t, []byte{0xa, 0xb, 0xc}, in) // backing array of original slice is unmodified

	in2 := [3]byte{0xa, 0xb, 0xc}
	require.Equal(t, expected, Reverse(in2[:]))
	require.Equal(t, in2, [3]byte{0xa, 0xb, 0xc}) // input array is unmodified

}
