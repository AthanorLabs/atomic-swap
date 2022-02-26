package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/rpcclient"
	"github.com/noot/atomic-swap/common/types"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type wsServer struct {
	sm    SwapManager
	alice Alice
	bob   Bob
}

func newWsServer(sm SwapManager, a Alice, b Bob) *wsServer {
	return &wsServer{
		sm:    sm,
		alice: a,
		bob:   b,
	}
}

func (s *wsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warnf("failed to update connection to websockets: %s", err)
		return
	}

	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Warnf("failed to read websockets message: %s", err)
			break
		}

		var req *Request
		err = json.Unmarshal(message, &req)
		if err != nil {
			_ = writeError(conn, err)
			continue
		}

		log.Debugf("received message over websockets: %s", message)
		err = s.handleRequest(conn, req)
		if err != nil {
			_ = writeError(conn, err)
		}
	}
}

type Request struct {
	JsonRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      uint64                 `json:"id"`
}

const (
	defaultJsonRPCVersion = "2.0"

	subscribeNewPeer    = "net_subscribeNewPeer"
	subscribeSwapStatus = "swap_subscribeStatus"
)

func (s *wsServer) handleRequest(conn *websocket.Conn, req *Request) error {
	switch req.Method {
	case subscribeNewPeer:
		return errors.New("unimplemented")
	case subscribeSwapStatus:
		idi, has := req.Params["id"] // TODO: make const
		if !has {
			return errors.New("params missing id field")
		}

		id, ok := idi.(float64)
		if !ok {
			return fmt.Errorf("failed to cast id parameter to float64: got %T", idi)
		}

		return s.subscribeSwapStatus(conn, uint64(id))
	default:
		return errors.New("invalid method")
	}
}

type SubscribeSwapStatusResponse struct {
	Stage string `json:"stage"`
}

// subscribeSwapStatus writes the swap's stage to the connection every time it updates.
// when the swap completes, it writes the final status then closes the connection.
// example: `{"jsonrpc":"2.0", "method":"swap_subscribeStatus", "params": {"id": 0}, "id": 0}`
func (s *wsServer) subscribeSwapStatus(conn *websocket.Conn, id uint64) error {
	var prevStage common.Stage
	for {
		info := s.sm.GetOngoingSwap()
		if info == nil {
			info := s.sm.GetPastSwap(id)
			if info == nil {
				return errors.New("unable to find swap with given ID")
			}

			resp := &SubscribeSwapStatusResponse{
				Stage: info.Status().String(),
			}

			if err := writeResponse(conn, resp); err != nil {
				return err
			}

			return nil
		}

		var swapState common.SwapState
		switch info.Provides() {
		case types.ProvidesETH:
			swapState = s.alice.GetOngoingSwapState()
		case types.ProvidesXMR:
			swapState = s.bob.GetOngoingSwapState()
		}

		if swapState == nil {
			// we probably completed the swap, continue to call GetPastSwap
			continue
		}

		currStage := swapState.Stage()
		if currStage == prevStage {
			time.Sleep(time.Millisecond * 10)
			continue
		}

		resp := &SubscribeSwapStatusResponse{
			Stage: currStage.String(),
		}

		if err := writeResponse(conn, resp); err != nil {
			return err
		}

		prevStage = currStage
	}
}

func writeResponse(conn *websocket.Conn, result interface{}) error {
	bz, err := json.Marshal(result)
	if err != nil {
		return err
	}

	resp := &rpcclient.ServerResponse{
		Version: defaultJsonRPCVersion,
		Result:  bz,
	}

	return conn.WriteJSON(resp)
}

func writeError(conn *websocket.Conn, err error) error {
	resp := &rpcclient.ServerResponse{
		Version: defaultJsonRPCVersion,
		Error: &rpcclient.Error{
			Message: err.Error(),
		},
	}

	return conn.WriteJSON(resp)
}
