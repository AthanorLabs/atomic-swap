package host

import (
	"context"
	"path"
	"testing"

	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/tests"
)

func init() {
	logging.SetLogLevel("net", "debug")
}

var testID = types.Hash{99}

type mockHandler struct {
	id types.Hash
}

func (h *mockHandler) GetOffers() []*types.Offer {
	return []*types.Offer{}
}

func (h *mockHandler) HandleInitiateMessage(_ *message.SendKeysMessage) (s SwapState, resp Message, err error) {
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

func (s *mockSwapState) HandleProtocolMessage(_ Message) error {
	return nil
}

func (s *mockSwapState) Exit() error {
	return nil
}

func basicTestConfig(t *testing.T) *net.Config {
	_, chainID := tests.NewEthClient(t)
	// t.TempDir() is unique on every call. Don't reuse this config with multiple hosts.
	tmpDir := t.TempDir()
	return &net.Config{
		Ctx:         context.Background(),
		Environment: common.Development,
		DataDir:     tmpDir,
		EthChainID:  chainID.Int64(),
		Port:        0, // OS randomized libp2p port
		KeyFile:     path.Join(tmpDir, "node.key"),
		Bootnodes:   nil,
	}
}

func newHost(t *testing.T, cfg *net.Config) *host {
	h, err := NewHost(cfg, &mockHandler{})
	require.NoError(t, err)
	//h.SetHandler(&mockHandler{})
	t.Cleanup(func() {
		err = h.Stop()
		require.NoError(t, err)
	})
	return h
}
