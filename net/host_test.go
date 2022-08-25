package net

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"

	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	logging.SetLogLevel("net", "debug")
	m.Run()
	os.Exit(0)
}

var defaultPort uint16 = 5009
var testID = types.Hash{99}

type mockHandler struct {
	id types.Hash
}

func (h *mockHandler) GetOffers() []*types.Offer {
	return []*types.Offer{}
}

func (h *mockHandler) HandleInitiateMessage(msg *SendKeysMessage) (s SwapState, resp Message, err error) {
	if (h.id != types.Hash{}) {
		return &mockSwapState{h.id}, &SendKeysMessage{}, nil
	}
	return &mockSwapState{}, &SendKeysMessage{}, nil
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

func (s *mockSwapState) HandleProtocolMessage(msg Message) (resp Message, done bool, err error) {
	return nil, false, nil
}

func (s *mockSwapState) Exit() error {
	return nil
}

func newHost(t *testing.T, port uint16) *host {
	cfg := &Config{
		Ctx:         context.Background(),
		Environment: common.Development,
		Basepath:    t.TempDir(),
		ChainID:     common.GanacheChainID,
		Port:        port,
		KeyFile:     path.Join(t.TempDir(), fmt.Sprintf("node-%d.key", port)),
		Bootnodes:   []string{},
		Handler:     &mockHandler{},
	}

	h, err := NewHost(cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = h.Stop()
		require.NoError(t, err)
	})
	return h
}

func TestNewHost(t *testing.T) {
	h := newHost(t, defaultPort)
	err := h.Start()
	require.NoError(t, err)
}
