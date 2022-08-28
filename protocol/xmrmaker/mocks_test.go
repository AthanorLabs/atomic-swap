// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/athanorlabs/atomic-swap/protocol/backend (interfaces: Backend)

// Package xmrmaker is a generated GoMock package.
package xmrmaker

import (
	context "context"
	big "math/big"
	reflect "reflect"
	time "time"

	wallet "github.com/MarinX/monerorpc/wallet"
	common "github.com/athanorlabs/atomic-swap/common"
	types "github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	net "github.com/athanorlabs/atomic-swap/net"
	message "github.com/athanorlabs/atomic-swap/net/message"
	swap "github.com/athanorlabs/atomic-swap/protocol/swap"
	txsender "github.com/athanorlabs/atomic-swap/protocol/txsender"
	swapfactory "github.com/athanorlabs/atomic-swap/swapfactory"
	ethereum "github.com/ethereum/go-ethereum"
	bind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	common0 "github.com/ethereum/go-ethereum/common"
	types0 "github.com/ethereum/go-ethereum/core/types"
	gomock "github.com/golang/mock/gomock"
)

// MockBackend is a mock of Backend interface.
type MockBackend struct {
	ctrl     *gomock.Controller
	recorder *MockBackendMockRecorder
}

// MockBackendMockRecorder is the mock recorder for MockBackend.
type MockBackendMockRecorder struct {
	mock *MockBackend
}

// NewMockBackend creates a new mock instance.
func NewMockBackend(ctrl *gomock.Controller) *MockBackend {
	mock := &MockBackend{ctrl: ctrl}
	mock.recorder = &MockBackendMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBackend) EXPECT() *MockBackendMockRecorder {
	return m.recorder
}

// BalanceAt mocks base method.
func (m *MockBackend) BalanceAt(arg0 context.Context, arg1 common0.Address, arg2 *big.Int) (*big.Int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BalanceAt", arg0, arg1, arg2)
	ret0, _ := ret[0].(*big.Int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BalanceAt indicates an expected call of BalanceAt.
func (mr *MockBackendMockRecorder) BalanceAt(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BalanceAt", reflect.TypeOf((*MockBackend)(nil).BalanceAt), arg0, arg1, arg2)
}

// CallOpts mocks base method.
func (m *MockBackend) CallOpts() *bind.CallOpts {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CallOpts")
	ret0, _ := ret[0].(*bind.CallOpts)
	return ret0
}

// CallOpts indicates an expected call of CallOpts.
func (mr *MockBackendMockRecorder) CallOpts() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CallOpts", reflect.TypeOf((*MockBackend)(nil).CallOpts))
}

// ChainID mocks base method.
func (m *MockBackend) ChainID() *big.Int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChainID")
	ret0, _ := ret[0].(*big.Int)
	return ret0
}

// ChainID indicates an expected call of ChainID.
func (mr *MockBackendMockRecorder) ChainID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChainID", reflect.TypeOf((*MockBackend)(nil).ChainID))
}

// Claim mocks base method.
func (m *MockBackend) Claim(arg0 types.Hash, arg1 swapfactory.SwapFactorySwap, arg2 [32]byte) (common0.Hash, *types0.Receipt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Claim", arg0, arg1, arg2)
	ret0, _ := ret[0].(common0.Hash)
	ret1, _ := ret[1].(*types0.Receipt)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Claim indicates an expected call of Claim.
func (mr *MockBackendMockRecorder) Claim(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Claim", reflect.TypeOf((*MockBackend)(nil).Claim), arg0, arg1, arg2)
}

// ClearXMRDepositAddress mocks base method.
func (m *MockBackend) ClearXMRDepositAddress(arg0 types.Hash) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ClearXMRDepositAddress", arg0)
}

// ClearXMRDepositAddress indicates an expected call of ClearXMRDepositAddress.
func (mr *MockBackendMockRecorder) ClearXMRDepositAddress(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClearXMRDepositAddress", reflect.TypeOf((*MockBackend)(nil).ClearXMRDepositAddress), arg0)
}

