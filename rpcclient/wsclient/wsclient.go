// Package wsclient provides client libraries for interacting with a local swapd instance
// over web sockets.
package wsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/cockroachdb/apd/v3"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"

	"github.com/gorilla/websocket"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("rpcclient")

// WsClient ...
type WsClient interface {
	Close()
	Discover(provides coins.ProvidesCoin, searchTime uint64) ([]peer.ID, error)
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
		relayerEndpoint string,
		relayerCommission *apd.Decimal,
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

func (c *wsClient) Discover(provides coins.ProvidesCoin, searchTime uint64) ([]peer.ID, error) {
	params := &rpctypes.DiscoverRequest{
		Provides:   provides,
		SearchTime: searchTime,
	}

	bz, err := json.Marshal(params)
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

	var resp *rpctypes.Response
	err = json.Unmarshal(message, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("websocket server returned error: %w", resp.Error)
	}

	log.Debugf("received message over websockets: %s", message)
	var dresp *rpctypes.DiscoverResponse
	if err := json.Unmarshal(resp.Result, &dresp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal swap ID response: %s", err)
	}

	return dresp.PeerIDs, nil
}

func (c *wsClient) Query(id peer.ID) (*rpctypes.QueryPeerResponse, error) {
	params := &rpctypes.QueryPeerRequest{
		PeerID: id,
	}

	bz, err := json.Marshal(params)
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

	var resp *rpctypes.Response
	err = json.Unmarshal(message, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("websocket server returned error: %w", resp.Error)
	}

	log.Debugf("received message over websockets: %s", message)
	var dresp *rpctypes.QueryPeerResponse
	if err := json.Unmarshal(resp.Result, &dresp); err != nil {
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

	bz, err := json.Marshal(params)
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

			var resp *rpctypes.Response
			err = json.Unmarshal(message, &resp)
			if err != nil {
				log.Warnf("failed to unmarshal response: %s", err)
				break
			}

			if resp.Error != nil {
				log.Warnf("websocket server returned error: %s", resp.Error)
				break
			}

			log.Debugf("received message over websockets: %s", message)
			var status *rpctypes.SubscribeSwapStatusResponse
			if err := json.Unmarshal(resp.Result, &status); err != nil {
				log.Warnf("failed to unmarshal response: %s", err)
				break
			}

			s := types.NewStatus(status.Status)
			respCh <- s
			if !s.IsOngoing() {
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

	bz, err := json.Marshal(params)
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
			s := types.NewStatus(status)
			respCh <- s
			if !s.IsOngoing() {
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

func (c *wsClient) readTakeOfferResponse() (string, error) {
	message, err := c.read()
	if err != nil {
		return "", fmt.Errorf("failed to read websockets message: %s", err)
	}

	var resp *rpctypes.Response
	err = json.Unmarshal(message, &resp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return "", fmt.Errorf("websocket server returned error: %w", resp.Error)
	}

	log.Debugf("received message over websockets: %s", message)
	var status *rpctypes.SubscribeSwapStatusResponse
	if err := json.Unmarshal(resp.Result, &status); err != nil {
		return "", fmt.Errorf("failed to unmarshal swap status response: %w", err)
	}

	return status.Status, nil
}

func (c *wsClient) MakeOfferAndSubscribe(
	min *apd.Decimal,
	max *apd.Decimal,
	exchangeRate *coins.ExchangeRate,
	ethAsset types.EthAsset,
	relayerEndpoint string,
	relayerCommission *apd.Decimal,
) (*rpctypes.MakeOfferResponse, <-chan types.Status, error) {
	params := &rpctypes.MakeOfferRequest{
		MinAmount:         min,
		MaxAmount:         max,
		ExchangeRate:      exchangeRate,
		EthAsset:          ethAsset.Address().String(),
		RelayerEndpoint:   relayerEndpoint,
		RelayerCommission: relayerCommission,
	}

	bz, err := json.Marshal(params)
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

	var resp *rpctypes.Response
	err = json.Unmarshal(message, &resp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return nil, nil, fmt.Errorf("websocket server returned error: %w", resp.Error)
	}

	// read synchronous response (offer ID)
	var respData *rpctypes.MakeOfferResponse
	if err := json.Unmarshal(resp.Result, &respData); err != nil {
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

			var resp *rpctypes.Response
			err = json.Unmarshal(message, &resp)
			if err != nil {
				log.Warnf("failed to unmarshal response: %s", err)
				break
			}

			if resp.Error != nil {
				log.Warnf("websocket server returned error: %s", resp.Error)
				break
			}

			log.Debugf("received message over websockets: %s", message)
			var status *rpctypes.SubscribeSwapStatusResponse
			if err := json.Unmarshal(resp.Result, &status); err != nil {
				log.Warnf("failed to unmarshal response: %s", err)
				break
			}

			s := types.NewStatus(status.Status)
			respCh <- s
			if !s.IsOngoing() {
				return
			}
		}
	}()

	return respData, respCh, nil
}
