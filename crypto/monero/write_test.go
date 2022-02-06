package mcrypto

import (
	"os"
	"testing"

	"github.com/noot/atomic-swap/common"

	"github.com/stretchr/testify/require"
)

func TestWriteKeysToFile(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)

	err = WriteKeysToFile(os.TempDir()+"/", kp, common.Development)
	require.NoError(t, err)
}
