package protocol

import (
	"os"
	"testing"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"

	"github.com/stretchr/testify/require"
)

func TestWriteKeysToFile(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	err = WriteKeysToFile(os.TempDir()+"/test.keys", kp, common.Development)
	require.NoError(t, err)
}

func TestWriteContractAddrssToFile(t *testing.T) {
	addr := "0xabcd"
	err := WriteContractAddressToFile(os.TempDir()+"/test.keys", addr)
	require.NoError(t, err)
}
