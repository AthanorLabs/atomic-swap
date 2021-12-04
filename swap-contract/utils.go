package swap

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
)

// GetSecretFromLog returns the secret from a Claim or Refund log
func GetSecretFromLog(log *ethtypes.Log, event string) (*monero.PrivateSpendKey, error) {
	if event != "Refunded" && event != "Claimed" {
		return nil, errors.New("invalid event name, must be one of Claimed or Refunded")
	}

	abi, err := abi.JSON(strings.NewReader(SwapABI))
	if err != nil {
		return nil, err
	}

	data := log.Data
	res, err := abi.Unpack(event, data)
	if err != nil {
		return nil, err
	}

	s := res[0].([32]byte)

	sk, err := monero.NewPrivateSpendKey(common.Reverse(s[:]))
	if err != nil {
		return nil, err
	}

	return sk, nil
}
