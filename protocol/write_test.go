package protocol

import (
	"path"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"

	"github.com/stretchr/testify/require"
)

func TestWriteKeysToFile(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	err = WriteKeysToFile(path.Join(t.TempDir(), "test.keys"), kp, common.Development)
	require.NoError(t, err)
}

func TestWriteContractAddrssToFile(t *testing.T) {
	addr := "0xabcd"
	err := WriteContractAddressToFile(path.Join(t.TempDir(), "test.keys"), addr)
	require.NoError(t, err)
}
