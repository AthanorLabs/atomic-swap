package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func (d *daemon) wait() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)

	select {
	case <-sigc:
		fmt.Println("signal interrupt, shutting down...")
		d.cancel()
	case <-d.ctx.Done():
		fmt.Println("protocol complete, shutting down...")
	}
}
