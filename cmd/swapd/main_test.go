package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func newTestContext(t *testing.T, description string, flags map[string]any) *cli.Context {
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
