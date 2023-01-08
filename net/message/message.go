// Package message provides the types for messages that are sent between swapd instances.
package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common/types"
)

const (
	QueryResponseType byte = iota //nolint
	SendKeysType
	NotifyETHLockedType
	NotifyXMRLockType
	NilType
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

// Message must be implemented by all network messages
type Message interface {
	String() string
	Encode() ([]byte, error)
	Type() byte
}

// DecodeMessage decodes the given bytes into a Message
func DecodeMessage(b []byte) (Message, error) {
	// 1-byte type followed by at least 2-bytes of JSON (`{}`)
	if len(b) < 3 {
		return nil, errors.New("invalid message bytes")
	}

	msgType := b[0]
	msgJSON := b[1:]
	var msg Message

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

	if err := json.Unmarshal(msgJSON, &msg); err != nil {
		return nil, fmt.Errorf("failed to decode %s message: %w", TypeToString(msg.Type()), err)
	}
	return msg, nil
}

// QueryResponse ...
type QueryResponse struct {
	Offers []*types.Offer
}

// String ...
func (m *QueryResponse) String() string {
	return fmt.Sprintf("QueryResponse Offers=%v",
		m.Offers,
	)
}

// Encode ...
func (m *QueryResponse) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
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
	OfferID            types.Hash
	ProvidedAmount     float64
	PublicSpendKey     string
	PublicViewKey      string
	PrivateViewKey     string
	DLEqProof          string
	Secp256k1PublicKey string
	EthAddress         string
}

// String ...
func (m *SendKeysMessage) String() string {
	return fmt.Sprintf("SendKeysMessage OfferID=%s ProvidedAmount=%v PublicSpendKey=%s PublicViewKey=%s PrivateViewKey=%s DLEqProof=%s Secp256k1PublicKey=%s EthAddress=%s", //nolint:lll
		m.OfferID,
		m.ProvidedAmount,
		m.PublicSpendKey,
		m.PublicViewKey,
		m.PrivateViewKey,
		m.DLEqProof,
		m.Secp256k1PublicKey,
		m.EthAddress,
	)
}

// Encode ...
func (m *SendKeysMessage) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{SendKeysType}, b...), nil
}

// Type ...
func (m *SendKeysMessage) Type() byte {
	return SendKeysType
}

// ContractSwap is the same as contracts.SwapFactorySwap
type ContractSwap struct {
	Owner        ethcommon.Address
	Claimer      ethcommon.Address
	PubKeyClaim  [32]byte
	PubKeyRefund [32]byte
	Timeout0     *big.Int
	Timeout1     *big.Int
	Asset        ethcommon.Address
	Value        *big.Int
	Nonce        *big.Int
}

// NotifyETHLocked is sent by XMRTaker to XMRMaker after deploying the swap contract
// and locking her ether in it
type NotifyETHLocked struct {
	Address        string
	TxHash         string
	ContractSwapID types.Hash
	ContractSwap   *ContractSwap
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
	b, err := json.Marshal(m)
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
	Address string // address the monero was sent to
	TxID    string // Monero transaction ID (transaction hash in hex)
}

// String ...
func (m *NotifyXMRLock) String() string {
	return "NotifyXMRLock"
}

// Encode ...
func (m *NotifyXMRLock) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{NotifyXMRLockType}, b...), nil
}

// Type ...
func (m *NotifyXMRLock) Type() byte {
	return NotifyXMRLockType
}
