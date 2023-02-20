// Package message provides the types for messages that are sent between swapd instances.
package message

import (
	"errors"
	"fmt"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/common/vjson"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/crypto/secp256k1"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
)

// Identifiers for our p2p message types. The first byte of a message has the
// identifier below telling us which type to decode the JSON message as.
const (
	QueryResponseType byte = iota
	SendKeysType
	NotifyETHLockedType
	NotifyXMRLockType
)

// TypeToString converts a message type into a string.
func TypeToString(t byte) string {
	switch t {
	case QueryResponseType:
		return "QueryResponse"
	case SendKeysType:
		return "SendKeysMessage"
	case NotifyETHLockedType:
		return "NotifyETHLocked"
	case NotifyXMRLockType:
		return "NotifyXMRLock"
	default:
		return "unknown"
	}
}

// DecodeMessage decodes the given bytes into a Message
func DecodeMessage(b []byte) (common.Message, error) {
	// 1-byte type followed by at least 2-bytes of JSON (`{}`)
	if len(b) < 3 {
		return nil, errors.New("invalid message bytes")
	}

	msgType := b[0]
	msgJSON := b[1:]
	var msg common.Message

	switch msgType {
	case QueryResponseType:
		msg = &QueryResponse{}
	case SendKeysType:
		msg = &SendKeysMessage{}
	case NotifyETHLockedType:
		msg = &NotifyETHLocked{}
	case NotifyXMRLockType:
		msg = &NotifyXMRLock{}
	default:
		return nil, fmt.Errorf("invalid message type=%d", msgType)
	}

	if err := vjson.UnmarshalStruct(msgJSON, msg); err != nil {
		return nil, fmt.Errorf("failed to decode %s message: %w", TypeToString(msg.Type()), err)
	}
	return msg, nil
}

// QueryResponse ...
type QueryResponse struct {
	Offers []*types.Offer `json:"offers" validate:"dive,required"`
}

// String ...
func (m *QueryResponse) String() string {
	return fmt.Sprintf("QueryResponse Offers=%v",
		m.Offers,
	)
}

// Encode ...
func (m *QueryResponse) Encode() ([]byte, error) {
	b, err := vjson.MarshalStruct(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{QueryResponseType}, b...), nil
}

// Type ...
func (m *QueryResponse) Type() byte {
	return QueryResponseType
}

// The below messages are swap protocol messages, exchanged after the swap has been agreed
// upon by both sides.

// SendKeysMessage is sent by both parties to each other to initiate the protocol
type SendKeysMessage struct {
	OfferID            types.Hash              `json:"offerID"`
	ProvidedAmount     *apd.Decimal            `json:"providedAmount" validate:"required"`
	PublicSpendKey     *mcrypto.PublicKey      `json:"publicSpendKey" validate:"required"`
	PrivateViewKey     *mcrypto.PrivateViewKey `json:"privateViewKey" validate:"required"`
	DLEqProof          string                  `json:"dleqProof" validate:"required"`
	Secp256k1PublicKey *secp256k1.PublicKey    `json:"secp256k1PublicKey" validate:"required"`
	EthAddress         ethcommon.Address       `json:"ethAddress"`
}

// String ...
func (m *SendKeysMessage) String() string {
	return fmt.Sprintf("SendKeysMessage OfferID=%s ProvidedAmount=%v PublicSpendKey=%s PrivateViewKey=%s DLEqProof=%s Secp256k1PublicKey=%s EthAddress=%s", //nolint:lll
		m.OfferID,
		m.ProvidedAmount,
		m.PublicSpendKey,
		m.PrivateViewKey,
		m.DLEqProof,
		m.Secp256k1PublicKey,
		m.EthAddress,
	)
}

// Encode ...
func (m *SendKeysMessage) Encode() ([]byte, error) {
	b, err := vjson.MarshalStruct(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{SendKeysType}, b...), nil
}

// Type ...
func (m *SendKeysMessage) Type() byte {
	return SendKeysType
}

// NotifyETHLocked is sent by XMRTaker to XMRMaker after deploying the swap contract
// and locking her ether in it
type NotifyETHLocked struct {
	Address        ethcommon.Address          `json:"address"`
	TxHash         types.Hash                 `json:"txHash"`
	ContractSwapID types.Hash                 `json:"contractSwapID"`
	ContractSwap   *contracts.SwapFactorySwap `json:"contractSwap" validate:"required"`
}

// String ...
func (m *NotifyETHLocked) String() string {
	return fmt.Sprintf("NotifyETHLocked Address=%s TxHash=%s ContractSwapID=%d ContractSwap=%v",
		m.Address,
		m.TxHash,
		m.ContractSwapID,
		m.ContractSwap,
	)
}

// Encode ...
func (m *NotifyETHLocked) Encode() ([]byte, error) {
	b, err := vjson.MarshalStruct(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{NotifyETHLockedType}, b...), nil
}

// Type ...
func (m *NotifyETHLocked) Type() byte {
	return NotifyETHLockedType
}

// NotifyXMRLock is sent by XMRMaker to XMRTaker after locking his XMR.
type NotifyXMRLock struct {
	Address string     `json:"address" validate:"required"` // address the monero was sent to
	TxID    types.Hash `json:"txID"`                        // Monero transaction ID (transaction hash in hex)
}

// String ...
func (m *NotifyXMRLock) String() string {
	return "NotifyXMRLock"
}

// Encode ...
func (m *NotifyXMRLock) Encode() ([]byte, error) {
	b, err := vjson.MarshalStruct(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{NotifyXMRLockType}, b...), nil
}

// Type ...
func (m *NotifyXMRLock) Type() byte {
	return NotifyXMRLockType
}
