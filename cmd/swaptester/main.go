// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package main provides the entrypoint of swaptester, an executable used for
// automatically testing multiple swaps.
package main

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	mrand "math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MarinX/monerorpc"
	monerodaemon "github.com/MarinX/monerorpc/daemon"
	"github.com/cockroachdb/apd/v3"
	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli/v2"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/rpcclient/wsclient"
)

const (
	flagConfig   = "config"
	flagTimeout  = "timeout"
	flagLogLevel = "log-level"
	flagDev      = "dev"

	defaultConfigFile               = "testerconfig.json"
	defaultXMRMakerMoneroWalletPort = 18083
)

var (
	defaultTimeout       = time.Minute * 15
	log                  = logging.Logger("cmd")
	isDev                = false
	defaultMoneroClient  monero.WalletClient
	moneroDaemonEndpoint = fmt.Sprintf("http://127.0.0.1:%d/json_rpc", common.DefaultMoneroDaemonDevPort)
)

var (
	app = &cli.App{
		Name:    "swaptester",
		Usage:   "A program for automatically testing swapd instances by performing many swaps",
		Version: cliutil.GetVersion(),
		Action:  runTester,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagConfig,
				Usage: "path to configuration file containing swapd websockets endpoints",
			},
			&cli.UintFlag{
				Name:  flagTimeout,
				Usage: "time for which to run tester, in minutes; default=15mins",
			},
			&cli.StringFlag{
				Name:  flagLogLevel,
				Usage: "set log level: one of [error|warn|info|debug]",
				Value: "info",
			},
			&cli.BoolFlag{
				Name:  flagDev,
				Usage: "run tester in development environment",
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
	err := cliutil.SetLogLevelsFromContext(c)
	if err != nil {
		return err
	}

	isDev = c.Bool(flagDev)
	if !isDev {
		// TODO: Why do this when dev flag is not given? Can the code work if it is given?
		// For this to work, you'll need to pass --wallet-port 18083 to the XMR Maker swapd
		defaultMoneroClient = monero.NewThinWalletClient(
			"127.0.0.1",
			common.DefaultMoneroDaemonDevPort,
			defaultXMRMakerMoneroWalletPort,
		)
	}

	var timeout time.Duration

	timeoutMins := c.Uint(flagTimeout)
	if timeoutMins == 0 {
		timeout = defaultTimeout
	} else {
		timeout = time.Minute * time.Duration(timeoutMins)
	}

	log.Infof("starting to test, total duration is %vmins", timeout.Minutes())

	timer := time.After(timeout)
	done := make(chan struct{})

	go func() {
		<-timer
		close(done)
	}()

	config := c.String(flagConfig)
	if config == "" {
		config = defaultConfigFile
	}

	bz, err := os.ReadFile(filepath.Clean(config))
	if err != nil {
		return err
	}

	var endpoints []string
	if err := json.Unmarshal(bz, &endpoints); err != nil {
		return fmt.Errorf("failed to unmarshal endpoints in config file: %w", err)
	}

	rsl := newResultLogger()
	defer rsl.printStats()

	var wg sync.WaitGroup
	wg.Add(len(endpoints))

	errChs := make([]chan error, len(endpoints))
	for i, endpoint := range endpoints {
		errChs[i] = make(chan error, 16)

		d := &daemon{
			rsl:      rsl,
			endpoint: endpoint,
			errCh:    errChs[i],
			wg:       &wg,
			idx:      i,
			stop:     make(chan struct{}),
		}
		go d.test(done)
	}

	wg.Wait()
	return nil
}

const (
	// XMR offer amounts
	minProvidesAmount = 0.01
	maxProvidesAmount = 0.1
	minExchangeRate   = 0.9
	maxExchangeRate   = 1.1
)

var (
	minProvidesAmountBD, _, _ = apd.NewFromString(fmt.Sprintf("%f", minProvidesAmount))
	maxProvidesAmountBD, _, _ = apd.NewFromString(fmt.Sprintf("%f", maxProvidesAmount))
)

func getRandomExchangeRate() *coins.ExchangeRate {
	rateFl := minExchangeRate + mrand.Float64()*(maxExchangeRate-minExchangeRate) //nolint:gosec
	rate, _, err := new(apd.Decimal).SetString(strconv.FormatFloat(rateFl, 'f', -1, 64))
	if err != nil {
		panic(err) // shouldn't be possible
	}
	return coins.ToExchangeRate(rate)
}

func getRndOfferAmount(xRate *coins.ExchangeRate, minXMRAmt, maxXMRAmt *apd.Decimal) (*apd.Decimal, error) {
	randVal, err := new(apd.Decimal).SetFloat64(mrand.Float64()) //nolint:gosec
	if err != nil {
		return nil, err
	}
	minETHAmt, err := xRate.ToETH(minXMRAmt)
	if err != nil {
		return nil, err
	}
	maxETHAmt, err := xRate.ToETH(maxXMRAmt)
	if err != nil {
		return nil, err
	}
	ed := apd.MakeErrDecimal(apd.BaseContext.WithPrecision(coins.MaxCoinPrecision))
	rangeLen := ed.Sub(new(apd.Decimal), maxETHAmt, minETHAmt)
	randDelta := ed.Mul(new(apd.Decimal), rangeLen, randVal)
	offerAmt := ed.Add(new(apd.Decimal), minETHAmt, randDelta)
	if ed.Err() != nil {
		return nil, ed.Err()
	}
	return offerAmt, nil
}

func generateBlocks() {
	cXMRMaker := defaultMoneroClient
	xmrmakerAddr, err := cXMRMaker.GetAddress(0)
	if err != nil {
		log.Errorf("failed to get default monero address: %s", err)
		return
	}
	log.Infof("development: generating blocks...")
	daemonCli := monerorpc.New(moneroDaemonEndpoint, nil).Daemon
	for i := 0; i < 128; i += 32 {
		_, err = daemonCli.GenerateBlocks(&monerodaemon.GenerateBlocksRequest{
			AmountOfBlocks: 32,
			WalletAddress:  xmrmakerAddr.Address,
		})
		if err != nil {
			log.Warnf("Error generating blocks: %s", err)
		}
	}
}

type daemon struct {
	rsl      *resultLogger
	endpoint string
	errCh    chan error
	wg       *sync.WaitGroup
	idx      int
	stop     chan struct{}
	swapMu   sync.Mutex
}

func (d *daemon) test(done <-chan struct{}) {
	log.Infof("starting tester for node %s at index %d...", d.endpoint, d.idx)

	defer d.wg.Done()
	go d.logErrors(done)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		var sleep int

		for {
			select {
			case <-time.After(time.Second * time.Duration(sleep)):
				d.makeOffer(done)
			case <-done:
				return
			case <-d.stop:
				return
			}

			sleep = getRandomInt(60) + 3
		}
	}()

	go func() {
		defer wg.Done()

		for {
			sleep := getRandomInt(60) + 3

			select {
			case <-time.After(time.Second * time.Duration(sleep)):
				d.takeOffer(done)
			case <-done:
				return
			case <-d.stop:
				return
			}
		}
	}()

	wg.Wait()
	log.Warnf("node %d returning from d.test", d.idx)
}

