// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package cliutil

import (
	"fmt"

	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli/v2"
)

const (
	// FlagLogLevel is the log level flag.
	FlagLogLevel = "log-level"
)

// SetLogLevelsFromContext sets the log levels for all packages from the CLI context.
func SetLogLevelsFromContext(c *cli.Context) error {
	const (
		levelError = "error"
		levelWarn  = "warn"
		levelInfo  = "info"
		levelDebug = "debug"
	)

	level := c.String(FlagLogLevel)
	switch level {
	case levelError, levelWarn, levelInfo, levelDebug:
	default:
		return fmt.Errorf("invalid log level %q", level)
	}

	setLogLevels(level)
	return nil
}

func setLogLevels(level string) {
	// alphabetically ordered
	_ = logging.SetLogLevel("cmd", level)
	_ = logging.SetLogLevel("coins", level)
	_ = logging.SetLogLevel("common", level)
	_ = logging.SetLogLevel("contracts", level)
	_ = logging.SetLogLevel("cmd", level)
	_ = logging.SetLogLevel("extethclient", level)
	_ = logging.SetLogLevel("ethereum/watcher", level)
	_ = logging.SetLogLevel("ethereum/block", level)
	_ = logging.SetLogLevel("monero", level)
	_ = logging.SetLogLevel("net", level)
	_ = logging.SetLogLevel("offers", level)
	_ = logging.SetLogLevel("p2pnet", level) // external
	_ = logging.SetLogLevel("pricefeed", level)
	_ = logging.SetLogLevel("protocol", level)
	_ = logging.SetLogLevel("relayer", level) // external and internal
	_ = logging.SetLogLevel("rpc", level)
	_ = logging.SetLogLevel("txsender", level)
	_ = logging.SetLogLevel("xmrmaker", level)
	_ = logging.SetLogLevel("xmrtaker", level)
}
