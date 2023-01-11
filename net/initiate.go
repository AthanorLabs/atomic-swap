package net

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	p2pnet "github.com/athanorlabs/go-p2p-net"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/net/message"
)

const (
	swapID          = "/swap/0"
	protocolTimeout = time.Second * 5
)

// Initiate attempts to initiate a swap with the given peer by sending a SendKeysMessage,
// the first message of the swap protocol.
func (h *Host) Initiate(who peer.AddrInfo, msg *SendKeysMessage, s common.SwapStateNet) error {
	h.swapMu.Lock()
	defer h.swapMu.Unlock()

	id := s.ID()

	if h.swaps[id] != nil {
		return errSwapAlreadyInProgress
	}

	ctx, cancel := context.WithTimeout(h.ctx, protocolTimeout)
	defer cancel()

	if h.h.Connectedness(who.ID) != libp2pnetwork.Connected {
		err := h.h.Connect(ctx, who)
		if err != nil {
			return err
		}
	}

	stream, err := h.h.NewStream(ctx, who.ID, protocol.ID(swapID))
	if err != nil {
		return fmt.Errorf("failed to open stream with peer: err=%w", err)
	}

	log.Debug(
		"opened protocol stream, peer=", who.ID,
	)

	if err := p2pnet.WriteStreamMessage(stream, msg, who.ID); err != nil {
		log.Warnf("failed to send initial SendKeysMessage to peer: err=%s", err)
		return err
	}

	h.swaps[id] = &swap{
		swapState: s,
		stream:    stream,
	}

	go h.handleProtocolStreamInner(stream, s)
	return nil
}

// handleProtocolStream is called when there is an incoming protocol stream.
func (h *Host) handleProtocolStream(stream libp2pnetwork.Stream) {
	if h.handler == nil {
		_ = stream.Close()
		return
	}

	msg, err := readStreamMessage(stream, maxMessageSize)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Debugf("Peer closed stream-id=%s, protocol exited", stream.ID())
		} else {
			log.Debugf("Failed to read message from peer, stream-id=%s: %s", stream.ID(), err)
		}
		_ = stream.Close()
		return
	}

	log.Debug(
		"received message from peer, peer=",
		stream.Conn().RemotePeer(),
		" type=",
		message.TypeToString(msg.Type()),
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

	if err := p2pnet.WriteStreamMessage(stream, resp, stream.Conn().RemotePeer()); err != nil {
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

	h.handleProtocolStreamInner(stream, s)
}

// handleProtocolStreamInner is called to handle a protocol stream, in both ingoing and outgoing cases.
func (h *Host) handleProtocolStreamInner(stream libp2pnetwork.Stream, s SwapState) {
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
		msg, err := readStreamMessage(stream, maxMessageSize)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Debug("Peer closed stream with us, protocol exited")
			} else {
				log.Debugf("Failed to read message from peer, id=%s protocol=%s: %s",
					stream.ID(), stream.Protocol(), err)
			}
			return
		}

		log.Debugf("received protocol=%s message from peer=%s type=%s",
			stream.Protocol(), stream.Conn().RemotePeer(), message.TypeToString(msg.Type()))

		err = s.HandleProtocolMessage(msg)
		if err != nil {
			log.Warnf("failed to handle protocol message: err=%s", err)
			return
		}
	}
}
