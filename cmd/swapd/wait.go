package main

import (
	"context"
	"fmt"
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
		fmt.Println("signal interrupt, shutting down...")
		cancel()
	case <-ctx.Done():
		fmt.Println("protocol complete, shutting down...")
	}
}
