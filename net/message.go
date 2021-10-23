package net

import (
	"encoding/json"
	"fmt"
)

type Message interface {
	String() string
	Encode() ([]byte, error)
}

type WantMessage struct {
	Want string
}

func (m *WantMessage) String() string {
	return fmt.Sprintf("WantMessage Want=%s", m.Want)
}

func (m *WantMessage) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// SendKeysMessage is sent by both parties to each other to initiate the protocol
type SendKeysMessage struct {
	PublicSpendKey string
	PrivateViewKey string
}

func (m *SendKeysMessage) String() string {
	return fmt.Sprintf("SendKeysMessage PublicSpendKey=%s PrivateViewKey=%v",
		m.PublicSpendKey,
		m.PrivateViewKey == "",
	)
}

func (m *SendKeysMessage) Encode() ([]byte, error) {
	return json.Marshal(m)
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
	return json.Marshal(m)
}

// NotifyXMRLock is sent by Bob to Alice after locking his XMR.
type NotifyXMRLock struct {
	Address string
}

func (m *NotifyXMRLock) String() string {
	return "NotifyXMRLock"
}

func (m *NotifyXMRLock) Encode() ([]byte, error) {
	return json.Marshal(m)
}
