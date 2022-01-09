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

var log = logging.Logger("cmd")

func GetEthereumPrivateKey(c *cli.Context, env common.Environment, devBob bool) (ethPrivKey string, err error) {
	if c.String("ethereum-privkey") != "" {
		ethPrivKeyFile := c.String("ethereum-privkey")
		key, err := os.ReadFile(filepath.Clean(ethPrivKeyFile))
		if err != nil {
			return "", fmt.Errorf("failed to read ethereum-privkey file: %w", err)
		}

		if key[len(key)-1] == '\n' {
			key = key[:len(key)-1]
		}

		ethPrivKey = string(key)
	} else {
		if env != common.Development {
			// TODO: allow this to be set via RPC
			return "", errors.New("must provide --ethereum-privkey file for non-development environment")
		}

		log.Warn("no ethereum private key file provided, using ganache deterministic key")
		if devBob {
			ethPrivKey = common.DefaultPrivKeyBob
		} else {
			ethPrivKey = common.DefaultPrivKeyAlice
		}
	}

	return ethPrivKey, nil
}

func GetEnvironment(c *cli.Context) (env common.Environment, cfg common.Config, err error) {
	switch c.String("env") {
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
		env = common.Development
		cfg = common.DevelopmentConfig
	default:
		return 0, common.Config{}, errors.New("--env must be one of mainnet, stagenet, or dev")
	}

	return env, cfg, nil
}
