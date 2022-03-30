package rpcclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/noot/atomic-swap/common/types"

	"github.com/gorilla/websocket"
	logging "github.com/ipfs/go-log"
)

// DefaultJSONRPCVersion ...
const DefaultJSONRPCVersion = "2.0"

var log = logging.Logger("rpcclient")

// WsClient ...
type WsClient interface {
	SubscribeSwapStatus(id uint64) (<-chan types.Status, error)
	TakeOfferAndSubscribe(multiaddr, offerID string,
		providesAmount float64) (id uint64, ch <-chan types.Status, err error)
	MakeOfferAndSubscribe(min, max float64,
		exchangeRate types.ExchangeRate) (string, <-chan *MakeOfferTakenResponse, <-chan types.Status, error)
}

type wsClient struct {
	conn *websocket.Conn
}

// NewWsClient ...
func NewWsClient(ctx context.Context, endpoint string) (*wsClient, error) { ///nolint:revive
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dial endpoint: %w", err)
	}

	if err = resp.Body.Close(); err != nil {
		return nil, err
	}

	return &wsClient{
		conn: conn,
	}, nil
}

// SubscribeSwapStatusRequestParams ...
type SubscribeSwapStatusRequestParams struct {
	ID uint64 `json:"id"`
}

// SubscribeSwapStatus returns a channel that is written to each time the swap's status updates.
// If there is no swap with the given ID, it returns an error.
func (c *wsClient) SubscribeSwapStatus(id uint64) (<-chan types.Status, error) {
	params := &SubscribeSwapStatusRequestParams{
		ID: id,
	}

	bz, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req := &Request{
		JSONRPC: DefaultJSONRPCVersion,
		Method:  "swap_subscribeStatus",
		Params:  bz,
		ID:      0,
	}

	if err := c.conn.WriteJSON(req); err != nil {
		return nil, err
	}

	respCh := make(chan types.Status)

	go func() {
		defer close(respCh)

		for {
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Warnf("failed to read websockets message: %s", err)
				break
			}

			var resp *Response
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
			var status *SubscribeSwapStatusResponse
			if err := json.Unmarshal(resp.Result, &status); err != nil {
				log.Warnf("failed to unmarshal response: %s", err)
				break
			}

			respCh <- types.NewStatus(status.Status)
		}
	}()

	return respCh, nil
}

// SubscribeTakeOfferParams ...
// TODO: duplciate of rpc.TakeOfferRequest
type SubscribeTakeOfferParams struct {
	Multiaddr      string  `json:"multiaddr"`
	OfferID        string  `json:"offerID"`
	ProvidesAmount float64 `json:"providesAmount"`
}

// TakeOfferResponse ...
type TakeOfferResponse struct {
	ID       uint64 `json:"id"`
	InfoFile string `json:"infoFile"`
}

func (c *wsClient) TakeOfferAndSubscribe(multiaddr, offerID string,
	providesAmount float64) (id uint64, ch <-chan types.Status, err error) {
	params := &SubscribeTakeOfferParams{
		Multiaddr:      multiaddr,
		OfferID:        offerID,
		ProvidesAmount: providesAmount,
	}

	bz, err := json.Marshal(params)
	if err != nil {
		return 0, nil, err
	}

	req := &Request{
		JSONRPC: DefaultJSONRPCVersion,
		Method:  "net_takeOfferAndSubscribe",
		Params:  bz,
		ID:      0,
	}

	if err = c.conn.WriteJSON(req); err != nil {
		return 0, nil, err
	}

	// read ID from connection
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read websockets message: %s", err)
	}

	var resp *Response
	err = json.Unmarshal(message, &resp)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return 0, nil, fmt.Errorf("websocket server returned error: %w", resp.Error)
	}

	log.Debugf("received message over websockets: %s", message)
	var idResp *TakeOfferResponse
	if err := json.Unmarshal(resp.Result, &idResp); err != nil {
		return 0, nil, fmt.Errorf("failed to unmarshal swap ID response: %s", err)
	}

	respCh := make(chan types.Status)

	go func() {
		defer close(respCh)

		for {
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Warnf("failed to read websockets message: %s", err)
				break
			}

			var resp *Response
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
			var status *SubscribeSwapStatusResponse
			if err := json.Unmarshal(resp.Result, &status); err != nil {
				log.Warnf("failed to unmarshal swap status response: %s", err)
				break
			}

			respCh <- types.NewStatus(status.Status)
		}
	}()

	return idResp.ID, respCh, nil
}

// MakeOfferRequest ...
// TODO: duplicate of rpc.MakeOfferRequest
type MakeOfferRequest struct {
	MinimumAmount float64            `json:"minimumAmount"`
	MaximumAmount float64            `json:"maximumAmount"`
	ExchangeRate  types.ExchangeRate `json:"exchangeRate"`
}

// MakeOfferResponse ...
type MakeOfferResponse struct {
	ID       string `json:"offerID"`
	InfoFile string `json:"infoFile"`
}

// MakeOfferTakenResponse contains the swap ID
type MakeOfferTakenResponse struct {
	ID uint64 `json:"id"`
}

func (c *wsClient) MakeOfferAndSubscribe(min, max float64,
	exchangeRate types.ExchangeRate) (string, <-chan *MakeOfferTakenResponse, <-chan types.Status, error) {
	params := &MakeOfferRequest{
		MinimumAmount: min,
		MaximumAmount: max,
		ExchangeRate:  exchangeRate,
	}

	bz, err := json.Marshal(params)
	if err != nil {
		return "", nil, nil, err
	}

	req := &Request{
		JSONRPC: DefaultJSONRPCVersion,
		Method:  "net_makeOfferAndSubscribe", // TODO: use const
		Params:  bz,
		ID:      0,
	}

	if err = c.conn.WriteJSON(req); err != nil {
		return "", nil, nil, err
	}

	// read ID from connection
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to read websockets message: %s", err)
	}

	var resp *Response
	err = json.Unmarshal(message, &resp)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return "", nil, nil, fmt.Errorf("websocket server returned error: %w", resp.Error)
	}

	// read synchronous response (offer ID and infofile)
	log.Debugf("received message over websockets: %s", message)
	var respData *MakeOfferResponse
	if err := json.Unmarshal(resp.Result, &respData); err != nil {
		return "", nil, nil, fmt.Errorf("failed to unmarshal response: %s", err)
	}

	takenCh := make(chan *MakeOfferTakenResponse)
	respCh := make(chan types.Status)

	go func() {
		defer close(respCh)
		defer close(takenCh)

		// read if swap was taken
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Warnf("failed to read websockets message: %s", err)
			return
		}

		var resp *Response
		err = json.Unmarshal(message, &resp)
		if err != nil {
			log.Warnf("failed to unmarshal response: %s", err)
			return
		}

		if resp.Error != nil {
			log.Warnf("websocket server returned error: %s", resp.Error)
			return
		}

		log.Debugf("received message over websockets: %s", message)
		var taken *MakeOfferTakenResponse
		if err := json.Unmarshal(resp.Result, &taken); err != nil {
			log.Warnf("failed to unmarshal response: %s", err)
			return
		}

		takenCh <- taken

		for {
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Warnf("failed to read websockets message: %s", err)
				break
			}

			var resp *Response
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
			var status *SubscribeSwapStatusResponse
			if err := json.Unmarshal(resp.Result, &status); err != nil {
				log.Warnf("failed to unmarshal response: %s", err)
				break
			}

			respCh <- types.NewStatus(status.Status)
		}
	}()

	return respData.ID, takenCh, respCh, nil
}
