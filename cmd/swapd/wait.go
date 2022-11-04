package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func signalHandler(ctx context.Context, cancel context.CancelFunc) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)

	select {
	case <-sigc:
		log.Info("Signal interrupt, shutting down...")
		cancel()
	case <-ctx.Done():
		log.Info("Protocol complete, shutting down...")
	}
}
