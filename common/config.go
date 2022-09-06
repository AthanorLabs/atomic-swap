package common

import (
	"fmt"
	"os"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

var homeDir, _ = os.UserHomeDir()

// Config contains constants that are defaults for various environments
type Config struct {
	DataDir              string
	MoneroDaemonEndpoint string
	EthereumChainID      int64
	ContractAddress      ethcommon.Address
	Bootnodes            []string
}

// MainnetConfig is the mainnet ethereum and monero configuration
var MainnetConfig = Config{
	DataDir:              fmt.Sprintf("%s/.atomicswap/mainnet", homeDir),
	MoneroDaemonEndpoint: "http://127.0.0.1:18081/json_rpc",
	EthereumChainID:      MainnetChainID,
}

// StagenetConfig is the monero stagenet and ethereum ropsten configuration
var StagenetConfig = Config{
	DataDir:              fmt.Sprintf("%s/.atomicswap/stagenet", homeDir),
	MoneroDaemonEndpoint: "http://127.0.0.1:38081/json_rpc",
	EthereumChainID:      GorliChainID,
	ContractAddress:      ethcommon.HexToAddress("0x2125320230096B33b55f6d7905Fef61A3a0906a0"),
	Bootnodes: []string{
		"/ip4/134.122.115.208/tcp/9900/p2p/12D3KooWDqCzbjexHEa8Rut7bzxHFpRMZyDRW1L6TGkL1KY24JH5",
		"/ip4/143.198.123.27/tcp/9900/p2p/12D3KooWSc4yFkPWBFmPToTMbhChH3FAgGH96DNzSg5fio1pQYoN",
		"/ip4/67.207.89.83/tcp/9900/p2p/12D3KooWLbfkLZZvvn8Lxs1KDU3u7gyvBk88ZNtJBbugytBr5RCG",
		"/ip4/134.122.115.208/tcp/9900/p2p/12D3KooWDqCzbjexHEa8Rut7bzxHFpRMZyDRW1L6TGkL1KY24JH5",
		"/ip4/164.92.103.160/tcp/9900/p2p/12D3KooWAZtRECEv7zN69zU1e7sPrHbMgfqFUn7QTLh1pKGiMuaM",
		"/ip4/164.92.103.159/tcp/9900/p2p/12D3KooWSNQF1eNyapxC2zA3jJExgLX7jWhEyw8B3k7zMW5ZRvQz",
		"/ip4/164.92.123.10/tcp/9900/p2p/12D3KooWG8z9fXVTB72XL8hQbahpfEjutREL9vbBQ4FzqtDKzTBu",
		"/ip4/161.35.110.210/tcp/9900/p2p/12D3KooWS8iKxqsGTiL3Yc1VaAfg99U5km1AE7bWYQiuavXj3Yz6",
		"/ip4/206.189.47.220/tcp/9900/p2p/12D3KooWGVzz2d2LSceVFFdqTYqmQXTqc5eWziw7PLRahCWGJhKB",
	},
}

// DevelopmentConfig is the monero and ethereum development environment configuration
var DevelopmentConfig = Config{
	DataDir:              fmt.Sprintf("%s/.atomicswap/dev", homeDir),
	MoneroDaemonEndpoint: "http://127.0.0.1:18081/json_rpc",
	EthereumChainID:      GanacheChainID,
}
