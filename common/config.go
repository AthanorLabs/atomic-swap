package common

import (
	"fmt"
	"os"
)

var homeDir, _ = os.UserHomeDir()

// Config contains constants that are defaults for various environments
type Config struct {
	Basepath             string
	MoneroDaemonEndpoint string
	EthereumChainID      int64
	Bootnodes            []string // TODO: when it's ready for users to test, add some bootnodes
}

// MainnetConfig is the mainnet ethereum and monero configuration
var MainnetConfig = Config{
	Basepath:             fmt.Sprintf("%s/.atomicswap/mainnet", homeDir),
	MoneroDaemonEndpoint: "http://127.0.0.1:18081/json_rpc",
	EthereumChainID:      MainnetChainID,
}

// StagenetConfig is the monero stagenet and ethereum ropsten configuration
var StagenetConfig = Config{
	Basepath:             fmt.Sprintf("%s/.atomicswap/stagenet", homeDir),
	MoneroDaemonEndpoint: "http://127.0.0.1:38081/json_rpc",
	EthereumChainID:      RopstenChainID,
}

// DevelopmentConfig is the monero and ethereum development environment configuration
var DevelopmentConfig = Config{
	Basepath:             fmt.Sprintf("%s/.atomicswap/dev", homeDir),
	MoneroDaemonEndpoint: "http://127.0.0.1:18081/json_rpc",
	EthereumChainID:      GanacheChainID,
}
