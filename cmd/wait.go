package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func wait(ctx context.Context) {
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)

		select {
		case <-sigc:
			fmt.Println("signal interrupt, shutting down...")
		case <-ctx.Done():
			fmt.Println("protocol complete, shutting down...")
		}

		os.Exit(0)
	}()

	wg.Wait()
}
