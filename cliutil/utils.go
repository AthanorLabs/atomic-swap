// Package cliutil provides utility functions intended for sharing by the main packages of multiple executables.
package cliutil

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/common"
)

var (
	// Only use this logger in functions called by programs that use formatted logs like swapd (not swapcli)
	log = logging.Logger("cmd")
)

func createAndWriteEthKeyFile(ethPrivKeyFile string, env common.Environment, devXMRMaker, devXMRTaker bool) error {
	var key *ecdsa.PrivateKey
	var err error

	switch {
	case env == common.Development && devXMRMaker:
		key, err = ethcrypto.HexToECDSA(common.DefaultPrivKeyXMRMaker)
	case env == common.Development && devXMRTaker:
		key, err = ethcrypto.HexToECDSA(common.DefaultPrivKeyXMRTaker)
	default:
		key, err = ethcrypto.GenerateKey()
	}
	if err != nil {
		return err
	}

	privKeyStr := hexutil.Encode(ethcrypto.FromECDSA(key))
	privKeyStr = strings.TrimPrefix(privKeyStr, "0x")

	if err := os.WriteFile(ethPrivKeyFile, []byte(privKeyStr), 0600); err != nil {
		return err
	}

	log.Infof("New ETH wallet key generated in %s", ethPrivKeyFile)
	log.Infof("Fund address %s to take an offer",
		ethcrypto.PubkeyToAddress(*(key.Public().(*ecdsa.PublicKey))).Hex())
	return nil
}

// GetEthereumPrivateKey reads or creates and returns an ethereum private key for the given the CLI options.
func GetEthereumPrivateKey(ethPrivKeyFile string, env common.Environment, devXMRMaker, devXMRTaker bool) (
	key *ecdsa.PrivateKey,
	err error,
) {
	if ethPrivKeyFile == "" {
		panic("missing required parameter ethPrivKeyFile")
	}
	exists, err := common.FileExists(ethPrivKeyFile)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err = createAndWriteEthKeyFile(ethPrivKeyFile, env, devXMRMaker, devXMRTaker); err != nil {
			return nil, err
		}
	}

	fileData, err := os.ReadFile(filepath.Clean(ethPrivKeyFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read ethereum-privkey file: %w", err)
	}
	ethPrivKeyHex := strings.TrimSpace(string(fileData))
	return ethcrypto.HexToECDSA(ethPrivKeyHex)
}

// GetVersion returns our version string for an executable
func GetVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown-version"
	}

	commitHash := "???????"
	dirty := ""

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			commitHash = setting.Value
		case "vcs.modified":
			if setting.Value == "true" {
				dirty = "-dirty"
			}
		}
	}

	return fmt.Sprintf("%s %.7s%s-%s",
		info.Main.Version, // " (devel)" unless passing a git tagged version to `go install`
		commitHash,        // 7 bytes is what "git rev-parse --short HEAD" returns
		dirty,             // add "-dirty" to commit hash if repo was not clean
		info.GoVersion,
	)
}
