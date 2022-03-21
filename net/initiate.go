package net

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/net/message"

	libp2pnetwork "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

const (
	swapID          = "/swap/0"
	protocolTimeout = time.Second * 5
)

func (h *host) Initiate(who peer.AddrInfo, msg *SendKeysMessage, s common.SwapState) error {
	h.swapMu.Lock()
	defer h.swapMu.Unlock()

	if h.swapState != nil {
		return errors.New("already have ongoing swap")
	}

	ctx, cancel := context.WithTimeout(h.ctx, protocolTimeout)
	defer cancel()

	if err := h.h.Connect(ctx, who); err != nil {
		return err
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

	h.swapState = s
	h.swapStream = stream
	go h.handleProtocolStreamInner(stream)
	return nil
}

// handleProtocolStream is called when there is an incoming protocol stream.
func (h *host) handleProtocolStream(stream libp2pnetwork.Stream) {
	if h.handler == nil {
		return
	}

	h.swapMu.Lock()
	defer h.swapMu.Unlock()

	if h.swapState != nil {
		log.Debug("failed to handling incoming swap stream, already have ongoing swap")
	}

	h.swapStream = stream
	h.handleProtocolStreamInner(stream)
}

// handleProtocolStreamInner is called to handle a protocol stream, in both ingoing and outgoing cases.
func (h *host) handleProtocolStreamInner(stream libp2pnetwork.Stream) {
	defer func() {
		log.Debugf("closing stream: peer=%s protocol=%s", stream.Conn().RemotePeer(), stream.Protocol())
		_ = stream.Close()
		if h.swapState != nil {
			log.Debugf("exiting swap...")
			if err := h.swapState.Exit(); err != nil {
				log.Errorf("failed to exit protocol: err=%s", err)
			}
			h.swapState = nil
		}
	}()

	msgBytes := make([]byte, 1<<17)

	for {
		tot, err := readStream(stream, msgBytes[:])
		if err != nil {
			log.Debug("peer closed stream with us, protocol exited")
			return
		}

		// decode message based on message type
		msg, err := message.DecodeMessage(msgBytes[:tot])
		if err != nil {
			log.Debug("failed to decode message from peer, id=", stream.ID(), " protocol=", stream.Protocol(), " err=", err)
			continue
		}

		log.Debug(
			"received message from peer, peer=", stream.Conn().RemotePeer(), " type=", msg.Type(),
		)

		var (
			resp Message
			done bool
		)

		if h.swapState == nil {
			im, ok := msg.(*SendKeysMessage)
			if !ok {
				log.Warnf("failed to handle protocol message: message was not SendKeysMessage")
				return
			}

			var s SwapState
			s, resp, err = h.handler.HandleInitiateMessage(im)
			if err != nil {
				log.Warnf("failed to handle protocol message: err=%s", err)
				return
			}

			h.swapState = s
		} else {
			resp, done, err = h.swapState.HandleProtocolMessage(msg)
			if err != nil {
				log.Warnf("failed to handle protocol message: err=%s", err)
				return
			}
		}

		if resp == nil {
			continue
		}

		if err := h.writeToStream(stream, resp); err != nil {
			log.Warnf("failed to send response to peer: err=%s", err)
			return
		}

		if done {
			log.Info("protocol complete!")
			return
		}
	}
}

// CloseProtocolStream closes the current swap protocol stream.
func (h *host) CloseProtocolStream() {
	stream := h.swapStream
	log.Debugf("closing stream: peer=%s protocol=%s", stream.Conn().RemotePeer(), stream.Protocol())
	_ = stream.Close()
}
