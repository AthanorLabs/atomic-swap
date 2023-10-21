// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package common is for miscellaneous constants, types and interfaces used by many packages.
package common

// Daemon default ports and URLs
const (
	DefaultMoneroDaemonMainnetPort  = 18081
	DefaultMoneroDaemonDevPort      = DefaultMoneroDaemonMainnetPort
	DefaultMoneroDaemonStagenetPort = 38081
	DefaultGanacheEndpoint          = "http://127.0.0.1:8545"
	DefaultGanacheWSEndpoint        = "ws://127.0.0.1:8545"
	DefaultSwapdPort                = 5000
)

// Ganache deterministic ethereum private wallet keys for the maker and taker in dev environments.
//
//nolint:gosec // don't flag hardcoded credentials
const (
	DefaultPrivKeyXMRTaker = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d" // index 0
	DefaultPrivKeyXMRMaker = "6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1" // index 1
)

// Strings for formatting time.Time types
const (
	TimeFmtSecs  = "2006-01-02-15:04:05"
	TimeFmtNSecs = "2006-01-02-15:04:05.999999999"
)

// Ethereum chain IDs
const (
	MainnetChainID   = 1
	OpMainnetChainID = 10 // Optimism
	SepoliaChainID   = 11155111
	GanacheChainID   = 1337
	HardhatChainID   = 31337
)
