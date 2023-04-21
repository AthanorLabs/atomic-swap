// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package cliutil

import (
	"context"
	"testing"

	logging "github.com/ipfs/go-log"
)

func TestDaemon_signalHandler(_ *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go SignalHandler(ctx, cancel, logging.Logger("test"))
}
