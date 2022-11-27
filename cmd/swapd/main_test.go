package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/rpcclient"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func newTestContext(t *testing.T, description string, flags map[string]any) *cli.Context {
	// The only external program any test in this package calls is monero-wallet-rpc, so we
	// make monero-bin the only directory in our path.
	curDir, err := os.Getwd()
	require.NoError(t, err)
	projectRoot := path.Dir(path.Dir(curDir)) // 2 dirs up from cmd/swaprecover
	os.Setenv("PATH", path.Join(projectRoot, "monero-bin"))

	set := flag.NewFlagSet(description, 0)
	for flag, value := range flags {
		switch v := value.(type) {
		case bool:
			set.Bool(flag, v, "")
		case string:
			set.String(flag, v, "")
		case uint:
			set.Uint(flag, v, "")
		case int64:
			set.Int64(flag, v, "")
		case []string:
			set.Var(&cli.StringSlice{}, flag, "")
		default:
			t.Fatalf("unexpected cli value type: %T", value)
		}
	}

	ctx := cli.NewContext(app, set, nil)

	for flag, value := range flags {
		switch v := value.(type) {
		case bool, uint, int64, string:
			require.NoError(t, ctx.Set(flag, fmt.Sprintf("%v", v)))
		case []string:
			for _, str := range v {
				require.NoError(t, ctx.Set(flag, str))
			}
		default:
			t.Fatalf("unexpected cli value type: %T", value)
		}
	}

	return ctx
}

func TestDaemon_DevXMRTaker(t *testing.T) {
	c := newTestContext(t,
		"test --dev-xmrtaker",
		map[string]any{
			flagEnv:         "dev",
			flagDevXMRTaker: true,
			flagDataDir:     t.TempDir(),
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := &daemon{
		ctx:    ctx,
		cancel: cancel,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		err := d.make(c) // blocks on RPC server start
		require.ErrorIs(t, err, context.Canceled)
		wg.Done()
	}()
	time.Sleep(500 * time.Millisecond) // let the server start
	cancel()
	wg.Wait()
}

func TestDaemon_DevXMRMaker(t *testing.T) {
	c := newTestContext(t,
		"test --dev-xmrmaker",
		map[string]any{
			flagEnv:         "dev",
			flagDevXMRMaker: true,
			flagDeploy:      true,
			flagDataDir:     t.TempDir(),
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := &daemon{
		ctx:    ctx,
		cancel: cancel,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		err := d.make(c) // blocks on RPC server start
		require.ErrorIs(t, err, context.Canceled)
		wg.Done()
	}()
	time.Sleep(500 * time.Millisecond) // let the server start
	cancel()
	wg.Wait()
}

func Test_expandBootnodes(t *testing.T) {
	cliNodes := []string{
		" node1, node2 ,node3,node4 ",
		"node5",
		"\tnode6\n",
		"node7,node8",
	}
	expected := []string{
		"node1",
		"node2",
		"node3",
		"node4",
		"node5",
		"node6",
		"node7",
		"node8",
	}
	require.EqualValues(t, expected, expandBootnodes(cliNodes))
}

func TestDaemon_PersistOffers(t *testing.T) {
	defaultXMRMakerSwapdEndpoint := fmt.Sprintf("http://localhost:%d", defaultXMRMakerRPCPort)
	startupTimeout := time.Millisecond * 100

	datadir := t.TempDir()
	wc := monero.CreateWalletClientWithWalletDir(t, datadir)
	monero.MineMinXMRBalance(t, wc, common.MoneroToPiconero(1))

	c := newTestContext(t,
		"test --dev-xmrmaker",
		map[string]any{
			flagEnv:              "dev",
			flagDevXMRMaker:      true,
			flagDeploy:           true,
			flagDataDir:          datadir,
			flagMoneroWalletPath: path.Join(datadir, "test-wallet"),
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := newEmptyDaemon(ctx, cancel)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := d.make(c) // blocks on RPC server start
		require.ErrorIs(t, err, context.Canceled)
	}()

	<-d.startedCh
	time.Sleep(startupTimeout) // let the server start

	// make an offer
	client := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	balance, err := client.Balances()
	require.NoError(t, err)
	require.GreaterOrEqual(t, balance.PiconeroUnlockedBalance, common.MoneroToPiconero(1))

	offerID, err := client.MakeOffer(0.1, 1, float64(1), types.EthAssetETH, "", 0)
	require.NoError(t, err)

	// shut down daemon
	cancel()
	wg.Wait()

	err = d.stop()
	require.NoError(t, err)

	// restart daemon
	t.Log("restarting daemon")
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	d = newEmptyDaemon(ctx, cancel)
	defer func() {
		require.NoError(t, d.stop())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = d.make(c) // blocks on RPC server start
		require.ErrorIs(t, err, context.Canceled)
	}()

	<-d.startedCh
	time.Sleep(startupTimeout) // let the server start

	offers, err := client.GetOffers()
	require.NoError(t, err)
	require.Equal(t, 1, len(offers))
	require.Equal(t, offerID, offers[0].ID.String())
}
