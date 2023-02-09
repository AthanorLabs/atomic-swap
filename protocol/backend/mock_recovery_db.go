// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/athanorlabs/atomic-swap/protocol/backend (interfaces: RecoveryDB)

// Package backend is a generated GoMock package.
package backend

import (
	reflect "reflect"

	common "github.com/ethereum/go-ethereum/common"
	gomock "github.com/golang/mock/gomock"

	types "github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	db "github.com/athanorlabs/atomic-swap/db"
)

// MockRecoveryDB is a mock of RecoveryDB interface.
type MockRecoveryDB struct {
	ctrl     *gomock.Controller
	recorder *MockRecoveryDBMockRecorder
}

// MockRecoveryDBMockRecorder is the mock recorder for MockRecoveryDB.
type MockRecoveryDBMockRecorder struct {
	mock *MockRecoveryDB
}

// NewMockRecoveryDB creates a new mock instance.
func NewMockRecoveryDB(ctrl *gomock.Controller) *MockRecoveryDB {
	mock := &MockRecoveryDB{ctrl: ctrl}
	mock.recorder = &MockRecoveryDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRecoveryDB) EXPECT() *MockRecoveryDBMockRecorder {
	return m.recorder
}

// DeleteSwap mocks base method.
func (m *MockRecoveryDB) DeleteSwap(arg0 common.Hash) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSwap", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSwap indicates an expected call of DeleteSwap.
func (mr *MockRecoveryDBMockRecorder) DeleteSwap(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSwap", reflect.TypeOf((*MockRecoveryDB)(nil).DeleteSwap), arg0)
}

