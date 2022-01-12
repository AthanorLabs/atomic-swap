package rpc

import (
	"net/http"
)

// PersonalService handles private keys and wallets.
type PersonalService struct {
	bob Bob
}

// NewPersonalService ...
func NewPersonalService(bob Bob) *PersonalService {
	return &PersonalService{
		bob: bob,
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
