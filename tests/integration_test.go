package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/noot/atomic-swap/cmd/client/client"
	"github.com/noot/atomic-swap/common"

	"github.com/stretchr/testify/require"
)

const (
	defaultAliceMultiaddr      = "/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2"
	defaultAliceTestLibp2pKey  = "alice.key"
	defaultAliceDaemonEndpoint = "http://localhost:5001"
	defaultBobDaemonEndpoint   = "http://localhost:5002"

	aliceProvideAmount = float64(33.3)
	bobProvideAmount   = float64(44.4)
)

func TestMain(m *testing.M) {
	cmd := exec.Command("../scripts/build.sh")
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func startSwapDaemon(t *testing.T, ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "../swapd", args...)
	err := cmd.Start()
	require.NoError(t, err)
	time.Sleep(time.Second * 5)
	return cmd
}

func startAlice(t *testing.T, ctx context.Context) *exec.Cmd {
	return startSwapDaemon(t, ctx, "--alice",
		"--amount", fmt.Sprintf("%v", aliceProvideAmount),
		"--libp2p-key", defaultAliceTestLibp2pKey,
	)
}

func startBob(t *testing.T, ctx context.Context) *exec.Cmd {
	return startSwapDaemon(t, ctx, "--bob",
		"--amount", fmt.Sprintf("%v", bobProvideAmount),
		"--bootnodes", defaultAliceMultiaddr,
		"--wallet-file", "test-wallet",
	)
}

// charlie doesn't provide any coin or participate in any swap.
// he is just a node running the p2p protocol.
func startCharlie(t *testing.T, ctx context.Context) *exec.Cmd {
	return startSwapDaemon(t, ctx, "--bootnodes", defaultAliceMultiaddr)
}

func startNodes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	aliceCmd := startAlice(t, ctx)
	bobCmd := startBob(t, ctx)
	charlieCmd := startCharlie(t, ctx)

	t.Cleanup(func() {
		cancel()
		_ = aliceCmd.Wait()
		_ = bobCmd.Wait()
		_ = charlieCmd.Wait()
	})
}

func TestStartAlice(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := startAlice(t, ctx)
	cancel()
	_ = cmd.Wait()
}

func TestStartBob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := startBob(t, ctx)
	cancel()
	_ = cmd.Wait()
}

func TestAlice_Discover(t *testing.T) {
	startNodes(t)
	c := client.NewClient(defaultAliceDaemonEndpoint)
	providers, err := c.Discover(common.ProvidesXMR, 3)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)
}

func TestBob_Discover(t *testing.T) {
	startNodes(t)
	c := client.NewClient(defaultBobDaemonEndpoint)
	providers, err := c.Discover(common.ProvidesETH, 3)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)
}

func TestAlice_Query(t *testing.T) {
	startNodes(t)
	c := client.NewClient(defaultAliceDaemonEndpoint)

	providers, err := c.Discover(common.ProvidesXMR, 3)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)

	resp, err := c.Query(providers[0][0])
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Provides))
	require.Equal(t, common.ProvidesXMR, resp.Provides[0])
	require.Equal(t, 1, len(resp.MaximumAmount))
	require.Equal(t, bobProvideAmount, resp.MaximumAmount[0])
}

func TestBob_Query(t *testing.T) {
	startNodes(t)
	c := client.NewClient(defaultBobDaemonEndpoint)

	providers, err := c.Discover(common.ProvidesETH, 3)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)

	resp, err := c.Query(providers[0][0])
	require.NoError(t, err)
	t.Log(resp)
	require.Equal(t, 1, len(resp.Provides))
	require.Equal(t, common.ProvidesETH, resp.Provides[0])
	require.Equal(t, 1, len(resp.MaximumAmount))
	require.Equal(t, aliceProvideAmount, resp.MaximumAmount[0])
}
