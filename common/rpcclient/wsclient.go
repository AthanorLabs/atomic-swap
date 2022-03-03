package rpcclient

import (
	"context"
	"encoding/json"

	"github.com/gorilla/websocket"
	logging "github.com/ipfs/go-log"

	"github.com/noot/atomic-swap/common"
)

// DefaultJSONRPCVersion ...
const DefaultJSONRPCVersion = "2.0"

var log = logging.Logger("rpcclient")

// WsClient ...
type WsClient interface {
	SubscribeSwapStatus(id uint64) (<-chan common.Stage, error)
	TakeOfferAndSubscribe(multiaddr, offerID string,
		providesAmount float64) (id uint64, ch <-chan common.StageOrExitStatus, err error)
}

type wsClient struct {
	conn *websocket.Conn
}

// NewWsClient ...
func NewWsClient(ctx context.Context, endpoint string) (*wsClient, error) { ///nolint:revive
	conn, _, err := (&websocket.Dialer{}).DialContext(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}

	return &wsClient{
		conn: conn,
	}, nil
}

// SubscribeSwapStatus returns a channel that is written to each time the swap's status updates.
// If there is no swap with the given ID, it returns an error.
func (c *wsClient) SubscribeSwapStatus(id uint64) (<-chan common.Stage, error) {
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

	respCh := make(chan common.Stage)
	defer close(respCh)

	go func() {
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
		}
	}()

	return respCh, nil
}
