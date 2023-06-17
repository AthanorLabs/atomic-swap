// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package block

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// UTContractMetaData contains all meta data concerning the UTContract contract.
var UTContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_stamp\",\"type\":\"uint256\"}],\"name\":\"checkStamp\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b506101cb8061001d5f395ff3fe608060405234801561000f575f80fd5b5060043610610029575f3560e01c8063d5f7a03a1461002d575b5f80fd5b610047600480360381019061004291906100cc565b610049565b005b8042111561008c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161008390610177565b60405180910390fd5b805f8190555050565b5f80fd5b5f819050919050565b6100ab81610099565b81146100b5575f80fd5b50565b5f813590506100c6816100a2565b92915050565b5f602082840312156100e1576100e0610095565b5b5f6100ee848285016100b8565b91505092915050565b5f82825260208201905092915050565b7f626c6f636b2e74696d657374616d7020776173206e6f74206c657373207468615f8201527f6e207374616d7000000000000000000000000000000000000000000000000000602082015250565b5f6101616027836100f7565b915061016c82610107565b604082019050919050565b5f6020820190508181035f83015261018e81610155565b905091905056fea264697066735822122077057003501b23f47c692e417ddc5f1429b9309ebe23d189b6906de5aa0d260964736f6c63430008140033",
}

// UTContractABI is the input ABI used to generate the binding from.
// Deprecated: Use UTContractMetaData.ABI instead.
var UTContractABI = UTContractMetaData.ABI

// UTContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use UTContractMetaData.Bin instead.
var UTContractBin = UTContractMetaData.Bin

// DeployUTContract deploys a new Ethereum contract, binding an instance of UTContract to it.
func DeployUTContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *UTContract, error) {
	parsed, err := UTContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(UTContractBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &UTContract{UTContractCaller: UTContractCaller{contract: contract}, UTContractTransactor: UTContractTransactor{contract: contract}, UTContractFilterer: UTContractFilterer{contract: contract}}, nil
}

// UTContract is an auto generated Go binding around an Ethereum contract.
type UTContract struct {
	UTContractCaller     // Read-only binding to the contract
	UTContractTransactor // Write-only binding to the contract
	UTContractFilterer   // Log filterer for contract events
}

// UTContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type UTContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UTContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type UTContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UTContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type UTContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UTContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type UTContractSession struct {
	Contract     *UTContract       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// UTContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type UTContractCallerSession struct {
	Contract *UTContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// UTContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type UTContractTransactorSession struct {
	Contract     *UTContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// UTContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type UTContractRaw struct {
	Contract *UTContract // Generic contract binding to access the raw methods on
}

// UTContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type UTContractCallerRaw struct {
	Contract *UTContractCaller // Generic read-only contract binding to access the raw methods on
}

// UTContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type UTContractTransactorRaw struct {
	Contract *UTContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewUTContract creates a new instance of UTContract, bound to a specific deployed contract.
func NewUTContract(address common.Address, backend bind.ContractBackend) (*UTContract, error) {
	contract, err := bindUTContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &UTContract{UTContractCaller: UTContractCaller{contract: contract}, UTContractTransactor: UTContractTransactor{contract: contract}, UTContractFilterer: UTContractFilterer{contract: contract}}, nil
}

// NewUTContractCaller creates a new read-only instance of UTContract, bound to a specific deployed contract.
func NewUTContractCaller(address common.Address, caller bind.ContractCaller) (*UTContractCaller, error) {
	contract, err := bindUTContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &UTContractCaller{contract: contract}, nil
}

// NewUTContractTransactor creates a new write-only instance of UTContract, bound to a specific deployed contract.
func NewUTContractTransactor(address common.Address, transactor bind.ContractTransactor) (*UTContractTransactor, error) {
	contract, err := bindUTContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &UTContractTransactor{contract: contract}, nil
}

// NewUTContractFilterer creates a new log filterer instance of UTContract, bound to a specific deployed contract.
func NewUTContractFilterer(address common.Address, filterer bind.ContractFilterer) (*UTContractFilterer, error) {
	contract, err := bindUTContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &UTContractFilterer{contract: contract}, nil
}

// bindUTContract binds a generic wrapper to an already deployed contract.
func bindUTContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := UTContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UTContract *UTContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _UTContract.Contract.UTContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UTContract *UTContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UTContract.Contract.UTContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UTContract *UTContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UTContract.Contract.UTContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UTContract *UTContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _UTContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UTContract *UTContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UTContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UTContract *UTContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UTContract.Contract.contract.Transact(opts, method, params...)
}

// CheckStamp is a paid mutator transaction binding the contract method 0xd5f7a03a.
//
// Solidity: function checkStamp(uint256 _stamp) returns()
func (_UTContract *UTContractTransactor) CheckStamp(opts *bind.TransactOpts, _stamp *big.Int) (*types.Transaction, error) {
	return _UTContract.contract.Transact(opts, "checkStamp", _stamp)
}

// CheckStamp is a paid mutator transaction binding the contract method 0xd5f7a03a.
//
// Solidity: function checkStamp(uint256 _stamp) returns()
func (_UTContract *UTContractSession) CheckStamp(_stamp *big.Int) (*types.Transaction, error) {
	return _UTContract.Contract.CheckStamp(&_UTContract.TransactOpts, _stamp)
}

// CheckStamp is a paid mutator transaction binding the contract method 0xd5f7a03a.
//
// Solidity: function checkStamp(uint256 _stamp) returns()
func (_UTContract *UTContractTransactorSession) CheckStamp(_stamp *big.Int) (*types.Transaction, error) {
	return _UTContract.Contract.CheckStamp(&_UTContract.TransactOpts, _stamp)
}
