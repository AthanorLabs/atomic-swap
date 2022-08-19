package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"
	"github.com/athanorlabs/atomic-swap/rpcclient/wsclient"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/require"
)

const (
	testMultiaddr = "/ip4/192.168.0.102/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2"
)

var (
	testSwapID            = types.Hash{99}
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
	var offer types.Offer
	offerJSON := fmt.Sprintf(`{"ID":%q}`, testSwapID.String())
	if err := json.Unmarshal([]byte(offerJSON), &offer); err != nil {
		panic(err)
	}
	return &net.QueryResponse{Offers: []*types.Offer{&offer}}, nil
}
func (*mockNet) Initiate(who peer.AddrInfo, msg *net.SendKeysMessage, s common.SwapStateNet) error {
	return nil
}
func (*mockNet) CloseProtocolStream(types.Hash) {}

type mockSwapManager struct{}

func (*mockSwapManager) GetPastIDs() []types.Hash {
	return []types.Hash{}
}
func (*mockSwapManager) GetPastSwap(id types.Hash) *swap.Info {
	return &swap.Info{}
}
func (*mockSwapManager) GetOngoingSwap(id types.Hash) *swap.Info {
	statusCh := make(chan types.Status, 1)
	statusCh <- types.CompletedSuccess

	return swap.NewInfo(
		id,
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
func (*mockSwapManager) CompleteOngoingSwap(types.Hash) {}

type mockXMRTaker struct{}

func (*mockXMRTaker) Provides() types.ProvidesCoin {
	return types.ProvidesETH
}
func (*mockXMRTaker) SetGasPrice(gasPrice uint64) {}
func (*mockXMRTaker) GetOngoingSwapState(types.Hash) common.SwapState {
	return new(mockSwapState)
}
func (*mockXMRTaker) InitiateProtocol(providesAmount float64, _ *types.Offer) (common.SwapState, error) {
	return new(mockSwapState), nil
}
func (*mockXMRTaker) Refund(types.Hash) (ethcommon.Hash, error) {
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
func (*mockSwapState) ID() types.Hash {
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
func (*mockProtocolBackend) ExternalSender() *txsender.ExternalSender {
	return nil
}
func (*mockProtocolBackend) SetEthAddress(ethcommon.Address)                  {}
func (*mockProtocolBackend) SetXMRDepositAddress(mcrypto.Address, types.Hash) {}

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

	ch, err := c.TakeOfferAndSubscribe(testMultiaddr, testSwapID.String(), 1)
	require.NoError(t, err)

	select {
	case status := <-ch:
		require.Equal(t, types.CompletedSuccess, status)
	case <-time.After(testTImeout):
		t.Fatal("test timed out")
	}
}
