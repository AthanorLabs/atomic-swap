package net

import (
	"context"
	"testing"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"

	"github.com/stretchr/testify/require"
)

type mockHandler struct{}

func (h *mockHandler) GetOffers() []*types.Offer {
	return nil
}

func (h *mockHandler) HandleInitiateMessage(msg *SendKeysMessage) (s SwapState, resp Message, err error) {
	return nil, nil, nil
}

func newHost(t *testing.T) *host {
	cfg := &Config{
		Ctx:         context.Background(),
		Environment: common.Development,
		ChainID:     common.GanacheChainID,
		Port:        5001,
		KeyFile:     "/tmp/node.key",
		Bootnodes:   []string{},
		Handler:     &mockHandler{},
	}

	h, err := NewHost(cfg)
	require.NoError(t, err)
	return h
}

func TestNewHost(t *testing.T) {
	_ = newHost(t)
}
