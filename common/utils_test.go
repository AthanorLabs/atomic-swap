package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReverse(t *testing.T) {
	in := []byte{0xa, 0xb, 0xc}
	expected := []byte{0xc, 0xb, 0xa}
	res := Reverse(in)
	require.Equal(t, expected, in)
	require.Equal(t, expected, res)

	in2 := [3]byte{0xa, 0xb, 0xc}
	res = Reverse(in2[:])
	require.Equal(t, expected, in2[:])
	require.Equal(t, expected, res)
}
