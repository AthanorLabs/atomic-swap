package net

import (
	"context"
	"time"

	libp2pnetwork "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Handler interface {
	HandleProtocolMessage(msg Message) (resp Message, done bool, err error)
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

	msgBytes := make([]byte, 2048)

	for {
		tot, err := readStream(stream, msgBytes[:])
		if err != nil {
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

		resp, done, err := h.handler.HandleProtocolMessage(msg)
		if err != nil {
			log.Warnf("failed to handle protocol message: err=%s", err)
			return
		}

		if done {
			log.Info("protocol complete!")
			return
		}

		if err := h.writeToStream(stream, resp); err != nil {
			log.Warnf("failed to send response to peer: err=%s", err)
			return
		}
	}
}

func (h *host) Initiate(who peer.ID, msg *InitiateMessage) error {
	ctx, cancel := context.WithTimeout(h.ctx, protocolTimeout)
	defer cancel()

	// if err := h.h.Connect(ctx, who); err != nil {
	// 	return nil, err
	// }

	stream, err := h.h.NewStream(ctx, who, protocolID+subProtocolID)
	if err != nil {
		return err
	}

	log.Debug(
		"opened protocol stream, peer=", who,
	)

	h.handleProtocolStream(stream)
	return nil
}
