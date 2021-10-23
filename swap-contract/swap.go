// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package swap

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
)

// SwapMetaData contains all meta data concerning the Swap contract.
var SwapMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_pubKeyRefund\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"s\",\"type\":\"uint256\"}],\"name\":\"DerivedPubKeyClaim\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"s\",\"type\":\"uint256\"}],\"name\":\"DerivedPubKeyRefund\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_s\",\"type\":\"uint256\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pubKeyClaim\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pubKeyRefund\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_s\",\"type\":\"uint256\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"set_ready\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040526000600560146101000a81548160ff02191690831515021790555060405162000d0738038062000d0783398181016040528101906200004491906200016b565b33600560006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555081600181905550806002819055506201518042620000a49190620001eb565b600381905550604051620000b8906200011d565b604051809103906000f080158015620000d5573d6000803e3d6000fd5b506000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550505062000248565b61014c8062000bbb83390190565b600080fd5b6000819050919050565b620001458162000130565b81146200015157600080fd5b50565b60008151905062000165816200013a565b92915050565b600080604083850312156200018557620001846200012b565b5b6000620001958582860162000154565b9250506020620001a88582860162000154565b9150509250929050565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000620001f882620001b2565b91506200020583620001b2565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156200023d576200023c620001bc565b5b828201905092915050565b61096380620002586000396000f3fe608060405234801561001057600080fd5b50600436106100575760003560e01c806303f7e2461461005c578063278ecde11461007a578063379607f514610096578063736290f8146100b257806374d7c138146100d0575b600080fd5b6100646100da565b60405161007191906105dc565b60405180910390f35b610094600480360381019061008f9190610632565b6100e0565b005b6100b060048036038101906100ab9190610632565b6102e1565b005b6100ba610531565b6040516100c791906105dc565b60405180910390f35b6100d8610537565b005b60025481565b6003544211158015610105575060001515600560149054906101000a900460ff161515145b8061011257506004544211155b61011b57600080fd5b60008060008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663c4f4912b846040518263ffffffff1660e01b8152600401610177919061066e565b604080518083038186803b15801561018e57600080fd5b505afa1580156101a2573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906101c6919061069e565b91509150600082826040516020016101df9291906106de565b604051602081830303815290604052805190602001209050600254811461023b576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102329061078a565b60405180910390fd5b7f349c9cedc1d596c3b1aa537408b5cd2e966f0ceb5ad4c4a6ff5943e392ddd9df8460405161026a919061066e565b60405180910390a1600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc479081150290604051600060405180830381858888f193505050501580156102da573d6000803e3d6000fd5b5050505050565b60011515600560149054906101000a900460ff161515141561034757600454421115610342576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610339906107f6565b60405180910390fd5b61038d565b60035442101561038c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161038390610888565b60405180910390fd5b5b60008060008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663c4f4912b846040518263ffffffff1660e01b81526004016103e9919061066e565b604080518083038186803b15801561040057600080fd5b505afa158015610414573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610438919061069e565b91509150600082826040516020016104519291906106de565b60405160208183030381529060405280519060200120905060015481146104ad576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104a49061078a565b60405180910390fd5b7f05e2253b8f6851b3d1e3e53c602b41bbcdf31b10621d844c02774c107791d653846040516104dc919061066e565b60405180910390a13373ffffffffffffffffffffffffffffffffffffffff166108fc479081150290604051600060405180830381858888f1935050505015801561052a573d6000803e3d6000fd5b5050505050565b60015481565b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461059157600080fd5b6001600560146101000a81548160ff02191690831515021790555062015180426105bb91906108d7565b600481905550565b6000819050919050565b6105d6816105c3565b82525050565b60006020820190506105f160008301846105cd565b92915050565b600080fd5b6000819050919050565b61060f816105fc565b811461061a57600080fd5b50565b60008135905061062c81610606565b92915050565b600060208284031215610648576106476105f7565b5b60006106568482850161061d565b91505092915050565b610668816105fc565b82525050565b6000602082019050610683600083018461065f565b92915050565b60008151905061069881610606565b92915050565b600080604083850312156106b5576106b46105f7565b5b60006106c385828601610689565b92505060206106d485828601610689565b9150509250929050565b60006040820190506106f3600083018561065f565b610700602083018461065f565b9392505050565b600082825260208201905092915050565b7f70726f76696465642073656372657420646f6573206e6f74206d61746368207460008201527f6865206578706563746564207075624b65790000000000000000000000000000602082015250565b6000610774603283610707565b915061077f82610718565b604082019050919050565b600060208201905081810360008301526107a381610767565b9050919050565b7f546f6f206c61746520746f20636c61696d210000000000000000000000000000600082015250565b60006107e0601283610707565b91506107eb826107aa565b602082019050919050565b6000602082019050818103600083015261080f816107d3565b9050919050565b7f2769735265616479203d3d2066616c7365272063616e6e6f7420636c61696d2060008201527f7965742100000000000000000000000000000000000000000000000000000000602082015250565b6000610872602483610707565b915061087d82610816565b604082019050919050565b600060208201905081810360008301526108a181610865565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60006108e2826105fc565b91506108ed836105fc565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03821115610922576109216108a8565b5b82820190509291505056fea264697066735822122048c011d89f4616423021397c3db6f57c3dace96bcb326490ac60531f34487ec264736f6c63430008090033608060405234801561001057600080fd5b5061012c806100206000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063c4f4912b14602d575b600080fd5b60436004803603810190603f9190609c565b6058565b604051604f92919060d1565b60405180910390f35b600080828391509150915091565b600080fd5b6000819050919050565b607c81606b565b8114608657600080fd5b50565b6000813590506096816075565b92915050565b60006020828403121560af5760ae6066565b5b600060bb848285016089565b91505092915050565b60cb81606b565b82525050565b600060408201905060e4600083018560c4565b60ef602083018460c4565b939250505056fea26469706673582212206933b5d025a67f52fa646193883f1d23a7f4c480751cbe152b28c15358968beb64736f6c63430008090033",
}

