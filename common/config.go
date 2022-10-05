package common

import (
	"os"
	"path"

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

// Config contains constants that are defaults for various environments
type Config struct {
	DataDir          string
	MoneroDaemonHost string
	MoneroDaemonPort uint
	EthereumChainID  int64
	ContractAddress  ethcommon.Address
	Bootnodes        []string
}

// MainnetConfig is the mainnet ethereum and monero configuration
var MainnetConfig = Config{
	DataDir:          path.Join(baseDir, "mainnet"),
	MoneroDaemonHost: "127.0.0.1",
	MoneroDaemonPort: DefaultMoneroDaemonMainnetPort,
	EthereumChainID:  MainnetChainID,
}

// StagenetConfig is the monero stagenet and ethereum Gorli configuration
var StagenetConfig = Config{
	DataDir:          path.Join(baseDir, "stagenet"),
	MoneroDaemonHost: "node.sethforprivacy.com",
	MoneroDaemonPort: DefaultMoneroDaemonStagenetPort,
	EthereumChainID:  GorliChainID,
	ContractAddress:  ethcommon.HexToAddress("0x64e902cD8A29bBAefb9D4e2e3A24d8250C606ee7"),
	Bootnodes: []string{
		"/ip4/134.122.115.208/tcp/9900/p2p/12D3KooWDqCzbjexHEa8Rut7bzxHFpRMZyDRW1L6TGkL1KY24JH5",
		"/ip4/143.198.123.27/tcp/9900/p2p/12D3KooWSc4yFkPWBFmPToTMbhChH3FAgGH96DNzSg5fio1pQYoN",
		"/ip4/67.207.89.83/tcp/9900/p2p/12D3KooWLbfkLZZvvn8Lxs1KDU3u7gyvBk88ZNtJBbugytBr5RCG",
		"/ip4/134.122.115.208/tcp/9900/p2p/12D3KooWDqCzbjexHEa8Rut7bzxHFpRMZyDRW1L6TGkL1KY24JH5",
		"/ip4/164.92.103.160/tcp/9900/p2p/12D3KooWAZtRECEv7zN69zU1e7sPrHbMgfqFUn7QTLh1pKGiMuaM",
		"/ip4/164.92.103.159/tcp/9900/p2p/12D3KooWSNQF1eNyapxC2zA3jJExgLX7jWhEyw8B3k7zMW5ZRvQz",
		"/ip4/164.92.123.10/tcp/9900/p2p/12D3KooWG8z9fXVTB72XL8hQbahpfEjutREL9vbBQ4FzqtDKzTBu",
		"/ip4/161.35.110.210/tcp/9900/p2p/12D3KooWS8iKxqsGTiL3Yc1VaAfg99U5km1AE7bWYQiuavXj3Yz6",
	},
}

// DevelopmentConfig is the monero and ethereum development environment configuration
var DevelopmentConfig = Config{
	DataDir:          path.Join(baseDir, "dev"),
	MoneroDaemonPort: DefaultMoneroDaemonDevPort,
	EthereumChainID:  GanacheChainID,
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
