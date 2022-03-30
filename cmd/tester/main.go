package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	// "errors"
	// "math/big"
	"os"
	"sync"
	"time"

	// ethcrypto "github.com/ethereum/go-ethereum/crypto"
	// "github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli"

	"github.com/noot/atomic-swap/common/rpcclient"
	"github.com/noot/atomic-swap/common/types"
	// mcrypto "github.com/noot/atomic-swap/crypto/monero"
	// "github.com/noot/atomic-swap/protocol/alice"
	// "github.com/noot/atomic-swap/protocol/bob"
	// recovery "github.com/noot/atomic-swap/recover"

	logging "github.com/ipfs/go-log"
)

const (
	flagEnv    = "env"
	flagConfig = "config"

	defaultConfigFile = "testerconfig.json"
)

var (
	log = logging.Logger("cmd")
	_   = logging.SetLogLevel("alice", "debug")
	_   = logging.SetLogLevel("bob", "debug")
	_   = logging.SetLogLevel("common", "debug")
	_   = logging.SetLogLevel("cmd", "debug")
	_   = logging.SetLogLevel("net", "debug")
	_   = logging.SetLogLevel("rpc", "debug")
)

var (
	app = &cli.App{
		Name:   "swaptester",
		Usage:  "A program for automatically testing swapd instances by performing many swaps",
		Action: runTester,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagEnv,
				Usage: "environment to use: one of mainnet, stagenet, or dev",
			},
			&cli.StringFlag{
				Name:  flagConfig,
				Usage: "path to configuration file containing swapd websockets endpoints",
			},
		},
	}
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func runTester(c *cli.Context) error {
	var defaultTimeout = time.Minute * 10
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	config := c.String(flagConfig)
	if config == "" {
		config = defaultConfigFile
	}

	bz, err := ioutil.ReadFile(config)
	if err != nil {
		return err
	}

	var endpoints []string
	if err := json.Unmarshal(bz, endpoints); err != nil {
		return fmt.Errorf("failed to unmarshal endpoints in config file: %w", err)
	}

	rsl := newResultLogger()

	var wg sync.WaitGroup
	wg.Add(len(endpoints))

	errChs := make([]chan error, len(endpoints))
	for i, endpoint := range endpoints {
		d := &daemon{
			rsl:      rsl,
			endpoint: endpoint,
			errCh:    errChs[i],
			wg:       wg,
			idx:      i,
		}
		go d.test(ctx)
	}

	wg.Wait()
	return nil
}

const (
	// XMR offer amounts
	minProvidesAmount = 0.01
	maxProvidesAmount = 0.1
	exchangeRate      = types.ExchangeRate(1) // TODO: vary this
)

type daemon struct {
	rsl      *resultLogger
	endpoint string
	wsc      rpcclient.WsClient
	errCh    chan error
	wg       sync.WaitGroup
	idx      int
}

func (d *daemon) test(ctx context.Context) {
	defer d.wg.Done()
	go d.logErrors(ctx)

	var err error
	d.wsc, err = rpcclient.NewWsClient(ctx, d.endpoint)
	if err != nil {
		d.errCh <- err
		return
	}

	for {
		d.makeOffer()
		if ctx.Err() != nil {
			return
		}
	}
}

func (d *daemon) logErrors(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-d.errCh:
			log.Errorf("endpoint %d: %w", d.idx, err)
		}
	}
}

func (d *daemon) makeOffer() {
	_, takenCh, statusCh, err := d.wsc.MakeOfferAndSubscribe(minProvidesAmount, maxProvidesAmount,
		exchangeRate)
	if err != nil {
		d.errCh <- err
		return
	}

	// TODO: also log swap duration

	select {
	case taken := <-takenCh:
		if taken == nil {
			return
		}
		// case <-time.After(testTimeout):
		// 	errCh <- errors.New("make offer subscription timed out")
	}

	for status := range statusCh {
		// fmt.Println("> maker got status:", status)
		if status.IsOngoing() {
			continue
		}

		if status != types.CompletedSuccess {
			d.errCh <- fmt.Errorf("swap did not refund successfully for maker: exit status was %s", status)
		}

		d.rsl.logMakerStatus(status)
		return
	}
}

type resultLogger struct {
	maker     map[types.Status]uint
	taker     map[types.Status]uint
	durations []time.Duration
}

func newResultLogger() *resultLogger {
	return &resultLogger{
		maker:     make(map[types.Status]uint),
		taker:     make(map[types.Status]uint),
		durations: []time.Duration{},
	}
}

func (l *resultLogger) logMakerStatus(s types.Status) {
	l.maker[s]++
}

func (l *resultLogger) logTakerStatus(s types.Status) {
	l.taker[s]++
}

func (l *resultLogger) logSwapDuration(duration time.Duration) {
	l.durations = append(l.durations, duration)
}
