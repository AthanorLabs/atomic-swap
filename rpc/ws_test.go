package rpc

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/net/message"
	"github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/rpcclient/wsclient"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/require"
)

const (
	testSwapID    uint64 = 77
	testMultiaddr        = "/ip4/192.168.0.102/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2"
)

var (
	testTImeout           = time.Second * 5
	defaultRPCPort uint16 = 3001
	defaultWSPort  uint16 = 4002
)

func defaultWSEndpoint() string {
	return fmt.Sprintf("ws://localhost:%d", defaultWSPort)
}

type mockNet struct{}

func (*mockNet) Addresses() []string {
	return nil
}
func (*mockNet) Advertise() {}
func (*mockNet) Discover(provides types.ProvidesCoin, searchTime time.Duration) ([]peer.AddrInfo, error) {
	return nil, nil
}
func (*mockNet) Query(who peer.AddrInfo) (*net.QueryResponse, error) {
	return &net.QueryResponse{
		Offers: []*types.Offer{
			{},
		},
	}, nil
}
func (*mockNet) Initiate(who peer.AddrInfo, msg *net.SendKeysMessage, s common.SwapState) error {
	return nil
}
func (*mockNet) CloseProtocolStream() {}

type mockSwapManager struct{}

func (*mockSwapManager) GetPastIDs() []uint64 {
	return []uint64{}
}
func (*mockSwapManager) GetPastSwap(id uint64) *swap.Info {
	return &swap.Info{}
}
func (*mockSwapManager) GetOngoingSwap() *swap.Info {
	statusCh := make(chan types.Status, 1)
	statusCh <- types.CompletedSuccess

	return swap.NewInfo(
		types.ProvidesETH,
		1,
		1,
		1,
		types.CompletedSuccess,
		statusCh,
	)
}
func (*mockSwapManager) AddSwap(*swap.Info) error {
	return nil
}
func (*mockSwapManager) CompleteOngoingSwap() {}

type mockXMRTaker struct{}

func (*mockXMRTaker) Provides() types.ProvidesCoin {
	return types.ProvidesETH
}
func (*mockXMRTaker) SetGasPrice(gasPrice uint64) {}
func (*mockXMRTaker) GetOngoingSwapState() common.SwapState {
	return new(mockSwapState)
}
func (*mockXMRTaker) InitiateProtocol(providesAmount float64, _ *types.Offer) (common.SwapState, error) {
	return new(mockSwapState), nil
}
func (*mockXMRTaker) Refund() (ethcommon.Hash, error) {
	return ethcommon.Hash{}, nil
}
func (*mockXMRTaker) SetSwapTimeout(_ time.Duration) {}

type mockSwapState struct{}

func (*mockSwapState) HandleProtocolMessage(msg message.Message) (resp message.Message, done bool, err error) {
	return nil, true, nil
}
func (*mockSwapState) Exit() error {
	return nil
}
func (*mockSwapState) SendKeysMessage() (*message.SendKeysMessage, error) {
	return &message.SendKeysMessage{}, nil
}
func (*mockSwapState) ID() uint64 {
	return testSwapID
}
func (*mockSwapState) InfoFile() string {
	return os.TempDir() + "test.infofile"
}

type mockProtocolBackend struct {
	sm *mockSwapManager
}

func newMockProtocolBackend() *mockProtocolBackend {
	return &mockProtocolBackend{
		sm: new(mockSwapManager),
	}
}

func (*mockProtocolBackend) SetGasPrice(uint64)                   {}
func (*mockProtocolBackend) SetSwapTimeout(timeout time.Duration) {}
func (b *mockProtocolBackend) SwapManager() swap.Manager {
	return b.sm
}

func newServer(t *testing.T) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})

	defaultRPCPort++
	defaultWSPort++

	cfg := &Config{
		Ctx:             ctx,
		Port:            defaultRPCPort,
		WsPort:          defaultWSPort,
		Net:             new(mockNet),
		ProtocolBackend: newMockProtocolBackend(),
		XMRTaker:        new(mockXMRTaker),
	}

	s, err := NewServer(cfg)
	require.NoError(t, err)
	errCh := s.Start()
	go func() {
		err := <-errCh
		require.NoError(t, err)
	}()
	time.Sleep(time.Millisecond * 300) // let server start up

	return s
}

func TestSubscribeSwapStatus(t *testing.T) {
	_ = newServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})
	c, err := wsclient.NewWsClient(ctx, defaultWSEndpoint())
	require.NoError(t, err)

	ch, err := c.SubscribeSwapStatus(testSwapID)
	require.NoError(t, err)

	select {
	case status := <-ch:
		require.Equal(t, types.CompletedSuccess, status)
	case <-time.After(testTImeout):
		t.Fatal("test timed out")
	}
}

// TODO: add unit test
// func TestSubscribeMakeOffer(t *testing.T) {
// 	_ = newServer(t)

// 	ctx, cancel := context.WithCancel(context.Background())
// 	t.Cleanup(func() {
// 		cancel()
// 	})
// 	c, err := rpcclient.NewWsClient(ctx, defaultWSEndpoint())
// 	require.NoError(t, err)

// 	id, ch, err := c.MakeOfferAndSubscribe(0.1, 1, 0.05)
// 	require.NoError(t, err)
// 	require.Equal(t, id, testSwapID)
// 	select {
// 	case status := <-ch:
// 		require.Equal(t, types.CompletedSuccess, status)
// 	case <-time.After(testTImeout):
// 		t.Fatal("test timed out")
// 	}
// }

func TestSubscribeTakeOffer(t *testing.T) {
	_ = newServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})
	c, err := wsclient.NewWsClient(ctx, defaultWSEndpoint())
	require.NoError(t, err)

	offerID := (&types.Offer{}).GetID()

	id, ch, err := c.TakeOfferAndSubscribe(testMultiaddr, offerID.String(), 1)
	require.NoError(t, err)
	require.Equal(t, id, testSwapID)
	select {
	case status := <-ch:
		require.Equal(t, types.CompletedSuccess, status)
	case <-time.After(testTImeout):
		t.Fatal("test timed out")
	}
}