// CloseWallet mocks base method.
func (m *MockBackend) CloseWallet() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseWallet")
	ret0, _ := ret[0].(error)
	return ret0
}

// CloseWallet indicates an expected call of CloseWallet.
func (mr *MockBackendMockRecorder) CloseWallet() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseWallet", reflect.TypeOf((*MockBackend)(nil).CloseWallet))
}

// CodeAt mocks base method.
func (m *MockBackend) CodeAt(arg0 context.Context, arg1 common0.Address, arg2 *big.Int) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CodeAt", arg0, arg1, arg2)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CodeAt indicates an expected call of CodeAt.
func (mr *MockBackendMockRecorder) CodeAt(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CodeAt", reflect.TypeOf((*MockBackend)(nil).CodeAt), arg0, arg1, arg2)
}

// Contract mocks base method.
func (m *MockBackend) Contract() *swapfactory.SwapFactory {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Contract")
	ret0, _ := ret[0].(*swapfactory.SwapFactory)
	return ret0
}

// Contract indicates an expected call of Contract.
func (mr *MockBackendMockRecorder) Contract() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Contract", reflect.TypeOf((*MockBackend)(nil).Contract))
}

// ContractAddr mocks base method.
func (m *MockBackend) ContractAddr() common0.Address {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContractAddr")
	ret0, _ := ret[0].(common0.Address)
	return ret0
}

// ContractAddr indicates an expected call of ContractAddr.
func (mr *MockBackendMockRecorder) ContractAddr() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContractAddr", reflect.TypeOf((*MockBackend)(nil).ContractAddr))
}

// CreateWallet mocks base method.
func (m *MockBackend) CreateWallet(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateWallet", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateWallet indicates an expected call of CreateWallet.
func (mr *MockBackendMockRecorder) CreateWallet(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateWallet", reflect.TypeOf((*MockBackend)(nil).CreateWallet), arg0, arg1)
}

// Ctx mocks base method.
func (m *MockBackend) Ctx() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ctx")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Ctx indicates an expected call of Ctx.
func (mr *MockBackendMockRecorder) Ctx() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ctx", reflect.TypeOf((*MockBackend)(nil).Ctx))
}

// ERC20BalanceAt mocks base method.
func (m *MockBackend) ERC20BalanceAt(arg0 context.Context, arg1, arg2 common0.Address, arg3 *big.Int) (*big.Int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ERC20BalanceAt", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*big.Int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ERC20BalanceAt indicates an expected call of ERC20BalanceAt.
func (mr *MockBackendMockRecorder) ERC20BalanceAt(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ERC20BalanceAt", reflect.TypeOf((*MockBackend)(nil).ERC20BalanceAt), arg0, arg1, arg2, arg3)
}

// ERC20Info mocks base method.
func (m *MockBackend) ERC20Info(arg0 context.Context, arg1 common0.Address) (string, string, byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ERC20Info", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(byte)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ERC20Info indicates an expected call of ERC20Info.
func (mr *MockBackendMockRecorder) ERC20Info(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ERC20Info", reflect.TypeOf((*MockBackend)(nil).ERC20Info), arg0, arg1)
}

// Env mocks base method.
func (m *MockBackend) Env() common.Environment {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Env")
	ret0, _ := ret[0].(common.Environment)
	return ret0
}

// Env indicates an expected call of Env.
func (mr *MockBackendMockRecorder) Env() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Env", reflect.TypeOf((*MockBackend)(nil).Env))
}

// EthAddress mocks base method.
func (m *MockBackend) EthAddress() common0.Address {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EthAddress")
	ret0, _ := ret[0].(common0.Address)
	return ret0
}

// EthAddress indicates an expected call of EthAddress.
func (mr *MockBackendMockRecorder) EthAddress() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EthAddress", reflect.TypeOf((*MockBackend)(nil).EthAddress))
}

