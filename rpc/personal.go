package rpc

import (
	"net/http"
	"time"
)

// PersonalService handles private keys and wallets.
type PersonalService struct {
	xmrtaker XMRTaker
	xmrmaker XMRMaker
}

// NewPersonalService ...
func NewPersonalService(xmrtaker XMRTaker, xmrmaker XMRMaker) *PersonalService {
	return &PersonalService{
		xmrtaker: xmrtaker,
		xmrmaker: xmrmaker,
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
	s.xmrtaker.SetSwapTimeout(timeout)
	return nil
}
