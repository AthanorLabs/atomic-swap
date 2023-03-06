// Package wsclient provides client libraries for interacting with a local swapd instance
// over web sockets.
package wsclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/cockroachdb/apd/v3"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/common/vjson"

	"github.com/gorilla/websocket"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("rpcclient")

// WsClient ...
type WsClient interface {
	Close()
	Discover(provides string, searchTime uint64) ([]peer.ID, error)
	Query(who peer.ID) (*rpctypes.QueryPeerResponse, error)
	SubscribeSwapStatus(id types.Hash) (<-chan types.Status, error)
	TakeOfferAndSubscribe(peerID peer.ID, offerID types.Hash, providesAmount *apd.Decimal) (
		ch <-chan types.Status,
		err error,
	)
	MakeOfferAndSubscribe(
		min *apd.Decimal,
		max *apd.Decimal,
		exchangeRate *coins.ExchangeRate,
		ethAsset types.EthAsset,
		relayerFee *apd.Decimal,
	) (*rpctypes.MakeOfferResponse, <-chan types.Status, error)
}

type wsClient struct {
	wmu  sync.Mutex
	rmu  sync.Mutex
	conn *websocket.Conn
}

// NewWsClient ...
func NewWsClient(ctx context.Context, endpoint string) (*wsClient, error) { ///nolint:revive
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dial WS endpoint: %w", err)
	}

	if err = resp.Body.Close(); err != nil {
		return nil, err
	}

	return &wsClient{
		conn: conn,
	}, nil
}

func (c *wsClient) Close() {
	_ = c.conn.Close()
}

func (c *wsClient) writeJSON(msg *rpctypes.Request) error {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	return c.conn.WriteJSON(msg)
}

func (c *wsClient) read() ([]byte, error) {
	c.rmu.Lock()
	defer c.rmu.Unlock()
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (c *wsClient) Discover(provides string, searchTime uint64) ([]peer.ID, error) {
	params := &rpctypes.DiscoverRequest{
		Provides:   provides,
		SearchTime: searchTime,
	}

	bz, err := vjson.MarshalStruct(params)
	if err != nil {
		return nil, err
	}

	req := &rpctypes.Request{
		JSONRPC: rpctypes.DefaultJSONRPCVersion,
		Method:  rpctypes.NetDiscover,
		Params:  bz,
		ID:      0,
	}

	if err = c.writeJSON(req); err != nil {
		return nil, err
	}

	message, err := c.read()
	if err != nil {
		return nil, fmt.Errorf("failed to read websockets message: %s", err)
	}

	resp := new(rpctypes.Response)
	err = vjson.UnmarshalStruct(message, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("websocket server returned error: %w", resp.Error)
	}

	log.Debugf("received message over websockets: %s", message)
	dresp := new(rpctypes.DiscoverResponse)
	if err := vjson.UnmarshalStruct(resp.Result, dresp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal swap ID response: %s", err)
	}

	return dresp.PeerIDs, nil
}

func (c *wsClient) Query(id peer.ID) (*rpctypes.QueryPeerResponse, error) {
	params := &rpctypes.QueryPeerRequest{
		PeerID: id,
	}

	bz, err := vjson.MarshalStruct(params)
	if err != nil {
		return nil, err
	}

	req := &rpctypes.Request{
		JSONRPC: rpctypes.DefaultJSONRPCVersion,
		Method:  rpctypes.NetQueryPeer,
		Params:  bz,
		ID:      0,
	}

	if err = c.writeJSON(req); err != nil {
		return nil, err
	}

	// read ID from connection
	message, err := c.read()
	if err != nil {
		return nil, fmt.Errorf("failed to read websockets message: %s", err)
	}

	resp := new(rpctypes.Response)
	err = vjson.UnmarshalStruct(message, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("websocket server returned error: %w", resp.Error)
	}

	log.Debugf("received message over websockets: %s", message)
	dresp := new(rpctypes.QueryPeerResponse)
	if err := vjson.UnmarshalStruct(resp.Result, dresp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal swap ID response: %s", err)
	}

	return dresp, nil
}

// SubscribeSwapStatus returns a channel that is written to each time the swap's status updates.
// If there is no swap with the given ID, it returns an error.
func (c *wsClient) SubscribeSwapStatus(id types.Hash) (<-chan types.Status, error) {
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

	if err = c.writeJSON(req); err != nil {
		return nil, err
	}

	respCh := make(chan types.Status)

	go func() {
		defer close(respCh)

		for {
			message, err := c.read()
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

func (c *wsClient) TakeOfferAndSubscribe(
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

	if err = c.writeJSON(req); err != nil {
		return nil, err
	}

	// read resp from connection to see if there's an immediate error
	status, err := c.readTakeOfferResponse()
	if err != nil {
		return nil, err
	}

	respCh := make(chan types.Status)

	go func() {
		defer close(respCh)

		for {
			respCh <- status
			if !status.IsOngoing() {
				return
			}

			status, err = c.readTakeOfferResponse()
			if err != nil {
				log.Warnf("%s", err)
				break
			}
		}
	}()

	return respCh, nil
}

func (c *wsClient) readTakeOfferResponse() (types.Status, error) {
	message, err := c.read()
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

func (c *wsClient) MakeOfferAndSubscribe(
	min *apd.Decimal,
	max *apd.Decimal,
	exchangeRate *coins.ExchangeRate,
	ethAsset types.EthAsset,
	relayerFee *apd.Decimal,
) (*rpctypes.MakeOfferResponse, <-chan types.Status, error) {
	params := &rpctypes.MakeOfferRequest{
		MinAmount:    min,
		MaxAmount:    max,
		ExchangeRate: exchangeRate,
		EthAsset:     ethAsset,
		RelayerFee:   relayerFee,
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

	if err = c.writeJSON(req); err != nil {
		return nil, nil, err
	}

	// read ID from connection
	message, err := c.read()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read websockets message: %s", err)
	}

	resp := new(rpctypes.Response)
	err = vjson.UnmarshalStruct(message, resp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return nil, nil, fmt.Errorf("websocket server returned error: %w", resp.Error)
	}

	// read synchronous response (offer ID)
	respData := new(rpctypes.MakeOfferResponse)
	if err := vjson.UnmarshalStruct(resp.Result, respData); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal response: %s", err)
	}

	respCh := make(chan types.Status)

	go func() {
		defer close(respCh)

		for {
			message, err := c.read()
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
