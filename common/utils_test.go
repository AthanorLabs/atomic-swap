package common

import (
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"

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

func TestGetTopic(t *testing.T) {
	refundedTopic := ethcommon.HexToHash("0x007c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f")
	require.Equal(t, GetTopic(RefundedEventSignature), refundedTopic)
}
