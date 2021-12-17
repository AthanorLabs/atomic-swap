package rpc

import (
	"errors"
	"net/http"

	mcrypto "github.com/noot/atomic-swap/monero/crypto"
)

// MoneroRecoverer is implemented by a backend which is able to recover monero
type MoneroRecoverer interface {
	WalletFromSecrets(aliceSecret, bobSecret string) (mcrypto.Address, error)
}

// RecoverService is the RPC service prefixed by recover_.
type RecoverService struct {
	mr MoneroRecoverer
}

// NewRecoverService ...
func NewRecoverService(mr MoneroRecoverer) *RecoverService {
	return &RecoverService{
		mr: mr,
	}
}

// RecoverMoneroRequest is used as input to recover_monero.
// 2/3 of the parameters must be provided for recovery to occur.
type RecoverMoneroRequest struct {
	AliceSecret     string `json:"alice_secret"`
	BobSecret       string `json:"bob_secret"`
	ContractAddress string `json:"contract_address"`
}

// RecoverMoneroResponse contains the address of the recovered wallet.
type RecoverMoneroResponse struct {
	Address string `json:"monero_address"`
}

// Monero attempts to recover a monero wallet from a swap that failed or exited early.
func (s *RecoverService) Monero(_ *http.Request, req *RecoverMoneroRequest, resp *RecoverMoneroResponse) error {
	switch {
	case req.AliceSecret != "" && req.BobSecret != "":
		addr, err := s.mr.WalletFromSecrets(req.AliceSecret, req.BobSecret)
		if err != nil {
			return err
		}

		resp.Address = string(addr)
	default:
		return errors.New("must provide 2/3 parameters")
	}

	return nil
}