// GetContractSwapInfo mocks base method.
func (m *MockRecoveryDB) GetContractSwapInfo(arg0 common.Hash) (*db.EthereumSwapInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContractSwapInfo", arg0)
	ret0, _ := ret[0].(*db.EthereumSwapInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContractSwapInfo indicates an expected call of GetContractSwapInfo.
func (mr *MockRecoveryDBMockRecorder) GetContractSwapInfo(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractSwapInfo", reflect.TypeOf((*MockRecoveryDB)(nil).GetContractSwapInfo), arg0)
}

// GetCounterpartySwapPrivateKey mocks base method.
func (m *MockRecoveryDB) GetCounterpartySwapPrivateKey(arg0 common.Hash) (*mcrypto.PrivateSpendKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCounterpartySwapPrivateKey", arg0)
	ret0, _ := ret[0].(*mcrypto.PrivateSpendKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCounterpartySwapPrivateKey indicates an expected call of GetCounterpartySwapPrivateKey.
func (mr *MockRecoveryDBMockRecorder) GetCounterpartySwapPrivateKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCounterpartySwapPrivateKey", reflect.TypeOf((*MockRecoveryDB)(nil).GetCounterpartySwapPrivateKey), arg0)
}

// GetSwapPrivateKey mocks base method.
func (m *MockRecoveryDB) GetSwapPrivateKey(arg0 common.Hash) (*mcrypto.PrivateSpendKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSwapPrivateKey", arg0)
	ret0, _ := ret[0].(*mcrypto.PrivateSpendKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSwapPrivateKey indicates an expected call of GetSwapPrivateKey.
func (mr *MockRecoveryDBMockRecorder) GetSwapPrivateKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSwapPrivateKey", reflect.TypeOf((*MockRecoveryDB)(nil).GetSwapPrivateKey), arg0)
}

// GetSwapRelayerInfo mocks base method.
func (m *MockRecoveryDB) GetSwapRelayerInfo(arg0 common.Hash) (*types.OfferExtra, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSwapRelayerInfo", arg0)
	ret0, _ := ret[0].(*types.OfferExtra)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSwapRelayerInfo indicates an expected call of GetSwapRelayerInfo.
func (mr *MockRecoveryDBMockRecorder) GetSwapRelayerInfo(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSwapRelayerInfo", reflect.TypeOf((*MockRecoveryDB)(nil).GetSwapRelayerInfo), arg0)
}

// GetXMRMakerSwapKeys mocks base method.
func (m *MockRecoveryDB) GetXMRMakerSwapKeys(arg0 common.Hash) (*mcrypto.PublicKey, *mcrypto.PrivateViewKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetXMRMakerSwapKeys", arg0)
	ret0, _ := ret[0].(*mcrypto.PublicKey)
	ret1, _ := ret[1].(*mcrypto.PrivateViewKey)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetXMRMakerSwapKeys indicates an expected call of GetXMRMakerSwapKeys.
func (mr *MockRecoveryDBMockRecorder) GetXMRMakerSwapKeys(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetXMRMakerSwapKeys", reflect.TypeOf((*MockRecoveryDB)(nil).GetXMRMakerSwapKeys), arg0)
}

// GetXMRTakerSwapKeys mocks base method.
func (m *MockRecoveryDB) GetXMRTakerSwapKeys(arg0 common.Hash) (*mcrypto.PublicKeyPair, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetXMRTakerSwapKeys", arg0)
	ret0, _ := ret[0].(*mcrypto.PublicKeyPair)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetXMRTakerSwapKeys indicates an expected call of GetXMRTakerSwapKeys.
func (mr *MockRecoveryDBMockRecorder) GetXMRTakerSwapKeys(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetXMRTakerSwapKeys", reflect.TypeOf((*MockRecoveryDB)(nil).GetXMRTakerSwapKeys), arg0)
}

// PutContractSwapInfo mocks base method.
func (m *MockRecoveryDB) PutContractSwapInfo(arg0 common.Hash, arg1 *db.EthereumSwapInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutContractSwapInfo", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutContractSwapInfo indicates an expected call of PutContractSwapInfo.
func (mr *MockRecoveryDBMockRecorder) PutContractSwapInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutContractSwapInfo", reflect.TypeOf((*MockRecoveryDB)(nil).PutContractSwapInfo), arg0, arg1)
}

// PutCounterpartySwapPrivateKey mocks base method.
func (m *MockRecoveryDB) PutCounterpartySwapPrivateKey(arg0 common.Hash, arg1 *mcrypto.PrivateSpendKey) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutCounterpartySwapPrivateKey", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutCounterpartySwapPrivateKey indicates an expected call of PutCounterpartySwapPrivateKey.
func (mr *MockRecoveryDBMockRecorder) PutCounterpartySwapPrivateKey(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutCounterpartySwapPrivateKey", reflect.TypeOf((*MockRecoveryDB)(nil).PutCounterpartySwapPrivateKey), arg0, arg1)
}

// PutSwapPrivateKey mocks base method.
func (m *MockRecoveryDB) PutSwapPrivateKey(arg0 common.Hash, arg1 *mcrypto.PrivateSpendKey) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutSwapPrivateKey", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutSwapPrivateKey indicates an expected call of PutSwapPrivateKey.
func (mr *MockRecoveryDBMockRecorder) PutSwapPrivateKey(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutSwapPrivateKey", reflect.TypeOf((*MockRecoveryDB)(nil).PutSwapPrivateKey), arg0, arg1)
}

// PutSwapRelayerInfo mocks base method.
func (m *MockRecoveryDB) PutSwapRelayerInfo(arg0 common.Hash, arg1 *types.OfferExtra) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutSwapRelayerInfo", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutSwapRelayerInfo indicates an expected call of PutSwapRelayerInfo.
func (mr *MockRecoveryDBMockRecorder) PutSwapRelayerInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutSwapRelayerInfo", reflect.TypeOf((*MockRecoveryDB)(nil).PutSwapRelayerInfo), arg0, arg1)
}

// PutXMRMakerSwapKeys mocks base method.
func (m *MockRecoveryDB) PutXMRMakerSwapKeys(arg0 common.Hash, arg1 *mcrypto.PublicKey, arg2 *mcrypto.PrivateViewKey) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutXMRMakerSwapKeys", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutXMRMakerSwapKeys indicates an expected call of PutXMRMakerSwapKeys.
func (mr *MockRecoveryDBMockRecorder) PutXMRMakerSwapKeys(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutXMRMakerSwapKeys", reflect.TypeOf((*MockRecoveryDB)(nil).PutXMRMakerSwapKeys), arg0, arg1, arg2)
}

// PutXMRTakerSwapKeys mocks base method.
func (m *MockRecoveryDB) PutXMRTakerSwapKeys(arg0 common.Hash, arg1 *mcrypto.PublicKeyPair) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutXMRTakerSwapKeys", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutXMRTakerSwapKeys indicates an expected call of PutXMRTakerSwapKeys.
func (mr *MockRecoveryDBMockRecorder) PutXMRTakerSwapKeys(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutXMRTakerSwapKeys", reflect.TypeOf((*MockRecoveryDB)(nil).PutXMRTakerSwapKeys), arg0, arg1)
}
