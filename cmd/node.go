package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/noot/atomic-swap/alice"
	"github.com/noot/atomic-swap/bob"
	"github.com/noot/atomic-swap/net"
)

type node struct {
	ctx    context.Context
	cancel context.CancelFunc
	amount uint
	alice  alice.Alice
	bob    bob.Bob
	host   net.Host
	outCh  chan<- *net.MessageInfo
	inCh   <-chan *net.MessageInfo
}

func (n *node) wait() {
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)

		select {
		case <-sigc:
			fmt.Println("signal interrupt, shutting down...")
			n.cancel()
		case <-n.ctx.Done():
			fmt.Println("protocol complete, shutting down...")
		}

		os.Exit(0)
	}()

	wg.Wait()
}
