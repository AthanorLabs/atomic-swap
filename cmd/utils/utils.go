package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli"

	"github.com/noot/atomic-swap/common"
)

const (
	flagEthereumPrivKey = "ethereum-privkey"
	flagEnv             = "env"
)

var log = logging.Logger("cmd")

var defaultEnvironment = common.Development

var (
	errNoEthereumPrivateKey = errors.New("must provide --ethereum-privkey file for non-development environment")
	errInvalidEnv           = errors.New("--env must be one of mainnet, stagenet, or dev")
)

// GetEthereumPrivateKey returns an ethereum private key hex string given the CLI options.
func GetEthereumPrivateKey(c *cli.Context, env common.Environment, devXMRMaker,
	useExternal bool) (ethPrivKey string, err error) {
	if c.String(flagEthereumPrivKey) != "" {
		ethPrivKeyFile := c.String(flagEthereumPrivKey)
		key, err := os.ReadFile(filepath.Clean(ethPrivKeyFile))
		if err != nil {
			return "", fmt.Errorf("failed to read ethereum-privkey file: %w", err)
		}

		if key[len(key)-1] == '\n' {
			key = key[:len(key)-1]
		}

		ethPrivKey = string(key)
	} else {
		if env != common.Development || useExternal {
			// TODO: allow this to be set via RPC
			log.Warnf("%s", errNoEthereumPrivateKey)
			return "", nil
		}

		log.Warn("no ethereum private key file provided, using ganache deterministic key")
		if devXMRMaker {
			ethPrivKey = common.DefaultPrivKeyXMRMaker
		} else {
			ethPrivKey = common.DefaultPrivKeyXMRTaker
		}
	}

	return ethPrivKey, nil
}

// GetEnvironment returns a common.Environment from the CLI options.
func GetEnvironment(c *cli.Context) (env common.Environment, cfg common.Config, err error) {
	switch c.String(flagEnv) {
	case "mainnet":
		env = common.Mainnet
		cfg = common.MainnetConfig
	case "stagenet":
		env = common.Stagenet
		cfg = common.StagenetConfig
	case "dev":
		env = common.Development
		cfg = common.DevelopmentConfig
	case "":
		env = defaultEnvironment
		cfg = common.DevelopmentConfig
	default:
		return 0, common.Config{}, errInvalidEnv
	}

	return env, cfg, nil
}
