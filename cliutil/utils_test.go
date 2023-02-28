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

func getKeyPath(t *testing.T) string {
	return path.Join(t.TempDir(), common.DefaultEthKeyFileName)
}

func verifyKeyFile(t *testing.T, keyPath string, expectedKeyHex string) {
	data, err := os.ReadFile(keyPath)
	require.NoError(t, err)
	require.Equal(t, expectedKeyHex, string(data))
}

func TestGetEthereumPrivateKey_devXMRMaker(t *testing.T) {
	devXMRMaker := true
	devXMRTaker := false
	keyPath := getKeyPath(t)
	key, err := GetEthereumPrivateKey(keyPath, common.Development, devXMRMaker, devXMRTaker)
	require.NoError(t, err)
	expectedKeyHex := hex.EncodeToString(ethcrypto.FromECDSA(key))
	require.Equal(t, common.DefaultPrivKeyXMRMaker, expectedKeyHex)
	verifyKeyFile(t, keyPath, expectedKeyHex)
}

func TestGetEthereumPrivateKey_devXMRTaker(t *testing.T) {
	devXMRMaker := false
	devXMRTaker := true
	keyPath := getKeyPath(t)
	key, err := GetEthereumPrivateKey(keyPath, common.Development, devXMRMaker, devXMRTaker)
	require.NoError(t, err)
	expectedKeyHex := hex.EncodeToString(ethcrypto.FromECDSA(key))
	require.Equal(t, common.DefaultPrivKeyXMRTaker, hex.EncodeToString(ethcrypto.FromECDSA(key)))
	verifyKeyFile(t, keyPath, expectedKeyHex)
}

func TestGetEthereumPrivateKey_nonDev_newKey(t *testing.T) {
	devXMRMaker := true // ignored, using stagenet
	devXMRTaker := true // ignored, using stagenet
	keyPath := getKeyPath(t)
	key, err := GetEthereumPrivateKey(keyPath, common.Stagenet, devXMRMaker, devXMRTaker)
	expectedKeyHex := hex.EncodeToString(ethcrypto.FromECDSA(key))
	require.NoError(t, err)
	verifyKeyFile(t, keyPath, expectedKeyHex)
	require.NotEqual(t, expectedKeyHex, common.DefaultPrivKeyXMRMaker)
}

func TestGetEthereumPrivateKey_fromFile(t *testing.T) {
	keyHex := "87c546d6cb8ec705bea47e2ab40f42a768b1e5900686b0cecc68c0e8b74cd789"
	fileData := []byte(fmt.Sprintf("  %s\n", keyHex)) // add whitespace that we should ignore
	keyPath := getKeyPath(t)
	require.NoError(t, os.WriteFile(keyPath, fileData, 0600))
	key, err := GetEthereumPrivateKey(keyPath, common.Mainnet, false, false)
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

func TestGetVersion(t *testing.T) {
	// Nothing we can test other than that it does not panic without a built executable
	require.NotEmpty(t, GetVersion())
	t.Log(GetVersion())
}

func Test_expandBootnodes(t *testing.T) {
	cliNodes := []string{
		" node1, node2 ,node3,node4 ",
		"node5",
		"\tnode6\n",
		"node7,node8",
	}
	expected := []string{
		"node1",
		"node2",
		"node3",
		"node4",
		"node5",
		"node6",
		"node7",
		"node8",
	}
	require.EqualValues(t, expected, ExpandBootnodes(cliNodes))
}

func Test_expandBootnodes_noNodes(t *testing.T) {
	// This can happen when the user specifies a single `--bootnodes ""` flag
	// to not use the default bootnodes for an environment.
	cliNodes := []string{""}
	nodes := ExpandBootnodes(cliNodes)
	require.Zero(t, len(nodes))
}
