package rpc

import (
	"net/http"
	"time"
)

// PersonalService handles private keys and wallets.
type PersonalService struct {
	xmrmaker XMRMaker
	pb       ProtocolBackend
}

// NewPersonalService ...
func NewPersonalService(xmrmaker XMRMaker, pb ProtocolBackend) *PersonalService {
	return &PersonalService{
		xmrmaker: xmrmaker,
		pb:       pb,
	}
}

// SetMoneroWalletFileRequest ...
type SetMoneroWalletFileRequest struct {
	WalletFile     string `json:"walletFile"`
	WalletPassword string `json:"password"`
}

// SetMoneroWalletFile opens the given wallet file in monero-wallet-rpc.
// It must exist in the monero-wallet-rpc wallet-dir that was specified on its startup.
func (s *PersonalService) SetMoneroWalletFile(_ *http.Request, req *SetMoneroWalletFileRequest, _ *interface{}) error {
	return s.xmrmaker.SetMoneroWalletFile(req.WalletFile, req.WalletPassword)
}

// SetSwapTimeoutRequest ...
type SetSwapTimeoutRequest struct {
	Timeout uint64 `json:"timeout"` // timeout in seconds
}

// SetSwapTimeout ...
func (s *PersonalService) SetSwapTimeout(_ *http.Request, req *SetSwapTimeoutRequest, _ *interface{}) error {
	timeout := time.Second * time.Duration(req.Timeout)
	s.pb.SetSwapTimeout(timeout)
	return nil
}

// SetGasPriceRequest ...
type SetGasPriceRequest struct {
	GasPrice uint64
}

// SetGasPrice sets the gas price (in wei) to be used for ethereum transactions.
func (s *PersonalService) SetGasPrice(_ *http.Request, req *SetGasPriceRequest, _ *interface{}) error {
	s.pb.SetGasPrice(req.GasPrice)
	s.pb.SetGasPrice(req.GasPrice)
	return nil
}
