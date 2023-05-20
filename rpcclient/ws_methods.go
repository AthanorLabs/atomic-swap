// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"fmt"

	"github.com/cockroachdb/apd/v3"
	"github.com/gorilla/websocket"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/common/vjson"
)

var log = logging.Logger("rpcclient")

func (c *Client) wsConnect() (*websocket.Conn, error) {
	conn, resp, err := websocket.DefaultDialer.DialContext(c.ctx, c.wsEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dial WS endpoint: %w", err)
	}

	if err = resp.Body.Close(); err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) writeJSON(conn *websocket.Conn, msg *rpctypes.Request) error {
	return conn.WriteJSON(msg)
}

func (c *Client) read(conn *websocket.Conn) ([]byte, error) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	return message, nil
}

// SubscribeSwapStatus returns a channel that is written to each time the swap's status updates.
// If there is no swap with the given ID, it returns an error.
func (c *Client) SubscribeSwapStatus(id types.Hash) (<-chan types.Status, error) {
	params := &rpctypes.SubscribeSwapStatusRequest{
		OfferID: id,
	}

	bz, err := vjson.MarshalStruct(params)
	if err != nil {
		return nil, err
	}

	req := &rpctypes.Request{
		JSONRPC: rpctypes.DefaultJSONRPCVersion,
		Method:  rpctypes.SubscribeSwapStatus,
		Params:  bz,
		ID:      0,
	}

	conn, err := c.wsConnect()
	if err != nil {
		return nil, err
	}

	if err = c.writeJSON(conn, req); err != nil {
		_ = conn.Close()
		return nil, err
	}

	respCh := make(chan types.Status)

	go func() {
		defer func() { _ = conn.Close() }()
		defer close(respCh)

		for {
			message, err := c.read(conn)
			if err != nil {
				log.Warnf("failed to read websockets message: %s", err)
				break
			}

			resp := new(rpctypes.Response)
			err = vjson.UnmarshalStruct(message, resp)
			if err != nil {
				log.Warnf("failed to unmarshal response: %s", err)
				break
			}

			if resp.Error != nil {
				log.Warnf("websocket server returned error: %s", resp.Error)
				break
			}

			log.Debugf("received message over websockets: %s", message)
			statusResp := new(rpctypes.SubscribeSwapStatusResponse)
			if err := vjson.UnmarshalStruct(resp.Result, statusResp); err != nil {
				log.Warnf("failed to unmarshal response: %s", err)
				break
			}

			status := statusResp.Status
			respCh <- status
			if !status.IsOngoing() {
				return
			}
		}
	}()

	return respCh, nil
}

// TakeOfferAndSubscribe calls the server-side net_takeOfferAndSubscribe method
// to take and offer and get status updates over websockets.
func (c *Client) TakeOfferAndSubscribe(
	peerID peer.ID,
	offerID types.Hash,
	providesAmount *apd.Decimal,
) (ch <-chan types.Status, err error) {
	params := &rpctypes.TakeOfferRequest{
		PeerID:         peerID,
		OfferID:        offerID,
		ProvidesAmount: providesAmount,
	}

	bz, err := vjson.MarshalStruct(params)
	if err != nil {
		return nil, err
	}

	req := &rpctypes.Request{
		JSONRPC: rpctypes.DefaultJSONRPCVersion,
		Method:  rpctypes.SubscribeTakeOffer,
		Params:  bz,
		ID:      0,
	}

	conn, err := c.wsConnect()
	if err != nil {
		return nil, err
	}

	if err = c.writeJSON(conn, req); err != nil {
		_ = conn.Close()
		return nil, err
	}

	// read resp from connection to see if there's an immediate error
	status, err := c.readTakeOfferResponse(conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	respCh := make(chan types.Status)

	go func() {
		defer func() { _ = conn.Close() }()
		defer close(respCh)

		for {
			respCh <- status
			if !status.IsOngoing() {
				return
			}

			status, err = c.readTakeOfferResponse(conn)
			if err != nil {
				log.Warnf("%s", err)
				break
			}
		}
	}()

	return respCh, nil
}

func (c *Client) readTakeOfferResponse(conn *websocket.Conn) (types.Status, error) {
	message, err := c.read(conn)
	if err != nil {
		return 0, fmt.Errorf("failed to read websockets message: %s", err)
	}

	resp := new(rpctypes.Response)
	err = vjson.UnmarshalStruct(message, resp)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return 0, fmt.Errorf("websocket server returned error: %w", resp.Error)
	}

	log.Debugf("received message over websockets: %s", message)
	statusResp := new(rpctypes.SubscribeSwapStatusResponse)
	if err := vjson.UnmarshalStruct(resp.Result, statusResp); err != nil {
		return 0, fmt.Errorf("failed to unmarshal swap status response: %w", err)
	}

	return statusResp.Status, nil
}

// MakeOfferAndSubscribe calls the server-side net_makeOfferAndSubscribe method
// to make an offer and get status updates over websockets.
func (c *Client) MakeOfferAndSubscribe(
	min *apd.Decimal,
	max *apd.Decimal,
	exchangeRate *coins.ExchangeRate,
	ethAsset types.EthAsset,
	useRelayer bool,
) (*rpctypes.MakeOfferResponse, <-chan types.Status, error) {
	params := &rpctypes.MakeOfferRequest{
		MinAmount:    min,
		MaxAmount:    max,
		ExchangeRate: exchangeRate,
		EthAsset:     ethAsset,
		UseRelayer:   useRelayer,
	}

	bz, err := vjson.MarshalStruct(params)
	if err != nil {
		return nil, nil, err
	}

	req := &rpctypes.Request{
		JSONRPC: rpctypes.DefaultJSONRPCVersion,
		Method:  rpctypes.SubscribeMakeOffer,
		Params:  bz,
		ID:      0,
	}

	conn, err := c.wsConnect()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	if err = c.writeJSON(conn, req); err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	// read ID from connection
	message, err := c.read(conn)
	if err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("failed to read websockets message: %s", err)
	}

	resp := new(rpctypes.Response)
	err = vjson.UnmarshalStruct(message, resp)
	if err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("websocket server returned error: %w", resp.Error)
	}

	// read synchronous response (offer ID)
	respData := new(rpctypes.MakeOfferResponse)
	if err := vjson.UnmarshalStruct(resp.Result, respData); err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("failed to unmarshal response: %s", err)
	}

	respCh := make(chan types.Status)

	go func() {
		defer func() { _ = conn.Close() }()
		defer close(respCh)

		for {
			message, err := c.read(conn)
			if err != nil {
				log.Warnf("failed to read websockets message: %s", err)
				break
			}

			resp := new(rpctypes.Response)
			err = vjson.UnmarshalStruct(message, resp)
			if err != nil {
				log.Warnf("failed to unmarshal response: %s", err)
				break
			}

			if resp.Error != nil {
				log.Warnf("websocket server returned error: %s", resp.Error)
				break
			}

			log.Debugf("received message over websockets: %s", message)
			statusResp := new(rpctypes.SubscribeSwapStatusResponse)
			if err := vjson.UnmarshalStruct(resp.Result, statusResp); err != nil {
				log.Warnf("failed to unmarshal response: %s", err)
				break
			}

			s := statusResp.Status
			respCh <- s
			if !s.IsOngoing() {
				return
			}
		}
	}()

	return respData, respCh, nil
}
