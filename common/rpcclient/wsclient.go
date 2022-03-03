package rpcclient

import (
	"context"
	"encoding/json"

	"github.com/gorilla/websocket"
	logging "github.com/ipfs/go-log"

	"github.com/noot/atomic-swap/common"
)

const DefaultJSONRPCVersion = "2.0"

var log = logging.Logger("rpcclient")

type WsClient interface {
	SubscribeSwapStatus(id uint64) (<-chan common.Stage, error)
}

type wsClient struct {
	conn *websocket.Conn
}

func NewWsClient(ctx context.Context, endpoint string) (*wsClient, error) {
	conn, _, err := (&websocket.Dialer{}).DialContext(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}

	return &wsClient{
		conn: conn,
	}, nil
}

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
