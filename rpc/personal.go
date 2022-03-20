package rpc

import (
	"net/http"
	"time"
)

// PersonalService handles private keys and wallets.
type PersonalService struct {
	alice Alice
	bob   Bob
}

// NewPersonalService ...
func NewPersonalService(alice Alice, bob Bob) *PersonalService {
	return &PersonalService{
		alice: alice,
		bob:   bob,
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
	return s.bob.SetMoneroWalletFile(req.WalletFile, req.WalletPassword)
}

type SetSwapTimeoutRequest struct {
	Timeout uint64 `json:"timeout"` // timeout in seconds
}

func (s *PersonalService) SetSwapTimeout(_ *http.Request, req *SetSwapTimeoutRequest, _ *interface{}) error {
	timeout := time.Second * time.Duration(req.Timeout)
	s.alice.SetSwapTimeout(timeout)
	return nil
}
