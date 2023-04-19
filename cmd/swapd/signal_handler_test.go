// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"context"
	"testing"
)

func TestDaemon_signalHandler(_ *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go signalHandler(ctx, cancel)
}
