// Package common is for miscellaneous constants, types and interfaces used by many packages.
package common

import "github.com/cockroachdb/apd/v3"

const (
	DefaultMoneroDaemonMainnetPort  = 18081 //nolint
	DefaultMoneroDaemonDevPort      = DefaultMoneroDaemonMainnetPort
	DefaultMoneroDaemonStagenetPort = 38081
	DefaultEthEndpoint              = "ws://127.0.0.1:8545"
	DefaultSwapdPort                = 5000

	// DefaultPrivKeyXMRTaker is the private key at index 0 from `ganache --deterministic`
	DefaultPrivKeyXMRTaker = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"

	// DefaultPrivKeyXMRMaker is the private key at index 1 from `ganache --deterministic`
	DefaultPrivKeyXMRMaker = "6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1"

	TimeFmtSecs  = "2006-01-02-15:04:05"
	TimeFmtNSecs = "2006-01-02-15:04:05.999999999"

	//nolint:revive
	// SwapFactory.sol function and event signatures
	ReadyEventSignature    = "Ready(bytes32)"
	ClaimedEventSignature  = "Claimed(bytes32,bytes32)"
	RefundedEventSignature = "Refunded(bytes32,bytes32)"

	//nolint:revive
	// Ethereum chain IDs
	MainnetChainID = 1
	GoerliChainID  = 5
	GanacheChainID = 1337
	HardhatChainID = 31337
)

// DefaultRelayerCommission is the default commission percentage for swap relayers.
// It's set to 0.01 or 1%.
var DefaultRelayerCommission = apd.New(1, -2)
