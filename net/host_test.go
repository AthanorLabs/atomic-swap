package net

import (
	"context"
	"fmt"
	"testing"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"

	"github.com/stretchr/testify/require"
)

var defaultPort uint16 = 5001

type mockHandler struct{}

func (h *mockHandler) GetOffers() []*types.Offer {
	return []*types.Offer{}
}

func (h *mockHandler) HandleInitiateMessage(msg *SendKeysMessage) (s SwapState, resp Message, err error) {
	return nil, &SendKeysMessage{}, nil
}

func newHost(t *testing.T, port uint16) *host {
	cfg := &Config{
		Ctx:         context.Background(),
		Environment: common.Development,
		ChainID:     common.GanacheChainID,
		Port:        port,
		KeyFile:     fmt.Sprintf("/tmp/node-%d.key", port),
		Bootnodes:   []string{},
		Handler:     &mockHandler{},
	}

	h, err := NewHost(cfg)
	require.NoError(t, err)
	return h
}

func TestNewHost(t *testing.T) {
	h := newHost(t, defaultPort)
	err := h.Start()
	require.NoError(t, err)
	err = h.Stop()
	require.NoError(t, err)
}
