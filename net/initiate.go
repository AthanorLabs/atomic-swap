package net

import (
	"context"
	"fmt"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"

	libp2pnetwork "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

const (
	swapID            = "/swap/0"
	protocolTimeout   = time.Second * 5
	messageBufferSize = 1 << 17
)

func (h *host) Initiate(who peer.AddrInfo, msg *SendKeysMessage, s common.SwapStateNet) error {
	h.swapMu.Lock()
	defer h.swapMu.Unlock()

	id := s.ID()

	if h.swaps[id] != nil {
		return errSwapAlreadyInProgress
	}

	ctx, cancel := context.WithTimeout(h.ctx, protocolTimeout)
	defer cancel()

	if h.h.Network().Connectedness(who.ID) != libp2pnetwork.Connected {
		err := h.h.Connect(ctx, who)
		if err != nil {
			return err
		}
	}

	stream, err := h.h.NewStream(ctx, who.ID, protocol.ID(h.protocolID+swapID))
	if err != nil {
		return fmt.Errorf("failed to open stream with peer: err=%w", err)
	}

	log.Debug(
		"opened protocol stream, peer=", who.ID,
	)

	if err := h.writeToStream(stream, msg); err != nil {
		log.Warnf("failed to send initial SendKeysMessage to peer: err=%s", err)
		return err
	}

	h.swaps[id] = &swap{
		swapState: s,
		stream:    stream,
	}

	go h.handleProtocolStreamInner(stream, s, make([]byte, messageBufferSize))
	return nil
}

// handleProtocolStream is called when there is an incoming protocol stream.
func (h *host) handleProtocolStream(stream libp2pnetwork.Stream) {
	if h.handler == nil {
		_ = stream.Close()
		return
	}

	buf := make([]byte, messageBufferSize)
	tot, err := readStream(stream, buf[:])
	if err != nil {
		log.Debug("peer closed stream with us, protocol exited")
		_ = stream.Close()
		return
	}

	// decode message based on message type
	msg, err := message.DecodeMessage(buf[:tot])
	if err != nil {
		log.Debug("failed to decode message from peer, id=", stream.ID(), " protocol=", stream.Protocol(), " err=", err)
		_ = stream.Close()
		return
	}

	log.Debug(
		"received message from peer, peer=", stream.Conn().RemotePeer(), " type=", msg.Type(),
	)

	im, ok := msg.(*SendKeysMessage)
	if !ok {
		log.Warnf("failed to handle protocol message: message was not SendKeysMessage")
		_ = stream.Close()
		return
	}

	var s SwapState
	s, resp, err := h.handler.HandleInitiateMessage(im)
	if err != nil {
		log.Warnf("failed to handle protocol message: err=%s", err)
		_ = stream.Close()
		return
	}

	if err := h.writeToStream(stream, resp); err != nil {
		log.Warnf("failed to send response to peer: err=%s", err)
		_ = s.Exit()
		_ = stream.Close()
		return
	}

	h.swapMu.Lock()
	h.swaps[s.ID()] = &swap{
		swapState: s,
		stream:    stream,
	}
	h.swapMu.Unlock()

	h.handleProtocolStreamInner(stream, s, buf)
}

// handleProtocolStreamInner is called to handle a protocol stream, in both ingoing and outgoing cases.
func (h *host) handleProtocolStreamInner(stream libp2pnetwork.Stream, s SwapState, buf []byte) {
	defer func() {
		log.Debugf("closing stream: peer=%s protocol=%s", stream.Conn().RemotePeer(), stream.Protocol())
		_ = stream.Close()

		log.Debugf("exiting swap...")
		if err := s.Exit(); err != nil {
			log.Errorf("failed to exit protocol: err=%s", err)
		}
		h.swapMu.Lock()
		delete(h.swaps, s.ID())
		h.swapMu.Unlock()
	}()

	for {
		tot, err := readStream(stream, buf[:])
		if err != nil {
			log.Debug("peer closed stream with us, protocol exited")
			return
		}

		// decode message based on message type
		msg, err := message.DecodeMessage(buf[:tot])
		if err != nil {
			log.Debug("failed to decode message from peer, id=", stream.ID(), " protocol=", stream.Protocol(), " err=", err)
			continue
		}

		log.Debug(
			"received message from peer, peer=", stream.Conn().RemotePeer(), " type=", msg.Type(),
		)

		resp, done, err := s.HandleProtocolMessage(msg)
		if err != nil {
			log.Warnf("failed to handle protocol message: err=%s", err)
			return
		}

		if resp == nil {
			continue
		}

		if err := h.writeToStream(stream, resp); err != nil {
			log.Warnf("failed to send response to peer: err=%s", err)
			return
		}

		if done {
			log.Debug("protocol complete!")
			return
		}
	}
}

// CloseProtocolStream closes the current swap protocol stream.
func (h *host) CloseProtocolStream(id types.Hash) {
	swap, has := h.swaps[id]
	if !has {
		return
	}

	log.Debugf("closing stream: peer=%s protocol=%s",
		swap.stream.Conn().RemotePeer(), swap.stream.Protocol(),
	)
	_ = swap.stream.Close()
}
