// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package common

import (
	"math/big"
	"os"
	"path"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

const (
	// DefaultMoneroWalletName is the default wallet name in {DATA_DIR}/wallet/
	DefaultMoneroWalletName = "swap-wallet"

	// DefaultLibp2pKeyFileName is the default libp2p private key file name in {DATA_DIR}
	DefaultLibp2pKeyFileName = "net.key"

	// DefaultEthKeyFileName is the default ethereum private key file name in {DATA_DIR}
	DefaultEthKeyFileName = "eth.key"
)

var homeDir, _ = os.UserHomeDir()
var baseDir = path.Join(homeDir, ".atomicswap")

// MoneroNode represents the host and port of monerod's RPC endpoint
type MoneroNode struct {
	Host string
	Port uint
}

// Config contains constants that are defaults for various environments
type Config struct {
	Env             Environment
	DataDir         string
	EthEndpoint     string
	MoneroNodes     []*MoneroNode
	SwapCreatorAddr ethcommon.Address
	Bootnodes       []string
}

// MainnetConfig is the mainnet ethereum and monero configuration
func MainnetConfig() *Config {
	return &Config{
		Env:         Mainnet,
		DataDir:     path.Join(baseDir, "mainnet"),
		EthEndpoint: "", // No mainnet default (permissionless URLs are not reliable)
		MoneroNodes: []*MoneroNode{
			{
				Host: "node.sethforprivacy.com",
				Port: 18089,
			},
			{
				Host: "xmr-node.cakewallet.com",
				Port: DefaultMoneroDaemonMainnetPort,
			},
			{
				Host: "node.monerodevs.org",
				Port: 18089,
			},
			{
				Host: "node.community.rino.io",
				Port: DefaultMoneroDaemonMainnetPort,
			},
		},
		SwapCreatorAddr: ethcommon.HexToAddress("0xa55aa5557ec22e85804729bc6935029bb84cf16a"),
		Bootnodes: []string{
			"/ip4/67.205.131.11/tcp/9909/p2p/12D3KooWGpCLC4y42rf6aR3cguVFJAruzFXT6mUEyp7C32jTsyJd",
			"/ip4/143.198.123.27/tcp/9909/p2p/12D3KooWDCE2ukB1Sw88hmLFk5BZRRViyYLeuAKPuu59nYyFWAec",
			"/ip4/67.207.89.83/tcp/9909/p2p/12D3KooWED1Y5nfno34Qhz2Xj9ubmwi4hv2qd676pH6Jb7ui36CR",
			"/ip4/109.105.198.218/tcp/9909/p2p/12D3KooWRbAV4AU75JXVUedmwakiM1yzzDzDCdmZWzGdXGQ4Yqdv",
			"/ip4/185.130.46.66/tcp/9909/p2p/12D3KooWDKf2FJG1AWTJthbs7fcCcsQa26f4pmCR25cktRg2X2aY",
			"/ip4/93.95.228.200/tcp/9909/p2p/12D3KooWJParpZ1zHDspoV4kogkBsHKrGxMeq3UGFxQUm6TZPojn",
			"/ip4/31.220.60.19/tcp/9909/p2p/12D3KooWLksqqtzwA4Epg5eCxA2EaJ3Q34RW554HSqUDp2bgromP",
		},
	}
}

// StagenetConfig is the monero stagenet and ethereum Sepolia configuration
func StagenetConfig() *Config {
	return &Config{
		Env:         Stagenet,
		DataDir:     path.Join(baseDir, "stagenet"),
		EthEndpoint: "https://rpc.sepolia.org/",
		MoneroNodes: []*MoneroNode{
			{
				Host: "node.sethforprivacy.com",
				Port: 38089,
			},
			{
				Host: "node.monerodevs.org",
				Port: 38089,
			},
			{
				Host: "stagenet.community.rino.io",
				Port: 38081,
			},
		},
		SwapCreatorAddr: ethcommon.HexToAddress("0x377ed3a60007048DF00135637521170628De89E5"),
		Bootnodes: []string{
			"/ip4/134.122.115.208/tcp/9900/p2p/12D3KooWHZ2G9XscjDGvG7p8uPBoYerDc9kWYnc8oJFGfFxS6gfq",
			"/ip4/143.198.123.27/tcp/9900/p2p/12D3KooWGzExs5zqebnDvqkUAKaiUuxF3DNbrfJ4prbfkxjXb366",
			"/ip4/67.207.89.83/tcp/9900/p2p/12D3KooWLidMFT3WJaBK7cEzmpDEcuRcNTux89SVrZmEiGLhTWoe",
			"/ip4/67.205.131.11/tcp/9900/p2p/12D3KooWJxA1tPLugao43gPvE4rCcgG7xwQ4dhbLCUQehAiohJk2",
			"/ip4/164.92.103.159/tcp/9900/p2p/12D3KooWSNQF1eNyapxC2zA3jJExgLX7jWhEyw8B3k7zMW5ZRvQz",
			"/ip4/164.92.123.10/tcp/9900/p2p/12D3KooWG8z9fXVTB72XL8hQbahpfEjutREL9vbBQ4FzqtDKzTBu",
			"/ip4/161.35.110.210/tcp/9900/p2p/12D3KooWS8iKxqsGTiL3Yc1VaAfg99U5km1AE7bWYQiuavXj3Yz6",
		},
	}
}

// DevelopmentConfig is the monero and ethereum development environment configuration
func DevelopmentConfig() *Config {
	return &Config{
		Env:         Development,
		DataDir:     path.Join(baseDir, "dev"),
		EthEndpoint: DefaultGanacheEndpoint,
		MoneroNodes: []*MoneroNode{
			{
				Host: "127.0.0.1",
				Port: DefaultMoneroDaemonMainnetPort,
			},
		},
	}
}

// MoneroWalletPath returns the path to the wallet file, whose default value
// depends on current value of the data dir.
func (c Config) MoneroWalletPath() string {
	return path.Join(c.DataDir, "wallet", DefaultMoneroWalletName)
}

// LibP2PKeyFile returns the path to the libp2p key file, whose default value
// depends on current value of the data dir.
func (c Config) LibP2PKeyFile() string {
	return path.Join(c.DataDir, DefaultLibp2pKeyFileName)
}

// EthKeyFileName returns the path to the ethereum key file, whose default value
// depends on current value of the data dir.
func (c Config) EthKeyFileName() string {
	return path.Join(c.DataDir, DefaultEthKeyFileName)
}

// ConfigDefaultsForEnv returns the configuration defaults for the given environment.
func ConfigDefaultsForEnv(env Environment) *Config {
	switch env {
	case Mainnet:
		return MainnetConfig()
	case Stagenet:
		return StagenetConfig()
	case Development:
		return DevelopmentConfig()
	default:
		panic("invalid environment")
	}
}

// SwapTimeoutFromEnv returns the duration between swap timeouts given the environment.
func SwapTimeoutFromEnv(env Environment) time.Duration {
	switch env {
	case Mainnet, Stagenet:
		return time.Hour
	case Development:
		return time.Minute * 2
	default:
		panic("invalid environment")
	}
}

// DefaultMoneroPortFromEnv returns the default Monerod RPC port for an environment
// Reference: https://monerodocs.org/interacting/monerod-reference/
func DefaultMoneroPortFromEnv(env Environment) uint {
	switch env {
	case Mainnet:
		return DefaultMoneroDaemonMainnetPort
	case Stagenet:
		return DefaultMoneroDaemonStagenetPort
	case Development:
		return DefaultMoneroDaemonDevPort
	default:
		panic("invalid environment")
	}
}

// ChainIDFromEnv returns the expected chainID that we should find on the
// ethereum endpoint when running int the passed environment.
func ChainIDFromEnv(env Environment) *big.Int {
	switch env {
	case Development:
		return big.NewInt(GanacheChainID)
	case Stagenet:
		return big.NewInt(SepoliaChainID)
	case Mainnet:
		return big.NewInt(MainnetChainID)
	default:
		panic("invalid environment")
	}
}
