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
		// Hopefully, we'll exit our main() before this sleep ends, but if not we allow
		// the default signal behavior to kill us after this function exits.
		time.Sleep(1 * time.Second)
		signal.Stop(sigc)
	}()

	select {
	case s := <-sigc:
		log.Infof("Received signal %s(%d), shutting down...", s, s)
		cancel()
	case <-ctx.Done():
	}
}