// SwapABI is the input ABI used to generate the binding from.
// Deprecated: Use SwapMetaData.ABI instead.
var SwapABI = SwapMetaData.ABI

// SwapBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SwapMetaData.Bin instead.
var SwapBin = SwapMetaData.Bin

// DeploySwap deploys a new Ethereum contract, binding an instance of Swap to it.
func DeploySwap(auth *bind.TransactOpts, backend bind.ContractBackend, _pubKeyClaim [32]byte, _pubKeyRefund [32]byte) (common.Address, *types.Transaction, *Swap, error) {
	parsed, err := SwapMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SwapBin), backend, _pubKeyClaim, _pubKeyRefund)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Swap{SwapCaller: SwapCaller{contract: contract}, SwapTransactor: SwapTransactor{contract: contract}, SwapFilterer: SwapFilterer{contract: contract}}, nil
}

// Swap is an auto generated Go binding around an Ethereum contract.
type Swap struct {
	SwapCaller     // Read-only binding to the contract
	SwapTransactor // Write-only binding to the contract
	SwapFilterer   // Log filterer for contract events
}

// SwapCaller is an auto generated read-only Go binding around an Ethereum contract.
type SwapCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SwapTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SwapFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SwapSession struct {
	Contract     *Swap             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SwapCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SwapCallerSession struct {
	Contract *SwapCaller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// SwapTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SwapTransactorSession struct {
	Contract     *SwapTransactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SwapRaw is an auto generated low-level Go binding around an Ethereum contract.
type SwapRaw struct {
	Contract *Swap // Generic contract binding to access the raw methods on
}

// SwapCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SwapCallerRaw struct {
	Contract *SwapCaller // Generic read-only contract binding to access the raw methods on
}

// SwapTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SwapTransactorRaw struct {
	Contract *SwapTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSwap creates a new instance of Swap, bound to a specific deployed contract.
func NewSwap(address common.Address, backend bind.ContractBackend) (*Swap, error) {
	contract, err := bindSwap(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Swap{SwapCaller: SwapCaller{contract: contract}, SwapTransactor: SwapTransactor{contract: contract}, SwapFilterer: SwapFilterer{contract: contract}}, nil
}

// NewSwapCaller creates a new read-only instance of Swap, bound to a specific deployed contract.
func NewSwapCaller(address common.Address, caller bind.ContractCaller) (*SwapCaller, error) {
	contract, err := bindSwap(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SwapCaller{contract: contract}, nil
}

// NewSwapTransactor creates a new write-only instance of Swap, bound to a specific deployed contract.
func NewSwapTransactor(address common.Address, transactor bind.ContractTransactor) (*SwapTransactor, error) {
	contract, err := bindSwap(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SwapTransactor{contract: contract}, nil
}

// NewSwapFilterer creates a new log filterer instance of Swap, bound to a specific deployed contract.
func NewSwapFilterer(address common.Address, filterer bind.ContractFilterer) (*SwapFilterer, error) {
	contract, err := bindSwap(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SwapFilterer{contract: contract}, nil
}

// bindSwap binds a generic wrapper to an already deployed contract.
func bindSwap(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SwapABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Swap *SwapRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Swap.Contract.SwapCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Swap *SwapRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Swap.Contract.SwapTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Swap *SwapRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Swap.Contract.SwapTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Swap *SwapCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Swap.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Swap *SwapTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Swap.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Swap *SwapTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Swap.Contract.contract.Transact(opts, method, params...)
}

// PubKeyClaim is a free data retrieval call binding the contract method 0x736290f8.
//
// Solidity: function pubKeyClaim() view returns(bytes32)
func (_Swap *SwapCaller) PubKeyClaim(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "pubKeyClaim")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PubKeyClaim is a free data retrieval call binding the contract method 0x736290f8.
//
// Solidity: function pubKeyClaim() view returns(bytes32)
func (_Swap *SwapSession) PubKeyClaim() ([32]byte, error) {
	return _Swap.Contract.PubKeyClaim(&_Swap.CallOpts)
}

// PubKeyClaim is a free data retrieval call binding the contract method 0x736290f8.
//
// Solidity: function pubKeyClaim() view returns(bytes32)
func (_Swap *SwapCallerSession) PubKeyClaim() ([32]byte, error) {
	return _Swap.Contract.PubKeyClaim(&_Swap.CallOpts)
}

// PubKeyRefund is a free data retrieval call binding the contract method 0x03f7e246.
//
// Solidity: function pubKeyRefund() view returns(bytes32)
func (_Swap *SwapCaller) PubKeyRefund(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "pubKeyRefund")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PubKeyRefund is a free data retrieval call binding the contract method 0x03f7e246.
//
// Solidity: function pubKeyRefund() view returns(bytes32)
func (_Swap *SwapSession) PubKeyRefund() ([32]byte, error) {
	return _Swap.Contract.PubKeyRefund(&_Swap.CallOpts)
}

// PubKeyRefund is a free data retrieval call binding the contract method 0x03f7e246.
//
// Solidity: function pubKeyRefund() view returns(bytes32)
func (_Swap *SwapCallerSession) PubKeyRefund() ([32]byte, error) {
	return _Swap.Contract.PubKeyRefund(&_Swap.CallOpts)
}

// Claim is a paid mutator transaction binding the contract method 0x379607f5.
//
// Solidity: function claim(uint256 _s) returns()
func (_Swap *SwapTransactor) Claim(opts *bind.TransactOpts, _s *big.Int) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "claim", _s)
}

// Claim is a paid mutator transaction binding the contract method 0x379607f5.
//
// Solidity: function claim(uint256 _s) returns()
func (_Swap *SwapSession) Claim(_s *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.Claim(&_Swap.TransactOpts, _s)
}

// Claim is a paid mutator transaction binding the contract method 0x379607f5.
//
// Solidity: function claim(uint256 _s) returns()
func (_Swap *SwapTransactorSession) Claim(_s *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.Claim(&_Swap.TransactOpts, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 _s) returns()
func (_Swap *SwapTransactor) Refund(opts *bind.TransactOpts, _s *big.Int) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "refund", _s)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 _s) returns()
func (_Swap *SwapSession) Refund(_s *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.Refund(&_Swap.TransactOpts, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 _s) returns()
func (_Swap *SwapTransactorSession) Refund(_s *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.Refund(&_Swap.TransactOpts, _s)
}

// SetReady is a paid mutator transaction binding the contract method 0x74d7c138.
//
// Solidity: function set_ready() returns()
func (_Swap *SwapTransactor) SetReady(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "set_ready")
}

// SetReady is a paid mutator transaction binding the contract method 0x74d7c138.
//
// Solidity: function set_ready() returns()
func (_Swap *SwapSession) SetReady() (*types.Transaction, error) {
	return _Swap.Contract.SetReady(&_Swap.TransactOpts)
}

// SetReady is a paid mutator transaction binding the contract method 0x74d7c138.
//
// Solidity: function set_ready() returns()
func (_Swap *SwapTransactorSession) SetReady() (*types.Transaction, error) {
	return _Swap.Contract.SetReady(&_Swap.TransactOpts)
}

// SwapDerivedPubKeyClaimIterator is returned from FilterDerivedPubKeyClaim and is used to iterate over the raw logs and unpacked data for DerivedPubKeyClaim events raised by the Swap contract.
type SwapDerivedPubKeyClaimIterator struct {
	Event *SwapDerivedPubKeyClaim // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SwapDerivedPubKeyClaimIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapDerivedPubKeyClaim)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SwapDerivedPubKeyClaim)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SwapDerivedPubKeyClaimIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapDerivedPubKeyClaimIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapDerivedPubKeyClaim represents a DerivedPubKeyClaim event raised by the Swap contract.
type SwapDerivedPubKeyClaim struct {
	S   *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterDerivedPubKeyClaim is a free log retrieval operation binding the contract event 0x05e2253b8f6851b3d1e3e53c602b41bbcdf31b10621d844c02774c107791d653.
//
// Solidity: event DerivedPubKeyClaim(uint256 s)
func (_Swap *SwapFilterer) FilterDerivedPubKeyClaim(opts *bind.FilterOpts) (*SwapDerivedPubKeyClaimIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "DerivedPubKeyClaim")
	if err != nil {
		return nil, err
	}
	return &SwapDerivedPubKeyClaimIterator{contract: _Swap.contract, event: "DerivedPubKeyClaim", logs: logs, sub: sub}, nil
}

// WatchDerivedPubKeyClaim is a free log subscription operation binding the contract event 0x05e2253b8f6851b3d1e3e53c602b41bbcdf31b10621d844c02774c107791d653.
//
// Solidity: event DerivedPubKeyClaim(uint256 s)
func (_Swap *SwapFilterer) WatchDerivedPubKeyClaim(opts *bind.WatchOpts, sink chan<- *SwapDerivedPubKeyClaim) (event.Subscription, error) {

	logs, sub, err := _Swap.contract.WatchLogs(opts, "DerivedPubKeyClaim")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapDerivedPubKeyClaim)
				if err := _Swap.contract.UnpackLog(event, "DerivedPubKeyClaim", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDerivedPubKeyClaim is a log parse operation binding the contract event 0x05e2253b8f6851b3d1e3e53c602b41bbcdf31b10621d844c02774c107791d653.
//
// Solidity: event DerivedPubKeyClaim(uint256 s)
func (_Swap *SwapFilterer) ParseDerivedPubKeyClaim(log types.Log) (*SwapDerivedPubKeyClaim, error) {
	event := new(SwapDerivedPubKeyClaim)
	if err := _Swap.contract.UnpackLog(event, "DerivedPubKeyClaim", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapDerivedPubKeyRefundIterator is returned from FilterDerivedPubKeyRefund and is used to iterate over the raw logs and unpacked data for DerivedPubKeyRefund events raised by the Swap contract.
type SwapDerivedPubKeyRefundIterator struct {
	Event *SwapDerivedPubKeyRefund // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SwapDerivedPubKeyRefundIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapDerivedPubKeyRefund)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SwapDerivedPubKeyRefund)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SwapDerivedPubKeyRefundIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapDerivedPubKeyRefundIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapDerivedPubKeyRefund represents a DerivedPubKeyRefund event raised by the Swap contract.
type SwapDerivedPubKeyRefund struct {
	S   *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterDerivedPubKeyRefund is a free log retrieval operation binding the contract event 0x349c9cedc1d596c3b1aa537408b5cd2e966f0ceb5ad4c4a6ff5943e392ddd9df.
//
// Solidity: event DerivedPubKeyRefund(uint256 s)
func (_Swap *SwapFilterer) FilterDerivedPubKeyRefund(opts *bind.FilterOpts) (*SwapDerivedPubKeyRefundIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "DerivedPubKeyRefund")
	if err != nil {
		return nil, err
	}
	return &SwapDerivedPubKeyRefundIterator{contract: _Swap.contract, event: "DerivedPubKeyRefund", logs: logs, sub: sub}, nil
}

// WatchDerivedPubKeyRefund is a free log subscription operation binding the contract event 0x349c9cedc1d596c3b1aa537408b5cd2e966f0ceb5ad4c4a6ff5943e392ddd9df.
//
// Solidity: event DerivedPubKeyRefund(uint256 s)
func (_Swap *SwapFilterer) WatchDerivedPubKeyRefund(opts *bind.WatchOpts, sink chan<- *SwapDerivedPubKeyRefund) (event.Subscription, error) {

	logs, sub, err := _Swap.contract.WatchLogs(opts, "DerivedPubKeyRefund")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapDerivedPubKeyRefund)
				if err := _Swap.contract.UnpackLog(event, "DerivedPubKeyRefund", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDerivedPubKeyRefund is a log parse operation binding the contract event 0x349c9cedc1d596c3b1aa537408b5cd2e966f0ceb5ad4c4a6ff5943e392ddd9df.
//
// Solidity: event DerivedPubKeyRefund(uint256 s)
func (_Swap *SwapFilterer) ParseDerivedPubKeyRefund(log types.Log) (*SwapDerivedPubKeyRefund, error) {
	event := new(SwapDerivedPubKeyRefund)
	if err := _Swap.contract.UnpackLog(event, "DerivedPubKeyRefund", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
