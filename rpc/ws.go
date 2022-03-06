package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/noot/atomic-swap/common/rpcclient"
	"github.com/noot/atomic-swap/common/types"

	"github.com/gorilla/websocket"
)

const (
	subscribeNewPeer    = "net_subscribeNewPeer"
	subscribeTakeOffer  = "net_takeOfferAndSubscribe"
	subscribeSwapStatus = "swap_subscribeStatus"
)

var upgrader = websocket.Upgrader{}

//nolint:revive
type (
	Request                     = rpcclient.Request
	Response                    = rpcclient.Response
	SubscribeSwapStatusResponse = rpcclient.SubscribeSwapStatusResponse
)

type wsServer struct {
	ctx   context.Context
	sm    SwapManager
	alice Alice
	bob   Bob
	ns    *NetService
}

func newWsServer(ctx context.Context, sm SwapManager, a Alice, b Bob, ns *NetService) *wsServer {
	return &wsServer{
		ctx:   ctx,
		sm:    sm,
		alice: a,
		bob:   b,
		ns:    ns,
	}
}

// ServeHTTP ...
func (s *wsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warnf("failed to update connection to websockets: %s", err)
		return
	}

	defer conn.Close() //nolint:errcheck

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

		return s.subscribeSwapStatus(s.ctx, conn, uint64(id))
	case subscribeTakeOffer:
		maddri, has := req.Params["multiaddr"]
		if !has {
			return errors.New("params missing multiaddr field")
		}

		maddr, ok := maddri.(string)
		if !ok {
			return fmt.Errorf("failed to cast multiaddr parameter to string: got %T", maddri)
		}

		offerIDi, has := req.Params["offerID"]
		if !has {
			return errors.New("params missing offerID field")
		}

		offerID, ok := offerIDi.(string)
		if !ok {
			return fmt.Errorf("failed to cast multiaddr parameter to string: got %T", offerIDi)
		}

		providesi, has := req.Params["providesAmount"]
		if !has {
			return errors.New("params missing providesAmount field")
		}

		providesAmount, ok := providesi.(float64)
		if !ok {
			return fmt.Errorf("failed to cast providesAmount parameter to float64: got %T", providesi)
		}

		id, ch, err := s.ns.takeOffer(maddr, offerID, providesAmount)
		if err != nil {
			return err
		}

		return s.subscribeTakeOffer(s.ctx, conn, id, ch)
	default:
		return errors.New("invalid method")
	}
}

func (s *wsServer) subscribeTakeOffer(ctx context.Context, conn *websocket.Conn,
	id uint64, statusCh <-chan types.Status) error {
	// firstly write swap ID
	idMsg := map[string]uint64{
		"id": id,
	}

	if err := writeResponse(conn, idMsg); err != nil {
		return err
	}

	for {
		select {
		case status, ok := <-statusCh:
			if !ok {
				return nil
			}

			resp := &SubscribeSwapStatusResponse{
				Stage: status.String(),
			}

			if err := writeResponse(conn, resp); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// subscribeSwapStatus writes the swap's stage to the connection every time it updates.
// when the swap completes, it writes the final status then closes the connection.
// example: `{"jsonrpc":"2.0", "method":"swap_subscribeStatus", "params": {"id": 0}, "id": 0}`
func (s *wsServer) subscribeSwapStatus(ctx context.Context, conn *websocket.Conn, id uint64) error {
	info := s.sm.GetOngoingSwap()
	if info == nil {
		return s.writeSwapExitStatus(conn, id)
	}

	statusCh := info.StatusCh()
	for {
		select {
		case status, ok := <-statusCh:
			if !ok {
				return nil
			}

			resp := &SubscribeSwapStatusResponse{
				Stage: status.String(),
			}

			if err := writeResponse(conn, resp); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *wsServer) writeSwapExitStatus(conn *websocket.Conn, id uint64) error {
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

func writeResponse(conn *websocket.Conn, result interface{}) error {
	bz, err := json.Marshal(result)
	if err != nil {
		return err
	}

	resp := &Response{
		Version: rpcclient.DefaultJSONRPCVersion,
		Result:  bz,
	}

	return conn.WriteJSON(resp)
}

func writeError(conn *websocket.Conn, err error) error {
	resp := &Response{
		Version: rpcclient.DefaultJSONRPCVersion,
		Error: &rpcclient.Error{
			Message: err.Error(),
		},
	}

	return conn.WriteJSON(resp)
}
