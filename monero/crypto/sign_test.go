package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrivateSpendKey_Sign(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)

	msg := []byte("testmessage")
	sig, err := kp.sk.Sign(msg)
	require.NoError(t, err)
	require.NotNil(t, sig)
}

func TestPrivateSpendKey_Verify(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)

	msg := []byte("testmessage")
	sig, err := kp.sk.Sign(msg)
	require.NoError(t, err)
	require.NotNil(t, sig)

	ok := kp.sk.Public().Verify(msg, sig)
	require.True(t, ok)
}
