// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

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
	swapID = "/swap/0"
)

// Initiate attempts to initiate a swap with the given peer by sending a SendKeysMessage,
// the first message of the swap protocol.
func (h *Host) Initiate(who peer.AddrInfo, sendKeysMessage common.Message, s common.SwapStateNet) error {
	h.swapMu.Lock()
	defer h.swapMu.Unlock()

	id := s.OfferID()

	if h.swaps[id] != nil {
		return errSwapAlreadyInProgress
	}

	ctx, cancel := context.WithTimeout(h.ctx, connectionTimeout)
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

	if err := p2pnet.WriteStreamMessage(stream, sendKeysMessage, who.ID); err != nil {
		log.Warnf("failed to send initial SendKeysMessage to peer: err=%s", err)
		return err
	}

	h.swaps[id] = &swap{
		swapState: s,
		stream:    stream,
		isTaker:   true,
	}

	go h.receiveInitiateResponse(stream, s)
	return nil
}

func (h *Host) receiveInitiateResponse(stream libp2pnetwork.Stream, s SwapState) {
	defer h.handleProtocolStreamClose(stream, s)

	const initiateResponseTimeout = time.Minute

	select {
	case msg := <-nextStreamMessage(stream, maxMessageSize):
		if msg == nil {
			log.Errorf("failed to read initial SendKeysMessage response")
			return
		}

		log.Debugf("received protocol=%s message from peer=%s type=%s",
			stream.Protocol(), stream.Conn().RemotePeer(), message.TypeToString(msg.Type()))

		err := s.HandleProtocolMessage(msg)
		if err != nil {
			log.Warnf("failed to handle protocol message: err=%s", err)
			return
		}
	case <-time.After(initiateResponseTimeout):
		log.Errorf("timed out waiting for SendKeysMessage response")
		return
	}

	h.handleProtocolStreamInner(stream, s)
}

// handleProtocolStream is called when there is an incoming protocol stream.
func (h *Host) handleProtocolStream(stream libp2pnetwork.Stream) {
	if h.makerHandler == nil {
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

	curPeer := stream.Conn().RemotePeer()

	log.Debugf("received message from peer=%s type=%s", curPeer, message.TypeToString(msg.Type()))

	im, ok := msg.(*SendKeysMessage)
	if !ok {
		log.Warnf("failed to handle protocol message: message was not SendKeysMessage")
		_ = stream.Close()
		return
	}

	var s SwapState
	s, resp, err := h.makerHandler.HandleInitiateMessage(curPeer, im)
	if err != nil {
		log.Warnf("failed to handle protocol message: err=%s", err)
		_ = stream.Close()
		return
	}

	if err := p2pnet.WriteStreamMessage(stream, resp, stream.Conn().RemotePeer()); err != nil {
		log.Warnf("failed to send response to peer: %s", err)
		if err = s.Exit(); err != nil {
			log.Warnf("Swap exit failure: %s", err)
		}
		_ = stream.Close()
		return
	}

	h.swapMu.Lock()
	h.swaps[s.OfferID()] = &swap{
		swapState: s,
		stream:    stream,
		isTaker:   false,
	}
	h.swapMu.Unlock()

	h.handleProtocolStreamInner(stream, s)
}

// handleProtocolStreamInner is called to handle a protocol stream, in both ingoing and outgoing cases.
func (h *Host) handleProtocolStreamInner(stream libp2pnetwork.Stream, s SwapState) {
	defer h.handleProtocolStreamClose(stream, s)

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
			log.Warnf("failed to handle protocol message: %s", err)
			return
		}
	}
}

func (h *Host) handleProtocolStreamClose(stream libp2pnetwork.Stream, s SwapState) {
	log.Debugf("closing stream: peer=%s protocol=%s", stream.Conn().RemotePeer(), stream.Protocol())
	_ = stream.Close()

	log.Debugf("exiting swap...")
	if err := s.Exit(); err != nil {
		log.Errorf("failed to exit protocol: %s", err)
	}
	h.swapMu.Lock()
	delete(h.swaps, s.OfferID())
	h.swapMu.Unlock()
}
