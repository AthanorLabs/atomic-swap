package net

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"path"
	"testing"

	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/tests"
)

const defaultMaxMessageSize = 1 << 17

func init() {
	logging.SetLogLevel("net", "debug")
}

func basicTestConfig(t *testing.T) *Config {
	_, chainID := tests.NewEthClient(t)
	// t.TempDir() is unique on every call. Don't reuse this config with multiple hosts.
	tmpDir := t.TempDir()
	return &Config{
		Ctx:         context.Background(),
		Environment: common.Development,
		DataDir:     tmpDir,
		EthChainID:  chainID.Int64(),
		Port:        0, // OS randomized libp2p port
		KeyFile:     path.Join(tmpDir, "node.key"),
		Bootnodes:   nil,
	}
}

func newHost(t *testing.T, cfg *Config) *host {
	h, err := NewHost(cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = h.Stop()
		require.NoError(t, err)
	})
	return h
}

func TestNewHost(t *testing.T) {
	cfg := basicTestConfig(t)
	h := newHost(t, cfg)
	err := h.Start()

	addresses := h.Addresses()
	require.NotEmpty(t, addresses)
	for _, addr := range h.Addresses() {
		t.Logf(addr)
	}

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
	readMsg, err := ReadStreamMessage(stream, defaultMaxMessageSize)
	require.NoError(t, err)
	require.Equal(t, msg.Type(), readMsg.Type())
}

func Test_readStreamMessage_EOF(t *testing.T) {
	// If the stream is closed before we read a length value, no message was truncated and
	// the returned error is io.EOF
	stream := bytes.NewReader(nil)
	_, err := ReadStreamMessage(stream, defaultMaxMessageSize)
	require.ErrorIs(t, err, io.EOF) // connection closed before we read any length

	// If the message was truncated either in the length or body, the error is io.ErrUnexpectedEOF
	serializedData := []byte{0x1} // truncated length
	stream = bytes.NewReader(serializedData)
	_, err = ReadStreamMessage(stream, defaultMaxMessageSize)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF) // connection after we read at least one byte

	serializedData = []byte{0x1, 0, 0, 0} // truncated encoded message
	stream = bytes.NewReader(serializedData)
	_, err = ReadStreamMessage(stream, defaultMaxMessageSize)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF) // connection after we read at least one byte
}

func Test_readStreamMessage_TooLarge(t *testing.T) {
	buf := make([]byte, 4+defaultMaxMessageSize+1)
	binary.LittleEndian.PutUint32(buf, defaultMaxMessageSize+1)
	_, err := ReadStreamMessage(bytes.NewReader(buf), defaultMaxMessageSize)
	require.ErrorContains(t, err, "too large")
}

func Test_readStreamMessage_NilStream(t *testing.T) {
	// Can our code actually trigger this error?
	_, err := ReadStreamMessage(nil, defaultMaxMessageSize)
	require.ErrorIs(t, err, errNilStream)
}

func Test_writeStreamMessage(t *testing.T) {
	msg := &message.QueryResponse{}
	peerID := peer.ID("")

	stream := &bytes.Buffer{}
	err := WriteStreamMessage(stream, msg, peerID)
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
