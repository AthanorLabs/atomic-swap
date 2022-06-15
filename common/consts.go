package common

const (
	MainnetChainID = 1 //nolint
	RopstenChainID = 3
	GanacheChainID = 1337

	DefaultXMRTakerMoneroEndpoint = "http://127.0.0.1:18084/json_rpc"
	DefaultXMRMakerMoneroEndpoint = "http://127.0.0.1:18083/json_rpc"
	DefaultMoneroDaemonEndpoint   = "http://127.0.0.1:18081/json_rpc"
	DefaultEthEndpoint            = "ws://localhost:8545"

	// `ganache-cli --deterministic` provides the accounts associated with the
	// private keys below 100 ETH on startup.
	// (0) 0x4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d
	// (1) 0x6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1
	// (2) 0x6370fd033278c143179d81c5526140625662b8daa446c22ee2d73db3707e620c
	// (3) 0x646f1ce2fdad0e6deeeb5c7e8e5543bdde65e86029e2fd9fc169899c440a7913
	// (4) 0xadd53f9a7e588d003326d1cbf9e4a43c061aadd9bc938c843a79e7b4fd2ad743
	// (5) 0x395df67f0c2d2d9fe1ad08d1bc8b6627011959b79c53d7dd6a3536a33ab8a4fd
	// (6) 0xe485d098507f54e7733a205420dfddbe58db035fa577fc294ebd14db90767a52
	// (7) 0xa453611d9419d0e56f499079478fd72c37b251a94bfde4d19872c44cf65386e3
	// (8) 0x829e924fdf021ba3dbbc4225edfece9aca04b929d6e75613329ca6f1d31c0bb4
	// (9) 0xb0057716d5917badaf911b193b12b910811c1497b5bada8d7711f758981c3773

	// TestPrivKeyXMRTaker (ganache key #0) is the Ethereum private key for swapd taker operations
	TestPrivKeyXMRTaker = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"

	// TestPrivKeyXMRMaker (ganache key #1) is the Ethereum private key for swapd maker operations
	TestPrivKeyXMRMaker = "6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1"

	// TestPrivKeySwapFactory (ganache key #2) is used by the daemon package unit tests
	TestPrivKeySwapFactory = "6370fd033278c143179d81c5526140625662b8daa446c22ee2d73db3707e620c"
)
