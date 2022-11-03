// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/athanorlabs/atomic-swap/net (interfaces: Host)

// Package xmrmaker is a generated GoMock package.
package xmrmaker

import (
	reflect "reflect"
	time "time"

	common "github.com/athanorlabs/atomic-swap/common"
	types "github.com/athanorlabs/atomic-swap/common/types"
	message "github.com/athanorlabs/atomic-swap/net/message"
	gomock "github.com/golang/mock/gomock"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

// MockHost is a mock of Host interface.
type MockHost struct {
	ctrl     *gomock.Controller
	recorder *MockHostMockRecorder
}

// MockHostMockRecorder is the mock recorder for MockHost.
type MockHostMockRecorder struct {
	mock *MockHost
}

// NewMockHost creates a new mock instance.
func NewMockHost(ctrl *gomock.Controller) *MockHost {
	mock := &MockHost{ctrl: ctrl}
	mock.recorder = &MockHostMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHost) EXPECT() *MockHostMockRecorder {
	return m.recorder
}

// Advertise mocks base method.
func (m *MockHost) Advertise() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Advertise")
}

// Advertise indicates an expected call of Advertise.
func (mr *MockHostMockRecorder) Advertise() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Advertise", reflect.TypeOf((*MockHost)(nil).Advertise))
}

// CloseProtocolStream mocks base method.
func (m *MockHost) CloseProtocolStream(arg0 types.Hash) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CloseProtocolStream", arg0)
}

// CloseProtocolStream indicates an expected call of CloseProtocolStream.
func (mr *MockHostMockRecorder) CloseProtocolStream(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseProtocolStream", reflect.TypeOf((*MockHost)(nil).CloseProtocolStream), arg0)
}

// Discover mocks base method.
func (m *MockHost) Discover(arg0 types.ProvidesCoin, arg1 time.Duration) ([]peer.AddrInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Discover", arg0, arg1)
	ret0, _ := ret[0].([]peer.AddrInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Discover indicates an expected call of Discover.
func (mr *MockHostMockRecorder) Discover(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Discover", reflect.TypeOf((*MockHost)(nil).Discover), arg0, arg1)
}

// Initiate mocks base method.
func (m *MockHost) Initiate(arg0 peer.AddrInfo, arg1 *message.SendKeysMessage, arg2 common.SwapStateNet) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Initiate", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Initiate indicates an expected call of Initiate.
func (mr *MockHostMockRecorder) Initiate(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Initiate", reflect.TypeOf((*MockHost)(nil).Initiate), arg0, arg1, arg2)
}

// Query mocks base method.
func (m *MockHost) Query(arg0 peer.AddrInfo) (*message.QueryResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", arg0)
	ret0, _ := ret[0].(*message.QueryResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query.
func (mr *MockHostMockRecorder) Query(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockHost)(nil).Query), arg0)
}

// SendSwapMessage mocks base method.
func (m *MockHost) SendSwapMessage(arg0 message.Message, arg1 types.Hash) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendSwapMessage", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendSwapMessage indicates an expected call of SendSwapMessage.
func (mr *MockHostMockRecorder) SendSwapMessage(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendSwapMessage", reflect.TypeOf((*MockHost)(nil).SendSwapMessage), arg0, arg1)
}

// Start mocks base method.
func (m *MockHost) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockHostMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockHost)(nil).Start))
}

// Stop mocks base method.
func (m *MockHost) Stop() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockHostMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockHost)(nil).Stop))
}
