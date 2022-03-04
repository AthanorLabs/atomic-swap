package rpcclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
	logging "github.com/ipfs/go-log"
	//"github.com/noot/atomic-swap/common"
)

// DefaultJSONRPCVersion ...
const DefaultJSONRPCVersion = "2.0"

var log = logging.Logger("rpcclient")

// WsClient ...
type WsClient interface {
	SubscribeSwapStatus(id uint64) (<-chan string, error)
	TakeOfferAndSubscribe(multiaddr, offerID string,
		providesAmount float64) (id uint64, ch <-chan string, err error)
}

type wsClient struct {
	conn *websocket.Conn
}

// NewWsClient ...
func NewWsClient(ctx context.Context, endpoint string) (*wsClient, error) { ///nolint:revive
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dial endpoint: %w", err)
	}

	return &wsClient{
		conn: conn,
	}, nil
}

// SubscribeSwapStatus returns a channel that is written to each time the swap's status updates.
// If there is no swap with the given ID, it returns an error.
func (c *wsClient) SubscribeSwapStatus(id uint64) (<-chan string, error) {
	req := &Request{
		JSONRPC: DefaultJSONRPCVersion,
		Method:  "swap_subscribeStatus",
		Params: map[string]interface{}{
			"id": id,
		},
		ID: 0,
	}

	if err := c.conn.WriteJSON(req); err != nil {
		return nil, err
	}

	respCh := make(chan string)

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

			respCh <- status.Stage
		}
	}()

	return respCh, nil
}

func (c *wsClient) TakeOfferAndSubscribe(multiaddr, offerID string,
	providesAmount float64) (id uint64, ch <-chan string, err error) {
	req := &Request{
		JSONRPC: DefaultJSONRPCVersion,
		Method:  "net_takeOfferAndSubscribe",
		Params: map[string]interface{}{
			"multiaddr":      multiaddr,
			"offerID":        offerID,
			"providesAmount": providesAmount,
		},
		ID: 0,
	}

	if err := c.conn.WriteJSON(req); err != nil {
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
	var idResp map[string]uint64
	if err := json.Unmarshal(resp.Result, &idResp); err != nil {
		return 0, nil, fmt.Errorf("failed to unmarshal response: %s", err)
	}

	id, ok := idResp["id"]
	if !ok {
		return 0, nil, errors.New("websocket response did not contain ID")
	}

	respCh := make(chan string)

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

			respCh <- status.Stage
		}
	}()

	return id, respCh, nil
}
