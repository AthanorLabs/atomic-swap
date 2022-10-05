package cliutil

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/athanorlabs/atomic-swap/common"
)

var (
	errInvalidEnv = errors.New("--env must be one of mainnet, stagenet, or dev")
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

	return os.WriteFile(ethPrivKeyFile, []byte(privKeyStr), 0600)
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

// GetEnvironment returns a common.Environment from the CLI options.
func GetEnvironment(envStr string) (env common.Environment, cfg common.Config, err error) {
	switch envStr {
	case "mainnet":
		env = common.Mainnet
		cfg = common.MainnetConfig
	case "stagenet":
		env = common.Stagenet
		cfg = common.StagenetConfig
	case "dev":
		env = common.Development
		cfg = common.DevelopmentConfig
	default:
		return 0, common.Config{}, errInvalidEnv
	}

	return env, cfg, nil
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
