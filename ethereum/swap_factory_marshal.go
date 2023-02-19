package contracts

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/common/vjson"
)

// swap is the same as the auto-generated SwapFactorySwap type, but with some type
// adjustments and annotations for JSON marshalling.
type swap struct {
	Owner        common.Address `json:"owner"`
	Claimer      common.Address `json:"claimer"`
	PubKeyClaim  types.Hash     `json:"pub_key_claim"`
	PubKeyRefund types.Hash     `json:"pub_key_refund"`
	Timeout0     *big.Int       `json:"timeout0" validate:"required"`
	Timeout1     *big.Int       `json:"timeout1" validate:"required"`
	Asset        common.Address `json:"asset"`
	Value        *big.Int       `json:"value" validate:"required"`
	Nonce        *big.Int       `json:"nonce" validate:"required"`
}

// MarshalJSON provides JSON marshalling for SwapFactorySwap
func (sfs *SwapFactorySwap) MarshalJSON() ([]byte, error) {
	return vjson.MarshalStruct(&swap{
		Owner:        sfs.Owner,
		Claimer:      sfs.Claimer,
		PubKeyClaim:  sfs.PubKeyClaim,
		PubKeyRefund: sfs.PubKeyRefund,
		Timeout0:     sfs.Timeout0,
		Timeout1:     sfs.Timeout1,
		Asset:        sfs.Asset,
		Value:        sfs.Value,
		Nonce:        sfs.Nonce,
	})
}

// UnmarshalJSON provides JSON unmarshalling for SwapFactorySwap
func (sfs *SwapFactorySwap) UnmarshalJSON(data []byte) error {
	s := &swap{}
	if err := vjson.UnmarshalStruct(data, s); err != nil {
		return err
	}
	*sfs = SwapFactorySwap{
		Owner:        s.Owner,
		Claimer:      s.Claimer,
		PubKeyClaim:  s.PubKeyClaim,
		PubKeyRefund: s.PubKeyRefund,
		Timeout0:     s.Timeout0,
		Timeout1:     s.Timeout1,
		Asset:        s.Asset,
		Value:        s.Value,
		Nonce:        s.Nonce,
	}
	return nil
}
