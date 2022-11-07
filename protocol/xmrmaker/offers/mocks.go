// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/athanorlabs/atomic-swap/protocol/xmrmaker/offers (interfaces: Database)

// Package offers is a generated GoMock package.
package offers

import (
	reflect "reflect"

	types "github.com/athanorlabs/atomic-swap/common/types"
	common "github.com/ethereum/go-ethereum/common"
	gomock "github.com/golang/mock/gomock"
)

// MockDatabase is a mock of Database interface.
type MockDatabase struct {
	ctrl     *gomock.Controller
	recorder *MockDatabaseMockRecorder
}

// MockDatabaseMockRecorder is the mock recorder for MockDatabase.
type MockDatabaseMockRecorder struct {
	mock *MockDatabase
}

// NewMockDatabase creates a new mock instance.
func NewMockDatabase(ctrl *gomock.Controller) *MockDatabase {
	mock := &MockDatabase{ctrl: ctrl}
	mock.recorder = &MockDatabaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDatabase) EXPECT() *MockDatabaseMockRecorder {
	return m.recorder
}

// ClearAllOffers mocks base method.
func (m *MockDatabase) ClearAllOffers() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ClearAllOffers")
	ret0, _ := ret[0].(error)
	return ret0
}

// ClearAllOffers indicates an expected call of ClearAllOffers.
func (mr *MockDatabaseMockRecorder) ClearAllOffers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClearAllOffers", reflect.TypeOf((*MockDatabase)(nil).ClearAllOffers))
}

// DeleteOffer mocks base method.
func (m *MockDatabase) DeleteOffer(arg0 common.Hash) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteOffer", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteOffer indicates an expected call of DeleteOffer.
func (mr *MockDatabaseMockRecorder) DeleteOffer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteOffer", reflect.TypeOf((*MockDatabase)(nil).DeleteOffer), arg0)
}

// GetAllOffers mocks base method.
func (m *MockDatabase) GetAllOffers() ([]*types.Offer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllOffers")
	ret0, _ := ret[0].([]*types.Offer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllOffers indicates an expected call of GetAllOffers.
func (mr *MockDatabaseMockRecorder) GetAllOffers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllOffers", reflect.TypeOf((*MockDatabase)(nil).GetAllOffers))
}

// PutOffer mocks base method.
func (m *MockDatabase) PutOffer(arg0 *types.Offer) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutOffer", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutOffer indicates an expected call of PutOffer.
func (mr *MockDatabaseMockRecorder) PutOffer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutOffer", reflect.TypeOf((*MockDatabase)(nil).PutOffer), arg0)
}
