package net

import (
	"context"
	"path"
	"testing"

	p2pnet "github.com/athanorlabs/go-p2p-net"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"
)

func init() {
	logging.SetLogLevel("net", "debug")
}

var testID = types.Hash{99}

type mockMakerHandler struct {
	t  *testing.T
	id types.Hash
}

func (h *mockMakerHandler) GetOffers() []*types.Offer {
	return []*types.Offer{}
}

func (h *mockMakerHandler) HandleInitiateMessage(msg *message.SendKeysMessage) (s SwapState, resp Message, err error) {
	if (h.id != types.Hash{}) {
		return &mockSwapState{h.id}, createSendKeysMessage(h.t), nil
	}
	return &mockSwapState{}, msg, nil
}

type mockTakerHandler struct {
	t *testing.T
}

func (h *mockTakerHandler) HandleRelayClaimRequest(_ *RelayClaimRequest) (*RelayClaimResponse, error) {
	return new(RelayClaimResponse), nil
}

type mockSwapState struct {
	id types.Hash
}

func (s *mockSwapState) ID() types.Hash {
	if (s.id != types.Hash{}) {
		return s.id
	}

	return testID
}

func (s *mockSwapState) HandleProtocolMessage(_ Message) error {
	return nil
}

func (s *mockSwapState) Exit() error {
	return nil
}

func basicTestConfig(t *testing.T) *p2pnet.Config {
	// t.TempDir() is unique on every call. Don't reuse this config with multiple hosts.
	tmpDir := t.TempDir()
	return &p2pnet.Config{
		Ctx:        context.Background(),
		DataDir:    tmpDir,
		Port:       0, // OS randomized libp2p port
		KeyFile:    path.Join(tmpDir, "node.key"),
		Bootnodes:  nil,
		ProtocolID: "/testid",
		ListenIP:   "127.0.0.1",
	}
}

func newHost(t *testing.T, cfg *p2pnet.Config) *Host {
	h, err := NewHost(cfg, true)
	require.NoError(t, err)
	h.SetHandlers(&mockMakerHandler{t: t}, &mockTakerHandler{t: t})
	t.Cleanup(func() {
		err = h.Stop()
		require.NoError(t, err)
	})
	return h
}
