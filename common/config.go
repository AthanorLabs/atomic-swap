package common

import (
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
	Env                      Environment
	DataDir                  string
	MoneroNodes              []*MoneroNode
	SwapFactoryAddress       ethcommon.Address
	ForwarderContractAddress ethcommon.Address
	Bootnodes                []string
}

// MainnetConfig is the mainnet ethereum and monero configuration
func MainnetConfig() *Config {
	return &Config{
		Env:     Mainnet,
		DataDir: path.Join(baseDir, "mainnet"),
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
	}
}

// StagenetConfig is the monero stagenet and ethereum Sepolia configuration
func StagenetConfig() *Config {
	return &Config{
		Env:     Stagenet,
		DataDir: path.Join(baseDir, "stagenet"),
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
		SwapFactoryAddress:       ethcommon.HexToAddress("0xE5656d3549478BdE248e5D7aE22ff8DCC5Cd11a3"),
		ForwarderContractAddress: ethcommon.HexToAddress("0x171f294A0Ca7085Ce5F73DE1BE28b9721e007B94"),
		Bootnodes: []string{
			"/ip4/134.122.115.208/tcp/9900/p2p/12D3KooWDqCzbjexHEa8Rut7bzxHFpRMZyDRW1L6TGkL1KY24JH5",
			"/ip4/143.198.123.27/tcp/9900/p2p/12D3KooWSc4yFkPWBFmPToTMbhChH3FAgGH96DNzSg5fio1pQYoN",
			"/ip4/67.207.89.83/tcp/9900/p2p/12D3KooWLbfkLZZvvn8Lxs1KDU3u7gyvBk88ZNtJBbugytBr5RCG",
			"/ip4/134.122.115.208/tcp/9900/p2p/12D3KooWDqCzbjexHEa8Rut7bzxHFpRMZyDRW1L6TGkL1KY24JH5",
			"/ip4/67.205.131.11/tcp/9900/p2p/12D3KooWT19g8cfBVYiGWkksU1ZojHCBNqTu3Hz5JLfhhytaHSwi",
			"/ip4/164.92.103.159/tcp/9900/p2p/12D3KooWSNQF1eNyapxC2zA3jJExgLX7jWhEyw8B3k7zMW5ZRvQz",
			"/ip4/164.92.123.10/tcp/9900/p2p/12D3KooWG8z9fXVTB72XL8hQbahpfEjutREL9vbBQ4FzqtDKzTBu",
			"/ip4/161.35.110.210/tcp/9900/p2p/12D3KooWS8iKxqsGTiL3Yc1VaAfg99U5km1AE7bWYQiuavXj3Yz6",
		},
	}
}

// DevelopmentConfig is the monero and ethereum development environment configuration
func DevelopmentConfig() *Config {
	return &Config{
		Env:     Development,
		DataDir: path.Join(baseDir, "dev"),
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