func (d *daemon) logErrors(done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case err := <-d.errCh:
			log.Errorf("endpoint %d: %s", d.idx, err)
			if strings.Contains(err.Error(), "connection refused") {
				close(d.stop)
			}
		}
	}
}

func (d *daemon) takeOffer(done <-chan struct{}) {
	log.Debugf("node %d discovering offers...", d.idx)
	wsc, err := wsclient.NewWsClient(context.Background(), d.endpoint)
	if err != nil {
		d.errCh <- err
		return
	}

	defer wsc.Close()

	const defaultDiscoverTimeout = uint64(3) // 3s
	peerIDs, err := wsc.Discover(string(coins.ProvidesXMR), defaultDiscoverTimeout)
	if err != nil {
		d.errCh <- err
		return
	}

	if len(peerIDs) == 0 {
		return
	}

	makerIdx := getRandomInt(len(peerIDs))
	peerID := peerIDs[makerIdx]

	log.Debugf("node %d querying peer %s...", d.idx, peerID)

	resp, err := wsc.Query(peerID)
	if err != nil {
		d.errCh <- err
		return
	}

	if len(resp.Offers) == 0 {
		return
	}

	offerIdx := getRandomInt(len(resp.Offers))
	offer := resp.Offers[offerIdx]

	// pick a random ETH amount in the allowed range
	providesAmount, err := getRndOfferAmount(offer.ExchangeRate, offer.MinAmount, offer.MaxAmount)
	if err != nil {
		d.errCh <- err
		return
	}

	start := time.Now()
	log.Infof("node %d taking offer %s", d.idx, offer.ID)

	takerStatusCh, err := wsc.TakeOfferAndSubscribe(peerID, offer.ID, providesAmount)
	if err != nil {
		d.errCh <- err
		return
	}

	d.swapMu.Lock()
	defer d.swapMu.Unlock()

	for {
		select {
		case <-done:
			return
		case status := <-takerStatusCh:
			log.Infof("> taker (node %d) got status: %s", d.idx, status)
			if status.IsOngoing() {
				continue
			}

			if status != types.CompletedSuccess {
				d.errCh <- fmt.Errorf("swap did not complete successfully for taker: got %s", status)
			}

			d.rsl.logTakerStatus(status)

			if status != types.CompletedAbort {
				d.rsl.logSwapDuration(time.Since(start))
			}

			return
		}
	}
}