// ExternalSender mocks base method.
func (m *MockBackend) ExternalSender() *txsender.ExternalSender {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExternalSender")
	ret0, _ := ret[0].(*txsender.ExternalSender)
	return ret0
}

// ExternalSender indicates an expected call of ExternalSender.
func (mr *MockBackendMockRecorder) ExternalSender() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExternalSender", reflect.TypeOf((*MockBackend)(nil).ExternalSender))
}

// FilterLogs mocks base method.
func (m *MockBackend) FilterLogs(arg0 context.Context, arg1 ethereum.FilterQuery) ([]types0.Log, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FilterLogs", arg0, arg1)
	ret0, _ := ret[0].([]types0.Log)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FilterLogs indicates an expected call of FilterLogs.
func (mr *MockBackendMockRecorder) FilterLogs(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FilterLogs", reflect.TypeOf((*MockBackend)(nil).FilterLogs), arg0, arg1)
}

// GenerateBlocks mocks base method.
func (m *MockBackend) GenerateBlocks(arg0 string, arg1 uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateBlocks", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// GenerateBlocks indicates an expected call of GenerateBlocks.
func (mr *MockBackendMockRecorder) GenerateBlocks(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateBlocks", reflect.TypeOf((*MockBackend)(nil).GenerateBlocks), arg0, arg1)
}

// GenerateFromKeys mocks base method.
func (m *MockBackend) GenerateFromKeys(arg0 *mcrypto.PrivateKeyPair, arg1, arg2 string, arg3 common.Environment) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateFromKeys", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// GenerateFromKeys indicates an expected call of GenerateFromKeys.
func (mr *MockBackendMockRecorder) GenerateFromKeys(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateFromKeys", reflect.TypeOf((*MockBackend)(nil).GenerateFromKeys), arg0, arg1, arg2, arg3)
}

// GenerateViewOnlyWalletFromKeys mocks base method.
func (m *MockBackend) GenerateViewOnlyWalletFromKeys(arg0 *mcrypto.PrivateViewKey, arg1 mcrypto.Address, arg2, arg3 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateViewOnlyWalletFromKeys", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// GenerateViewOnlyWalletFromKeys indicates an expected call of GenerateViewOnlyWalletFromKeys.
func (mr *MockBackendMockRecorder) GenerateViewOnlyWalletFromKeys(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateViewOnlyWalletFromKeys", reflect.TypeOf((*MockBackend)(nil).GenerateViewOnlyWalletFromKeys), arg0, arg1, arg2, arg3)
}

// GetAccounts mocks base method.
func (m *MockBackend) GetAccounts() (*wallet.GetAccountsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccounts")
	ret0, _ := ret[0].(*wallet.GetAccountsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccounts indicates an expected call of GetAccounts.
func (mr *MockBackendMockRecorder) GetAccounts() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccounts", reflect.TypeOf((*MockBackend)(nil).GetAccounts))
}

// GetAddress mocks base method.
func (m *MockBackend) GetAddress(arg0 uint64) (*wallet.GetAddressResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAddress", arg0)
	ret0, _ := ret[0].(*wallet.GetAddressResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAddress indicates an expected call of GetAddress.
func (mr *MockBackendMockRecorder) GetAddress(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAddress", reflect.TypeOf((*MockBackend)(nil).GetAddress), arg0)
}

// GetBalance mocks base method.
func (m *MockBackend) GetBalance(arg0 uint64) (*wallet.GetBalanceResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBalance", arg0)
	ret0, _ := ret[0].(*wallet.GetBalanceResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalance indicates an expected call of GetBalance.
func (mr *MockBackendMockRecorder) GetBalance(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalance", reflect.TypeOf((*MockBackend)(nil).GetBalance), arg0)
}

// GetHeight mocks base method.
func (m *MockBackend) GetHeight() (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHeight")
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHeight indicates an expected call of GetHeight.
func (mr *MockBackendMockRecorder) GetHeight() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeight", reflect.TypeOf((*MockBackend)(nil).GetHeight))
}

// LatestBlockTimestamp mocks base method.
func (m *MockBackend) LatestBlockTimestamp(arg0 context.Context) (time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LatestBlockTimestamp", arg0)
	ret0, _ := ret[0].(time.Time)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LatestBlockTimestamp indicates an expected call of LatestBlockTimestamp.
func (mr *MockBackendMockRecorder) LatestBlockTimestamp(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LatestBlockTimestamp", reflect.TypeOf((*MockBackend)(nil).LatestBlockTimestamp), arg0)
}

// LockClient mocks base method.
func (m *MockBackend) LockClient() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "LockClient")
}

// LockClient indicates an expected call of LockClient.
func (mr *MockBackendMockRecorder) LockClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LockClient", reflect.TypeOf((*MockBackend)(nil).LockClient))
}

// Net mocks base method.
func (m *MockBackend) Net() net.MessageSender {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Net")
	ret0, _ := ret[0].(net.MessageSender)
	return ret0
}

// Net indicates an expected call of Net.
func (mr *MockBackendMockRecorder) Net() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Net", reflect.TypeOf((*MockBackend)(nil).Net))
}

// NewSwap mocks base method.
func (m *MockBackend) NewSwap(arg0 types.Hash, arg1, arg2 [32]byte, arg3 common0.Address, arg4, arg5 *big.Int, arg6 types.EthAsset, arg7 *big.Int) (common0.Hash, *types0.Receipt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewSwap", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	ret0, _ := ret[0].(common0.Hash)
	ret1, _ := ret[1].(*types0.Receipt)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// NewSwap indicates an expected call of NewSwap.
func (mr *MockBackendMockRecorder) NewSwap(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewSwap", reflect.TypeOf((*MockBackend)(nil).NewSwap), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}

// NewSwapFactory mocks base method.
func (m *MockBackend) NewSwapFactory(arg0 common0.Address) (*swapfactory.SwapFactory, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewSwapFactory", arg0)
	ret0, _ := ret[0].(*swapfactory.SwapFactory)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewSwapFactory indicates an expected call of NewSwapFactory.
func (mr *MockBackendMockRecorder) NewSwapFactory(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewSwapFactory", reflect.TypeOf((*MockBackend)(nil).NewSwapFactory), arg0)
}

// OpenWallet mocks base method.
func (m *MockBackend) OpenWallet(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OpenWallet", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// OpenWallet indicates an expected call of OpenWallet.
func (mr *MockBackendMockRecorder) OpenWallet(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenWallet", reflect.TypeOf((*MockBackend)(nil).OpenWallet), arg0, arg1)
}

// Refresh mocks base method.
func (m *MockBackend) Refresh() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Refresh")
	ret0, _ := ret[0].(error)
	return ret0
}

// Refresh indicates an expected call of Refresh.
func (mr *MockBackendMockRecorder) Refresh() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Refresh", reflect.TypeOf((*MockBackend)(nil).Refresh))
}

// Refund mocks base method.
func (m *MockBackend) Refund(arg0 types.Hash, arg1 swapfactory.SwapFactorySwap, arg2 [32]byte) (common0.Hash, *types0.Receipt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Refund", arg0, arg1, arg2)
	ret0, _ := ret[0].(common0.Hash)
	ret1, _ := ret[1].(*types0.Receipt)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Refund indicates an expected call of Refund.
func (mr *MockBackendMockRecorder) Refund(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Refund", reflect.TypeOf((*MockBackend)(nil).Refund), arg0, arg1, arg2)
}

// SendSwapMessage mocks base method.
func (m *MockBackend) SendSwapMessage(arg0 message.Message, arg1 types.Hash) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendSwapMessage", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendSwapMessage indicates an expected call of SendSwapMessage.
func (mr *MockBackendMockRecorder) SendSwapMessage(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendSwapMessage", reflect.TypeOf((*MockBackend)(nil).SendSwapMessage), arg0, arg1)
}

// SetBaseXMRDepositAddress mocks base method.
func (m *MockBackend) SetBaseXMRDepositAddress(arg0 mcrypto.Address) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetBaseXMRDepositAddress", arg0)
}

// SetBaseXMRDepositAddress indicates an expected call of SetBaseXMRDepositAddress.
func (mr *MockBackendMockRecorder) SetBaseXMRDepositAddress(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetBaseXMRDepositAddress", reflect.TypeOf((*MockBackend)(nil).SetBaseXMRDepositAddress), arg0)
}

// SetContract mocks base method.
func (m *MockBackend) SetContract(arg0 *swapfactory.SwapFactory) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetContract", arg0)
}

// SetContract indicates an expected call of SetContract.
func (mr *MockBackendMockRecorder) SetContract(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetContract", reflect.TypeOf((*MockBackend)(nil).SetContract), arg0)
}

// SetContractAddress mocks base method.
func (m *MockBackend) SetContractAddress(arg0 common0.Address) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetContractAddress", arg0)
}

// SetContractAddress indicates an expected call of SetContractAddress.
func (mr *MockBackendMockRecorder) SetContractAddress(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetContractAddress", reflect.TypeOf((*MockBackend)(nil).SetContractAddress), arg0)
}

// SetEthAddress mocks base method.
func (m *MockBackend) SetEthAddress(arg0 common0.Address) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetEthAddress", arg0)
}

// SetEthAddress indicates an expected call of SetEthAddress.
func (mr *MockBackendMockRecorder) SetEthAddress(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetEthAddress", reflect.TypeOf((*MockBackend)(nil).SetEthAddress), arg0)
}

// SetGasPrice mocks base method.
func (m *MockBackend) SetGasPrice(arg0 uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetGasPrice", arg0)
}

// SetGasPrice indicates an expected call of SetGasPrice.
func (mr *MockBackendMockRecorder) SetGasPrice(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetGasPrice", reflect.TypeOf((*MockBackend)(nil).SetGasPrice), arg0)
}

// SetReady mocks base method.
func (m *MockBackend) SetReady(arg0 types.Hash, arg1 swapfactory.SwapFactorySwap) (common0.Hash, *types0.Receipt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetReady", arg0, arg1)
	ret0, _ := ret[0].(common0.Hash)
	ret1, _ := ret[1].(*types0.Receipt)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// SetReady indicates an expected call of SetReady.
func (mr *MockBackendMockRecorder) SetReady(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetReady", reflect.TypeOf((*MockBackend)(nil).SetReady), arg0, arg1)
}

// SetSwapTimeout mocks base method.
func (m *MockBackend) SetSwapTimeout(arg0 time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetSwapTimeout", arg0)
}

// SetSwapTimeout indicates an expected call of SetSwapTimeout.
func (mr *MockBackendMockRecorder) SetSwapTimeout(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSwapTimeout", reflect.TypeOf((*MockBackend)(nil).SetSwapTimeout), arg0)
}

// SetXMRDepositAddress mocks base method.
func (m *MockBackend) SetXMRDepositAddress(arg0 mcrypto.Address, arg1 types.Hash) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetXMRDepositAddress", arg0, arg1)
}

// SetXMRDepositAddress indicates an expected call of SetXMRDepositAddress.
func (mr *MockBackendMockRecorder) SetXMRDepositAddress(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetXMRDepositAddress", reflect.TypeOf((*MockBackend)(nil).SetXMRDepositAddress), arg0, arg1)
}

// SwapManager mocks base method.
func (m *MockBackend) SwapManager() swap.Manager {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SwapManager")
	ret0, _ := ret[0].(swap.Manager)
	return ret0
}

// SwapManager indicates an expected call of SwapManager.
func (mr *MockBackendMockRecorder) SwapManager() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SwapManager", reflect.TypeOf((*MockBackend)(nil).SwapManager))
}

// SwapTimeout mocks base method.
func (m *MockBackend) SwapTimeout() time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SwapTimeout")
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// SwapTimeout indicates an expected call of SwapTimeout.
func (mr *MockBackendMockRecorder) SwapTimeout() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SwapTimeout", reflect.TypeOf((*MockBackend)(nil).SwapTimeout))
}

// SweepAll mocks base method.
func (m *MockBackend) SweepAll(arg0 mcrypto.Address, arg1 uint64) (*wallet.SweepAllResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SweepAll", arg0, arg1)
	ret0, _ := ret[0].(*wallet.SweepAllResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SweepAll indicates an expected call of SweepAll.
func (mr *MockBackendMockRecorder) SweepAll(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SweepAll", reflect.TypeOf((*MockBackend)(nil).SweepAll), arg0, arg1)
}

// TransactionReceipt mocks base method.
func (m *MockBackend) TransactionReceipt(arg0 context.Context, arg1 common0.Hash) (*types0.Receipt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TransactionReceipt", arg0, arg1)
	ret0, _ := ret[0].(*types0.Receipt)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TransactionReceipt indicates an expected call of TransactionReceipt.
func (mr *MockBackendMockRecorder) TransactionReceipt(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TransactionReceipt", reflect.TypeOf((*MockBackend)(nil).TransactionReceipt), arg0, arg1)
}

// Transfer mocks base method.
func (m *MockBackend) Transfer(arg0 mcrypto.Address, arg1, arg2 uint64) (*wallet.TransferResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Transfer", arg0, arg1, arg2)
	ret0, _ := ret[0].(*wallet.TransferResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Transfer indicates an expected call of Transfer.
func (mr *MockBackendMockRecorder) Transfer(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Transfer", reflect.TypeOf((*MockBackend)(nil).Transfer), arg0, arg1, arg2)
}

// TxOpts mocks base method.
func (m *MockBackend) TxOpts() (*bind.TransactOpts, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TxOpts")
	ret0, _ := ret[0].(*bind.TransactOpts)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TxOpts indicates an expected call of TxOpts.
func (mr *MockBackendMockRecorder) TxOpts() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TxOpts", reflect.TypeOf((*MockBackend)(nil).TxOpts))
}

// UnlockClient mocks base method.
func (m *MockBackend) UnlockClient() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UnlockClient")
}

// UnlockClient indicates an expected call of UnlockClient.
func (mr *MockBackendMockRecorder) UnlockClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnlockClient", reflect.TypeOf((*MockBackend)(nil).UnlockClient))
}

// WaitForReceipt mocks base method.
func (m *MockBackend) WaitForReceipt(arg0 context.Context, arg1 common0.Hash) (*types0.Receipt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitForReceipt", arg0, arg1)
	ret0, _ := ret[0].(*types0.Receipt)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WaitForReceipt indicates an expected call of WaitForReceipt.
func (mr *MockBackendMockRecorder) WaitForReceipt(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitForReceipt", reflect.TypeOf((*MockBackend)(nil).WaitForReceipt), arg0, arg1)
}

// WaitForTimestamp mocks base method.
func (m *MockBackend) WaitForTimestamp(arg0 context.Context, arg1 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitForTimestamp", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitForTimestamp indicates an expected call of WaitForTimestamp.
func (mr *MockBackendMockRecorder) WaitForTimestamp(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitForTimestamp", reflect.TypeOf((*MockBackend)(nil).WaitForTimestamp), arg0, arg1)
}

// XMRDepositAddress mocks base method.
func (m *MockBackend) XMRDepositAddress(arg0 *types.Hash) (mcrypto.Address, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "XMRDepositAddress", arg0)
	ret0, _ := ret[0].(mcrypto.Address)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// XMRDepositAddress indicates an expected call of XMRDepositAddress.
func (mr *MockBackendMockRecorder) XMRDepositAddress(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "XMRDepositAddress", reflect.TypeOf((*MockBackend)(nil).XMRDepositAddress), arg0)
}
