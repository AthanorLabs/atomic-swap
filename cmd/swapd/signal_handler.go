// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func signalHandler(ctx context.Context, cancel context.CancelFunc) {
	sigc := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		// Hopefully, we'll exit our main() before this sleep ends, but if not remove the
		// signal handler allowing the next signal to kill us.
		time.Sleep(1 * time.Second)
		signal.Stop(sigc)
	}()

	select {
	case s := <-sigc:
		log.Infof("Received signal %s(%d), shutting down...", s, s)
		cancel()
	case <-ctx.Done():
		log.Info("Protocol complete, shutting down...")
	}
}
