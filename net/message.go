package net

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/noot/atomic-swap/common"
)

const (
	QueryResponseType byte = iota
	InitiateMessageType
	SendKeysMessageType
	NotifyContractDeployedType
	NotifyXMRLockType
	NotifyReadyType
	NotifyClaimedType
	NotifyRefundType
)

type Message interface {
	String() string
	Encode() ([]byte, error)
	Type() byte
}

func decodeMessage(b []byte) (Message, error) {
	if len(b) == 0 {
		return nil, errors.New("invalid message bytes")
	}

	switch b[0] {
	case QueryResponseType:
		var m *QueryResponse
		if err := json.Unmarshal(b[1:], &m); err != nil {
			return nil, err
		}
		return m, nil
	case InitiateMessageType:
		var m *InitiateMessage
		if err := json.Unmarshal(b[1:], &m); err != nil {
			return nil, err
		}
		return m, nil
	case SendKeysMessageType:
		var m *SendKeysMessage
		if err := json.Unmarshal(b[1:], &m); err != nil {
			return nil, err
		}
		return m, nil
	case NotifyContractDeployedType:
		var m *NotifyContractDeployed
		if err := json.Unmarshal(b[1:], &m); err != nil {
			return nil, err
		}
		return m, nil
	case NotifyXMRLockType:
		var m *NotifyXMRLock
		if err := json.Unmarshal(b[1:], &m); err != nil {
			return nil, err
		}
		return m, nil
	case NotifyReadyType:
		var m *NotifyReady
		if err := json.Unmarshal(b[1:], &m); err != nil {
			return nil, err
		}
		return m, nil
	case NotifyClaimedType:
		var m *NotifyClaimed
		if err := json.Unmarshal(b[1:], &m); err != nil {
			return nil, err
		}
		return m, nil
	default:
		return nil, errors.New("invalid message type")
	}
}

type QueryResponse struct {
	Provides      []common.ProvidesCoin
	MaximumAmount []float64
	ExchangeRate  common.ExchangeRate
}

func (m *QueryResponse) String() string {
	return fmt.Sprintf("QueryResponse Provides=%v MaximumAmount=%v ExchangeRate=%v",
		m.Provides,
		m.MaximumAmount,
		m.ExchangeRate,
	)
}

func (m *QueryResponse) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{QueryResponseType}, b...), nil
}

func (m *QueryResponse) Type() byte {
	return QueryResponseType
}

type InitiateMessage struct {
	Provides       common.ProvidesCoin
	ProvidesAmount float64
	DesiredAmount  float64
	*SendKeysMessage
}

func (m *InitiateMessage) String() string {
	return fmt.Sprintf("InitiateMessage Provides=%v ProvidesAmount=%v DesiredAmount=%v Keys=%s",
		m.Provides,
		m.ProvidesAmount,
		m.DesiredAmount,
		m.SendKeysMessage,
	)
}

func (m *InitiateMessage) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{InitiateMessageType}, b...), nil
}

func (m *InitiateMessage) Type() byte {
	return InitiateMessageType
}

// The below messages are sawp protocol messages, exchanged after the swap has been agreed
// upon by both sides.

// SendKeysMessage is sent by both parties to each other to initiate the protocol
type SendKeysMessage struct {
	PublicSpendKey string
	PublicViewKey  string
	PrivateViewKey string
	EthAddress     string
}

func (m *SendKeysMessage) String() string {
	return fmt.Sprintf("SendKeysMessage PublicSpendKey=%s PublicViewKey=%s PrivateViewKey=%s EthAddress=%s",
		m.PublicSpendKey,
		m.PublicViewKey,
		m.PrivateViewKey,
		m.EthAddress,
	)
}

func (m *SendKeysMessage) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{SendKeysMessageType}, b...), nil
}

func (m *SendKeysMessage) Type() byte {
	return SendKeysMessageType
}

// NotifyContractDeployed is sent by Alice to Bob after deploying the swap contract
// and locking her ether in it
type NotifyContractDeployed struct {
	Address string
}

func (m *NotifyContractDeployed) String() string {
	return "NotifyContractDeployed"
}

func (m *NotifyContractDeployed) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{NotifyContractDeployedType}, b...), nil
}

func (m *NotifyContractDeployed) Type() byte {
	return NotifyContractDeployedType
}

// NotifyXMRLock is sent by Bob to Alice after locking his XMR.
type NotifyXMRLock struct {
	Address string
}

func (m *NotifyXMRLock) String() string {
	return "NotifyXMRLock"
}

func (m *NotifyXMRLock) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{NotifyXMRLockType}, b...), nil
}

func (m *NotifyXMRLock) Type() byte {
	return NotifyXMRLockType
}

// NotifyReady is sent by Alice to Bob after calling Ready() on the contract.
type NotifyReady struct{}

func (m *NotifyReady) String() string {
	return "NotifyReady"
}

func (m *NotifyReady) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{NotifyReadyType}, b...), nil
}

func (m *NotifyReady) Type() byte {
	return NotifyReadyType
}

// NotifyClaimed is sent by Bob to Alice after claiming his ETH.
type NotifyClaimed struct {
	TxHash string
}

func (m *NotifyClaimed) String() string {
	return fmt.Sprintf("NotifyClaimed %s", m.TxHash)
}

func (m *NotifyClaimed) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{NotifyClaimedType}, b...), nil
}

func (m *NotifyClaimed) Type() byte {
	return NotifyClaimedType
}

// NotifyRefund is sent by Alice to Bob after calling Refund() on the contract.
type NotifyRefund struct {
	TxHash string
}

func (m *NotifyRefund) String() string {
	return fmt.Sprintf("NotifyClaimed %s", m.TxHash)
}

func (m *NotifyRefund) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{NotifyRefundType}, b...), nil
}

func (m *NotifyRefund) Type() byte {
	return NotifyRefundType
}
