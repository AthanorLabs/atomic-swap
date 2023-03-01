package rpc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/common/vjson"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: checkOriginFunc,
}

func checkOriginFunc(_ *http.Request) bool {
	return true
}

type wsServer struct {
	ctx     context.Context
	sm      SwapManager
	ns      *NetService
	backend ProtocolBackend
	taker   XMRTaker
}

func newWsServer(ctx context.Context, sm SwapManager, ns *NetService, backend ProtocolBackend,
	taker XMRTaker) *wsServer {
	s := &wsServer{
		ctx:     ctx,
		sm:      sm,
		ns:      ns,
		backend: backend,
		taker:   taker,
	}

	return s
}

// ServeHTTP ...
func (s *wsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warnf("failed to update connection to websockets: %s", err)
		return
	}

	defer func() { _ = conn.Close() }()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Debugf("failed to read websockets message: %s", err)
			break
		}

		req := new(rpctypes.Request)
		err = vjson.UnmarshalStruct(message, req)
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

func (s *wsServer) handleRequest(conn *websocket.Conn, req *rpctypes.Request) error {
	switch req.Method {
	case rpctypes.SubscribeSigner:
		params := new(rpctypes.SignerRequest)
		if err := vjson.UnmarshalStruct(req.Params, params); err != nil {
			return fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		return s.handleSigner(s.ctx, conn, params.OfferID, params.EthAddress, params.XMRAddress)
	case rpctypes.SubscribeNewPeer:
		return errUnimplemented
	case rpctypes.NetDiscover:
		params := new(rpctypes.DiscoverRequest)
		if err := vjson.UnmarshalStruct(req.Params, params); err != nil {
			return fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		resp := new(rpctypes.DiscoverResponse)
		err := s.ns.Discover(nil, params, resp)
		if err != nil {
			return err
		}

		return writeResponse(conn, resp)
	case rpctypes.NetQueryPeer:
		params := new(rpctypes.QueryPeerRequest)
		if err := vjson.UnmarshalStruct(req.Params, params); err != nil {
			return fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		resp := new(rpctypes.QueryPeerResponse)
		err := s.ns.QueryPeer(nil, params, resp)
		if err != nil {
			return err
		}

		return writeResponse(conn, resp)
	case rpctypes.SubscribeSwapStatus:
		params := new(rpctypes.SubscribeSwapStatusRequest)
		if err := vjson.UnmarshalStruct(req.Params, params); err != nil {
			return fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		return s.subscribeSwapStatus(s.ctx, conn, params.OfferID)
	case rpctypes.SubscribeTakeOffer:
		params := new(rpctypes.TakeOfferRequest)
		if err := vjson.UnmarshalStruct(req.Params, params); err != nil {
			return fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		ch, err := s.ns.takeOffer(params.PeerID, params.OfferID, params.ProvidesAmount)
		if err != nil {
			return err
		}

		return s.subscribeTakeOffer(s.ctx, conn, ch)
	case rpctypes.SubscribeMakeOffer:
		params := new(rpctypes.MakeOfferRequest)
		if err := vjson.UnmarshalStruct(req.Params, params); err != nil {
			return fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		offerResp, offerExtra, err := s.ns.makeOffer(params)
		if err != nil {
			return err
		}

		return s.subscribeMakeOffer(s.ctx, conn, offerResp.OfferID, offerExtra)
	default:
		return errInvalidMethod
	}
}

func (s *wsServer) handleSigner(
	ctx context.Context,
	conn *websocket.Conn,
	offerID types.Hash,
	ethAddress ethcommon.Address,
	xmrAddr *mcrypto.Address,
) error {
	signer, err := s.taker.ExternalSender(offerID)
	if err != nil {
		return err
	}

	if err = xmrAddr.ValidateEnv(s.backend.Env()); err != nil {
		return err
	}

	s.backend.ETHClient().SetAddress(ethAddress)
	s.backend.SetXMRDepositAddress(xmrAddr, offerID)
	defer s.backend.ClearXMRDepositAddress(offerID)

	txsOutCh := signer.OngoingCh(offerID)
	txsInCh := signer.IncomingCh(offerID)

	var timeout time.Duration
	switch s.backend.Env() {
	case common.Mainnet, common.Stagenet:
		timeout = time.Hour
	case common.Development:
		timeout = time.Minute * 5
	}

	for {
		select {
		// TODO: check if conn closes or swap exited (#165)
		case <-time.After(timeout):
			return fmt.Errorf("signer timed out")
		case <-ctx.Done():
			return nil
		case tx := <-txsOutCh:
			log.Debugf("outbound tx: %v", tx)
			resp := &rpctypes.SignerResponse{
				OfferID: offerID,
				To:      tx.To,
				Data:    tx.Data,
				Value:   tx.Value,
			}

			err := conn.WriteJSON(resp)
			if err != nil {
				return err
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				return err
			}

			params := new(rpctypes.SignerTxSigned)
			if err := vjson.UnmarshalStruct(message, &params); err != nil {
				return fmt.Errorf("failed to unmarshal parameters: %w", err)
			}

			if params.OfferID != offerID {
				return fmt.Errorf("got unexpected offerID %s, expected %s", params.OfferID, offerID)
			}

			txsInCh <- params.TxHash
		}
	}
}

func (s *wsServer) subscribeTakeOffer(ctx context.Context, conn *websocket.Conn,
	statusCh <-chan types.Status) error {
	for {
		select {
		case status, ok := <-statusCh:
			if !ok {
				return nil
			}

			resp := &rpctypes.SubscribeSwapStatusResponse{
				Status: status,
			}

			if err := writeResponse(conn, resp); err != nil {
				return err
			}

			if !status.IsOngoing() {
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *wsServer) subscribeMakeOffer(ctx context.Context, conn *websocket.Conn,
	offerID types.Hash, offerExtra *types.OfferExtra) error {
	resp := &rpctypes.MakeOfferResponse{
		PeerID:  s.ns.net.PeerID(),
		OfferID: offerID,
	}

	if err := writeResponse(conn, resp); err != nil {
		return err
	}

	for {
		select {
		case status, ok := <-offerExtra.StatusCh:
			if !ok {
				return nil
			}

			resp := &rpctypes.SubscribeSwapStatusResponse{
				Status: status,
			}

			if err := writeResponse(conn, resp); err != nil {
				return err
			}

			if !status.IsOngoing() {
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// subscribeSwapStatus writes the swap's stage to the connection every time it updates.
// when the swap completes, it writes the final status then closes the connection.
// example: `{"jsonrpc":"2.0", "method":"swap_subscribeStatus", "params": {"id": 0}, "id": 0}`
func (s *wsServer) subscribeSwapStatus(ctx context.Context, conn *websocket.Conn, id types.Hash) error {
	// we can ignore the error here, since the error will only be if the swap cannot be found
	// as ongoing, in which case `writeSwapExitStatus` will look for it in the past swaps.
	info, err := s.sm.GetOngoingSwap(id)
	if err != nil {
		return s.writeSwapExitStatus(conn, id)
	}

	statusCh := info.StatusCh()
	for {
		select {
		case status, ok := <-statusCh:
			if !ok {
				return nil
			}

			resp := &rpctypes.SubscribeSwapStatusResponse{
				Status: status,
			}

			if err := writeResponse(conn, resp); err != nil {
				return err
			}

			if !status.IsOngoing() {
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *wsServer) writeSwapExitStatus(conn *websocket.Conn, id types.Hash) error {
	info, err := s.sm.GetPastSwap(id)
	if err != nil {
		return err
	}

	resp := &rpctypes.SubscribeSwapStatusResponse{
		Status: info.Status,
	}

	if err := writeResponse(conn, resp); err != nil {
		return err
	}

	return nil
}

func writeResponse(conn *websocket.Conn, result interface{}) error {
	bz, err := vjson.MarshalStruct(result)
	if err != nil {
		return err
	}

	resp := &rpctypes.Response{
		Version: rpctypes.DefaultJSONRPCVersion,
		Result:  bz,
	}

	return conn.WriteJSON(resp)
}

func writeError(conn *websocket.Conn, err error) error {
	resp := &rpctypes.Response{
		Version: rpctypes.DefaultJSONRPCVersion,
		Error: &rpctypes.Error{
			Message: err.Error(),
		},
	}

	return conn.WriteJSON(resp)
}
