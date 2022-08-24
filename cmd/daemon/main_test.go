package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func newTestContext(t *testing.T, description string, flags []string, values []interface{}) *cli.Context {
	require.Equal(t, len(flags), len(values))

	set := flag.NewFlagSet(description, 0)
	for i := range values {
		switch v := values[i].(type) {
		case bool:
			set.Bool(flags[i], v, "")
		case string:
			set.String(flags[i], v, "")
		case uint:
			set.Uint(flags[i], v, "")
		case int64:
			set.Int64(flags[i], v, "")
		case []string:
			set.Var(&cli.StringSlice{}, flags[i], "")
		default:
			t.Fatalf("unexpected cli value type: %T", values[i])
		}
	}

	ctx := cli.NewContext(app, set, nil)
	var (
		err error
		i   int
	)

	for i = range values {
		switch v := values[i].(type) {
		case bool:
			err = ctx.Set(flags[i], strconv.FormatBool(v))
		case string:
			err = ctx.Set(flags[i], values[i].(string))
		case uint:
			err = ctx.Set(flags[i], strconv.Itoa(int(values[i].(uint))))
		case int64:
			err = ctx.Set(flags[i], strconv.Itoa(int(values[i].(int64))))
		case []string:
			for _, str := range values[i].([]string) {
				err = ctx.Set(flags[i], str)
				require.NoError(t, err)
			}
		default:
			t.Fatalf("unexpected cli value type: %T", values[i])
		}
	}

	require.NoError(t, err, fmt.Sprintf("failed to set cli flag: %T, err: %s", flags[i], err))
	return ctx
}

func TestDaemon_DevXMRTaker(t *testing.T) {
	c := newTestContext(t,
		"test --dev-xmrtaker",
		[]string{flagDevXMRTaker, flagBasepath},
		[]interface{}{true, t.TempDir()},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := &daemon{
		ctx:    ctx,
		cancel: cancel,
	}

	err := d.make(c)
	require.NoError(t, err)
}

func TestDaemon_DevXMRMaker(t *testing.T) {
	c := newTestContext(t,
		"test --dev-xmrmaker",
		[]string{flagDevXMRMaker, flagDeploy, flagBasepath},
		[]interface{}{true, true, t.TempDir()},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := &daemon{
		ctx:    ctx,
		cancel: cancel,
	}

	err := d.make(c)
	require.NoError(t, err)
}