func getRandomInt(max int) int {
	i, err := crand.Int(crand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(err)
	}
	return int(i.Int64())
}

func (d *daemon) makeOffer(done <-chan struct{}) {
	log.Infof("node %d making offer...", d.idx)
	wsc, err := wsclient.NewWsClient(context.Background(), d.endpoint)
	if err != nil {
		d.errCh <- err
		return
	}

	defer wsc.Close()

	useRelayer := getRandomInt(2) == 0
	offerID, statusCh, err := wsc.MakeOfferAndSubscribe(minProvidesAmountBD,
		maxProvidesAmountBD,
		getRandomExchangeRate(),
		types.EthAssetETH,
		useRelayer,
	)
	if err != nil {
		log.Errorf("failed to make offer (node %d): %s", d.idx, err)
		d.errCh <- err

		if strings.Contains(err.Error(), "unlocked balance is less than maximum") {
			if isDev {
				generateBlocks()
			} else {
				_, err := monero.WaitForBlocks(context.Background(), defaultMoneroClient, 10)
				if err != nil {
					log.Errorf("failed to wait for blocks: %s", err)
				}
			}
		}

		return
	}

	log.Infof("node %d made offer %s", d.idx, offerID)

	d.swapMu.Lock()
	defer d.swapMu.Unlock()

	start := time.Now()

	for {
		select {
		case <-done:
			return
		case status := <-statusCh:
			log.Infof("> maker (node %d) got status: %s", d.idx, status)
			if status.IsOngoing() {
				continue
			}

			if status != types.CompletedSuccess {
				d.errCh <- fmt.Errorf("swap did not complete successfully for maker: exit status was %s", status)
			}

			d.rsl.logMakerStatus(status)
			if status != types.CompletedAbort {
				d.rsl.logSwapDuration(time.Since(start))
			}

			return
		}
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

func (l *resultLogger) averageDuration() time.Duration {
	if len(l.durations) == 0 {
		return 0
	}

	sum := time.Duration(0)
	for _, dur := range l.durations {
		sum += dur
	}
	return sum / time.Duration(len(l.durations))
}

func (l *resultLogger) printStats() {
	log.Infof("> total swaps=%d", len(l.durations))
	log.Infof("> [maker] aborted %d | refunded %d | success %d",
		l.maker[types.CompletedAbort],
		l.maker[types.CompletedRefund],
		l.maker[types.CompletedSuccess],
	)
	log.Infof("> [taker] aborted %d | refunded %d | success %d",
		l.taker[types.CompletedAbort],
		l.taker[types.CompletedRefund],
		l.taker[types.CompletedSuccess],
	)
	log.Infof("> average swap duration: %dms", l.averageDuration().Milliseconds())
}
