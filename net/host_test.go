package net

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/tests"
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

func (s *mockSwapState) HandleProtocolMessage(msg Message) error {
	return nil
}

func (s *mockSwapState) Exit() error {
	return nil
}

func newHost(t *testing.T, port uint16) *host {
	_, chainID := tests.NewEthClient(t)
	cfg := &Config{
		Ctx:           context.Background(),
		Environment:   common.Development,
		DataDir:       t.TempDir(),
		EthChainID:    chainID.Int64(),
		Port:          port,
		KeyFile:       path.Join(t.TempDir(), fmt.Sprintf("node-%d.key", port)),
		Bootnodes:     []string{},
		StaticNATPort: false,
	}

	h, err := NewHost(cfg)
	require.NoError(t, err)
	h.SetHandler(&mockHandler{})
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

func Test_readStreamMessage(t *testing.T) {
	msg := &message.QueryResponse{}
	msgBytes, err := msg.Encode()
	require.NoError(t, err)
	var lenBytes [4]byte
	binary.LittleEndian.PutUint32(lenBytes[:], uint32(len(msgBytes)))
	streamData := append(lenBytes[:], msgBytes...)
	stream := bytes.NewReader(streamData)
	readMsg, err := readStreamMessage(stream)
	require.NoError(t, err)
	require.Equal(t, msg.Type(), readMsg.Type())
}

func Test_readStreamMessage_EOF(t *testing.T) {
	// If the stream is closed before we read a length value, no message was truncated and
	// the returned error is io.EOF
	stream := bytes.NewReader(nil)
	_, err := readStreamMessage(stream)
	require.ErrorIs(t, err, io.EOF) // connection closed before we read any length

	// If the message was truncated either in the length or body, the error is io.ErrUnexpectedEOF
	serializedData := []byte{0x1} // truncated length
	stream = bytes.NewReader(serializedData)
	_, err = readStreamMessage(stream)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF) // connection after we read at least one byte

	serializedData = []byte{0x1, 0, 0, 0} // truncated encoded message
	stream = bytes.NewReader(serializedData)
	_, err = readStreamMessage(stream)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF) // connection after we read at least one byte
}

func Test_writeStreamMessage(t *testing.T) {
	msg := &message.QueryResponse{}
	peerID := peer.ID("")

	stream := &bytes.Buffer{}
	err := writeStreamMessage(stream, msg, peerID)
	require.NoError(t, err)
	serializedData := stream.Bytes()
	require.Greater(t, len(serializedData), 4)
	lenMsg := binary.LittleEndian.Uint32(serializedData)
	msgBytes := serializedData[4:]
	require.Equal(t, int(lenMsg), len(msgBytes))
	writtenMsg, err := message.DecodeMessage(msgBytes)
	require.NoError(t, err)
	require.Equal(t, msg.Type(), writtenMsg.Type())
}
