package net

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/tests"

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
	ethCli, chainID := tests.NewEthClient(t)
	cfg := &Config{
		Ctx:         context.Background(),
		Environment: common.Development,
		DataDir:     t.TempDir(),
		EthChainID:  chainID.Int64(),
		Port:        port,
		KeyFile:     path.Join(t.TempDir(), fmt.Sprintf("node-%d.key", port)),
		Bootnodes:   []string{},
		Handler:     &mockHandler{},
		EthAddress:  common.EthereumPrivateKeyToAddress(tests.GetTakerTestKey(t)),
		EthCli:      ethCli,
		MoneroCli:   monero.CreateWalletClient(t),
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

func TestGetBalances(t *testing.T) {
	h := newHost(t, defaultPort)
	balances, err := h.Balances()
	require.NoError(t, err)
	require.NotEmpty(t, balances.MoneroAddress)
	require.NotEmpty(t, balances.EthAddress)
	require.Greater(t, balances.EthBalance.Int64(), int64(0))
}
