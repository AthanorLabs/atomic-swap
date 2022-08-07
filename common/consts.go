package common

const (
	MainnetChainID = 1 //nolint
	RopstenChainID = 3
	GanacheChainID = 1337

	DefaultXMRTakerMoneroEndpoint = "http://127.0.0.1:18084/json_rpc"
	DefaultXMRMakerMoneroEndpoint = "http://127.0.0.1:18083/json_rpc"
	DefaultMoneroDaemonEndpoint   = "http://127.0.0.1:18081/json_rpc"
	DefaultEthEndpoint            = "ws://localhost:8545"

	// DefaultPrivKeyXMRTaker is the private key at index 0 from `ganache --deterministic`
	DefaultPrivKeyXMRTaker = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"

	// DefaultPrivKeyXMRMaker is the private key at index 1 from `ganache --deterministic`
	DefaultPrivKeyXMRMaker = "6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1"

	TimeFmtSecs  = "2006-01-02-15:04:05"
	TimeFmtNSecs = "2006-01-02-15:04:05.999999999"
)
