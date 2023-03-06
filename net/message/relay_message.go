package message

import (
	"fmt"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/common/vjson"
)

// RelayClaimRequest implements common.Message for our p2p relay claim requests
type RelayClaimRequest struct {
	ClaimerAddress    ethcommon.Address `json:"claimerAddress" validate:"required"`
	SFContractAddress ethcommon.Address `json:"sfContractAddress" validate:"required"`
	Gas               *big.Int          `json:"gas" validate:"required"`
	Nonce             *big.Int          `json:"nonce" validate:"required"`
	Data              []byte            `json:"data" validate:"required"`
	Signature         []byte            `json:"signature" validate:"required"`
	ValidUntilTime    *big.Int          `json:"validUntilTime" validate:"required"`
	DomainSeparator   types.Hash        `json:"domainSeparator" validate:"required"`
	RequestTypeHash   types.Hash        `json:"requestTypeHash" validate:"required"`
	SuffixData        []byte            `json:"suffixData,omitempty"`
}

// RelayClaimResponse implements common.Message for our p2p relay claim responses
type RelayClaimResponse struct {
	TxHash ethcommon.Hash `json:"transactionHash" validate:"required"`
}

// String converts the RelayClaimRequest to a string usable for debugging purposes
func (m *RelayClaimRequest) String() string {
	return fmt.Sprintf("RelayClaimResponse=%#v", m)
}

// Encode implements the Encode() method of the common.Message interface which
// prepends a message type byte before the message's JSON encoding.
func (m *RelayClaimRequest) Encode() ([]byte, error) {
	b, err := vjson.MarshalStruct(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{RelayClaimRequestType}, b...), nil
}

// Type implements the Type() method of the common.Message interface
func (m *RelayClaimRequest) Type() byte {
	return RelayClaimRequestType
}

// String converts the RelayClaimRequest to a string usable for debugging purposes
func (m *RelayClaimResponse) String() string {
	return fmt.Sprintf("RelayClaimResponse=%#v", m)
}

// Encode implements the Encode() method of the common.Message interface which
// prepends a message type byte before the message's JSON encoding.
func (m *RelayClaimResponse) Encode() ([]byte, error) {
	b, err := vjson.MarshalStruct(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{RelayClaimResponseType}, b...), nil
}

// Type implements the Type() method of the common.Message interface
func (m *RelayClaimResponse) Type() byte {
	return RelayClaimResponseType
}
