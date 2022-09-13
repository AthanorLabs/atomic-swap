package cliutil

import (
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
)

func TestGetEthereumPrivateKey_devXMRMaker(t *testing.T) {
	devXMRMaker := true
	devXMRTaker := false
	key, err := GetEthereumPrivateKey("", common.Development, devXMRMaker, devXMRTaker)
	require.NoError(t, err)
	require.Equal(t, common.DefaultPrivKeyXMRMaker, hex.EncodeToString(ethcrypto.FromECDSA(key)))
}

func TestGetEthereumPrivateKey_devXMRTaker(t *testing.T) {
	devXMRMaker := false
	devXMRTaker := true
	key, err := GetEthereumPrivateKey("", common.Development, devXMRMaker, devXMRTaker)
	require.NoError(t, err)
	require.Equal(t, common.DefaultPrivKeyXMRTaker, hex.EncodeToString(ethcrypto.FromECDSA(key)))
}

func TestGetEthereumPrivateKey_devXMRMaker_nonDevEnv(t *testing.T) {
	devXMRMaker := true
	devXMRTaker := false
	_, err := GetEthereumPrivateKey("", common.Stagenet, devXMRMaker, devXMRTaker)
	require.ErrorIs(t, err, errNoEthereumPrivateKey)
}

func TestGetEthereumPrivateKey_fromFile(t *testing.T) {
	keyHex := "87c546d6cb8ec705bea47e2ab40f42a768b1e5900686b0cecc68c0e8b74cd789"
	fileData := []byte(fmt.Sprintf("  %s\n", keyHex)) // add whitespace that we should ignore
	keyFile := path.Join(t.TempDir(), "eth.key")
	require.NoError(t, os.WriteFile(keyFile, fileData, 0600))
	key, err := GetEthereumPrivateKey(keyFile, common.Mainnet, false, false)
	require.NoError(t, err)
	require.Equal(t, keyHex, hex.EncodeToString(ethcrypto.FromECDSA(key)))
}

func TestGetEthereumPrivateKey_fromFileFail(t *testing.T) {
	keyHex := "87c546d6cb8ec705bea47e2ab40f42a768b1e5900686b0cecc68c0e8b74cd789"
	keyBytes, err := hex.DecodeString(keyHex)
	require.NoError(t, err)
	keyFile := path.Join(t.TempDir(), "eth.key")
	require.NoError(t, os.WriteFile(keyFile, keyBytes, 0600)) // key is binary instead of hex
	_, err = GetEthereumPrivateKey(keyFile, common.Mainnet, false, false)
	require.ErrorContains(t, err, "invalid hex character")
}

func TestGetEnvironment(t *testing.T) {
	expected := map[string]common.Environment{
		"mainnet":  common.Mainnet,
		"stagenet": common.Stagenet,
		"dev":      common.Development,
	}

	for cliVal, expectedResult := range expected {
		env, cfg, err := GetEnvironment(cliVal)
		require.NoError(t, err)
		require.Equal(t, expectedResult, env)
		require.NotEmpty(t, cfg.DataDir)
	}
}

func TestGetEnvironment_fail(t *testing.T) {
	_, _, err := GetEnvironment("goerli")
	require.ErrorIs(t, err, errInvalidEnv)
}

func TestGetVersion(t *testing.T) {
	// Nothing we can test other than that it does not panic without a built executable
	require.NotEmpty(t, GetVersion())
	t.Log(GetVersion())
}
