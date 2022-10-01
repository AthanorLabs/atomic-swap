package rpc

import (
	"net/http"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
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

// Balances returns combined information of both the Monero and Ethereum account addresses
// and balances.
func (s *PersonalService) Balances(_ *http.Request, _ *interface{}, resp *rpctypes.BalancesResponse) error {
	mAddr, mBal, err := s.xmrmaker.GetMoneroBalance()
	if err != nil {
		return err
	}

	eAddr, eBal, err := s.pb.EthBalance()
	if err != nil {
		return err
	}

	*resp = rpctypes.BalancesResponse{
		MoneroAddress:           mAddr,
		PiconeroBalance:         common.MoneroAmount(mBal.Balance),
		PiconeroUnlockedBalance: common.MoneroAmount(mBal.UnlockedBalance),
		BlocksToUnlock:          mBal.BlocksToUnlock,
		EthAddress:              eAddr.String(),
		EthBalance:              eBal,
	}
	return nil
}
