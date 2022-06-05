package txsender

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type Sender interface {
	NewSwap() (ethcommon.Hash, ethtypes.Receipt, error)
	Ready() (ethcommon.Hash, ethtypes.Receipt, error)
	Claim() (ethcommon.Hash, ethtypes.Receipt, error)
	Refund() (ethcommon.Hash, ethtypes.Receipt, error)
}

type privateKeySender struct {
	contract *swapfactory.SwapFactory
	txOpts   *bind.TransactOpts
}

func NewSenderWithPrivateKey(contract *swapfactory.SwapFactory, txOpts *bind.TransactOpts) Sender {
	return &Sender{
		contract: contract,
		txOpts:   txOpts,
	}
}
