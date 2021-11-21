package net

import (
	"context"
	"errors"
	"fmt"
	"time"

	libp2pnetwork "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

// Handler handles incoming protocol messages.
// It is implemented by *alice.alice and *bob.bob
type Handler interface {
	HandleInitiateMessage(msg *InitiateMessage) (s SwapState, resp Message, err error)
}

// SwapState handles incoming protocol messages for an initiated protocol.
// It is implemented by *alice.swapState and *bob.swapState
type SwapState interface {
	HandleProtocolMessage(msg Message) (resp Message, done bool, err error)
	ProtocolComplete()
}

// SetSwapState sets the current SwapState for the host. It errors if a swap is
// already occuring.
func (h *host) SetSwapState(s SwapState) error {
	if h.swapState != nil {
		return errors.New("swap already occuring")
	}

	h.swapState = s
	return nil
}

const (
	subProtocolID   = "/protocol/0"
	protocolTimeout = time.Second * 5
)

func (h *host) handleProtocolStream(stream libp2pnetwork.Stream) {
	defer func() {
		log.Debugf("closing stream: peer=%s protocol=%s", stream.Conn().RemotePeer(), stream.Protocol())
		_ = stream.Close()
	}()

	// TODO: need a lock for this, otherwise two streams can enter this func
	if h.swapState != nil {
		// TODO: check if peer is the peer we initiated with, otherwise return
	}

	defer func() {
		h.swapState.ProtocolComplete()
	}()

	msgBytes := make([]byte, 2048)

	for {
		tot, err := readStream(stream, msgBytes[:])
		if err != nil {
			log.Debug("peer closed stream with us, protocol exited")
			return
		}

		// decode message based on message type
		msg, err := decodeMessage(msgBytes[:tot])
		if err != nil {
			log.Debug("failed to decode message from peer, id=", stream.ID(), " protocol=", stream.Protocol(), " err=", err)
			continue
		}

		log.Debug(
			"received message from peer, peer=", stream.Conn().RemotePeer(), " msg=", msg.String(),
		)

		var (
			resp Message
			done bool
		)

		if h.swapState == nil {
			im, ok := msg.(*InitiateMessage)
			if !ok {
				log.Warnf("failed to handle protocol message: message was not InitiateMessage")
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

func (h *host) Initiate(who peer.AddrInfo, msg *InitiateMessage, s SwapState) error {
	// TODO: need a lock for this, otherwise two streams can enter this func
	if h.swapState != nil {
		return errors.New("already have ongoing swap")
	}

	ctx, cancel := context.WithTimeout(h.ctx, protocolTimeout)
	defer cancel()

	if err := h.h.Connect(ctx, who); err != nil {
		return err
	}

	stream, err := h.h.NewStream(ctx, who.ID, protocolID+subProtocolID)
	if err != nil {
		return fmt.Errorf("failed to open stream with peer: err=%w", err)
	}

	log.Debug(
		"opened protocol stream, peer=", who.ID,
	)

	if err := h.writeToStream(stream, msg); err != nil {
		log.Warnf("failed to send InitiateMessage to peer: err=%s", err)
		return err
	}

	h.swapState = s
	h.handleProtocolStream(stream)
	return nil
}
