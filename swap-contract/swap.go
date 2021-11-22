// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package swap

import (
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
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// SwapABI is the input ABI used to generate the binding from.
const SwapABI = "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_claimHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_refundHash\",\"type\":\"bytes32\"},{\"internalType\":\"addresspayable\",\"name\":\"_claimer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_timeoutDuration\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"claimHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"refundHash\",\"type\":\"bytes32\"}],\"name\":\"Constructed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"b\",\"type\":\"bool\"}],\"name\":\"IsReady\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"refundHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"set_ready\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeout_0\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeout_1\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// SwapBin is the compiled bytecode used for deploying new contracts.
var SwapBin = "0x61014060405260008060006101000a81548160ff02191690831515021790555060405162000ea138038062000ea183398181016040528101906200004491906200021f565b3373ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff16815250508360c081815250508260e081815250508173ffffffffffffffffffffffffffffffffffffffff1660a08173ffffffffffffffffffffffffffffffffffffffff16815250508042620000ca9190620002c0565b6101008181525050600281620000e191906200031d565b42620000ee9190620002c0565b61012081815250507f8d36aa70807342c3036697a846281194626fd4afa892356ad5979e03831ab08060c05160e0516040516200012d9291906200038f565b60405180910390a150505050620003bc565b600080fd5b6000819050919050565b620001598162000144565b81146200016557600080fd5b50565b60008151905062000179816200014e565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000620001ac826200017f565b9050919050565b620001be816200019f565b8114620001ca57600080fd5b50565b600081519050620001de81620001b3565b92915050565b6000819050919050565b620001f981620001e4565b81146200020557600080fd5b50565b6000815190506200021981620001ee565b92915050565b600080600080608085870312156200023c576200023b6200013f565b5b60006200024c8782880162000168565b94505060206200025f8782880162000168565b93505060406200027287828801620001cd565b9250506060620002858782880162000208565b91505092959194509250565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000620002cd82620001e4565b9150620002da83620001e4565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0382111562000312576200031162000291565b5b828201905092915050565b60006200032a82620001e4565b91506200033783620001e4565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff048311821515161562000373576200037262000291565b5b828202905092915050565b620003898162000144565b82525050565b6000604082019050620003a660008301856200037e565b620003b560208301846200037e565b9392505050565b60805160a05160c05160e0516101005161012051610a5d620004446000396000818161013e0152818161020201526105130152600081816101620152818161022c015261053d01526000818161018601526102a601526000818161046101526105b6015260006104850152600081816101aa0152818161036501526103b40152610a5d6000f3fe608060405234801561001057600080fd5b506004361061007d5760003560e01c80637249fbb61161005b5780637249fbb6146100dc57806374d7c138146100f857806386e4efc514610102578063bd66528a146101205761007d565b806345bb8e09146100825780634ded8d52146100a057806368fe850e146100be575b600080fd5b61008a61013c565b60405161009791906106a5565b60405180910390f35b6100a8610160565b6040516100b591906106a5565b60405180910390f35b6100c6610184565b6040516100d391906106d9565b60405180910390f35b6100f660048036038101906100f19190610725565b6101a8565b005b61010061039c565b005b61010a61045f565b60405161011791906106d9565b60405180910390f35b61013a60048036038101906101359190610725565b610483565b005b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461020057600080fd5b7f00000000000000000000000000000000000000000000000000000000000000004210158061026557507f000000000000000000000000000000000000000000000000000000000000000042108015610264575060008054906101000a900460ff16155b5b6102a4576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161029b906107d5565b60405180910390fd5b7f0000000000000000000000000000000000000000000000000000000000000000816040516020016102d691906106d9565b604051602081830303815290604052805190602001201461032c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161032390610867565b60405180910390fd5b7ffe509803c09416b28ff3d8f690c8b0c61462a892c46d5430c8fb20abe472daf08160405161035b91906106d9565b60405180910390a17f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16ff5b60008054906101000a900460ff1615801561040257507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b61040b57600080fd5b60016000806101000a81548160ff0219169083151502179055507f2724cf6c3ad6a3399ad72482e4013d0171794f3ef4c462b7e24790c658cb3cd4600160405161045591906108a2565b60405180910390a1565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610511576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161050890610909565b60405180910390fd5b7f00000000000000000000000000000000000000000000000000000000000000004210801561057557507f000000000000000000000000000000000000000000000000000000000000000042101580610574575060008054906101000a900460ff165b5b6105b4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105ab90610975565b60405180910390fd5b7f0000000000000000000000000000000000000000000000000000000000000000816040516020016105e691906106d9565b604051602081830303815290604052805190602001201461063c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161063390610a07565b60405180910390fd5b7feddf608ef698454af2fb41c1df7b7e5154ff0d46969f895e0f39c7dfe7e6380a8160405161066b91906106d9565b60405180910390a13373ffffffffffffffffffffffffffffffffffffffff16ff5b6000819050919050565b61069f8161068c565b82525050565b60006020820190506106ba6000830184610696565b92915050565b6000819050919050565b6106d3816106c0565b82525050565b60006020820190506106ee60008301846106ca565b92915050565b600080fd5b610702816106c0565b811461070d57600080fd5b50565b60008135905061071f816106f9565b92915050565b60006020828403121561073b5761073a6106f4565b5b600061074984828501610710565b91505092915050565b600082825260208201905092915050565b7f4974277320426f622773207475726e206e6f772c20706c65617365207761697460008201527f2100000000000000000000000000000000000000000000000000000000000000602082015250565b60006107bf602183610752565b91506107ca82610763565b604082019050919050565b600060208201905081810360008301526107ee816107b2565b9050919050565b7f736563726574206973206e6f7420707265696d61676520746f20726566756e6460008201527f4861736800000000000000000000000000000000000000000000000000000000602082015250565b6000610851602483610752565b915061085c826107f5565b604082019050919050565b6000602082019050818103600083015261088081610844565b9050919050565b60008115159050919050565b61089c81610887565b82525050565b60006020820190506108b76000830184610893565b92915050565b7f6f6e6c7920636c61696d65722063616e20636c61696d21000000000000000000600082015250565b60006108f3601783610752565b91506108fe826108bd565b602082019050919050565b60006020820190508181036000830152610922816108e6565b9050919050565b7f746f6f206c617465206f72206561726c7920746f20636c61696d210000000000600082015250565b600061095f601b83610752565b915061096a82610929565b602082019050919050565b6000602082019050818103600083015261098e81610952565b9050919050565b7f736563726574206973206e6f7420707265696d61676520746f20636c61696d4860008201527f6173680000000000000000000000000000000000000000000000000000000000602082015250565b60006109f1602383610752565b91506109fc82610995565b604082019050919050565b60006020820190508181036000830152610a20816109e4565b905091905056fea2646970667358221220dccfad7c09ca699bb63ab4df2c4fe5067a448a4bdf7032ac6928d1faa8dfe1bc64736f6c634300080a0033"

// DeploySwap deploys a new Ethereum contract, binding an instance of Swap to it.
func DeploySwap(auth *bind.TransactOpts, backend bind.ContractBackend, _claimHash [32]byte, _refundHash [32]byte, _claimer common.Address, _timeoutDuration *big.Int) (common.Address, *types.Transaction, *Swap, error) {
	parsed, err := abi.JSON(strings.NewReader(SwapABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SwapBin), backend, _claimHash, _refundHash, _claimer, _timeoutDuration)
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

// ClaimHash is a free data retrieval call binding the contract method 0x86e4efc5.
//
// Solidity: function claimHash() view returns(bytes32)
func (_Swap *SwapCaller) ClaimHash(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "claimHash")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ClaimHash is a free data retrieval call binding the contract method 0x86e4efc5.
//
// Solidity: function claimHash() view returns(bytes32)
func (_Swap *SwapSession) ClaimHash() ([32]byte, error) {
	return _Swap.Contract.ClaimHash(&_Swap.CallOpts)
}

// ClaimHash is a free data retrieval call binding the contract method 0x86e4efc5.
//
// Solidity: function claimHash() view returns(bytes32)
func (_Swap *SwapCallerSession) ClaimHash() ([32]byte, error) {
	return _Swap.Contract.ClaimHash(&_Swap.CallOpts)
}

// RefundHash is a free data retrieval call binding the contract method 0x68fe850e.
//
// Solidity: function refundHash() view returns(bytes32)
func (_Swap *SwapCaller) RefundHash(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "refundHash")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RefundHash is a free data retrieval call binding the contract method 0x68fe850e.
//
// Solidity: function refundHash() view returns(bytes32)
func (_Swap *SwapSession) RefundHash() ([32]byte, error) {
	return _Swap.Contract.RefundHash(&_Swap.CallOpts)
}

// RefundHash is a free data retrieval call binding the contract method 0x68fe850e.
//
// Solidity: function refundHash() view returns(bytes32)
func (_Swap *SwapCallerSession) RefundHash() ([32]byte, error) {
	return _Swap.Contract.RefundHash(&_Swap.CallOpts)
}

// Timeout0 is a free data retrieval call binding the contract method 0x4ded8d52.
//
// Solidity: function timeout_0() view returns(uint256)
func (_Swap *SwapCaller) Timeout0(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "timeout_0")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Timeout0 is a free data retrieval call binding the contract method 0x4ded8d52.
//
// Solidity: function timeout_0() view returns(uint256)
func (_Swap *SwapSession) Timeout0() (*big.Int, error) {
	return _Swap.Contract.Timeout0(&_Swap.CallOpts)
}

// Timeout0 is a free data retrieval call binding the contract method 0x4ded8d52.
//
// Solidity: function timeout_0() view returns(uint256)
func (_Swap *SwapCallerSession) Timeout0() (*big.Int, error) {
	return _Swap.Contract.Timeout0(&_Swap.CallOpts)
}

// Timeout1 is a free data retrieval call binding the contract method 0x45bb8e09.
//
// Solidity: function timeout_1() view returns(uint256)
func (_Swap *SwapCaller) Timeout1(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "timeout_1")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Timeout1 is a free data retrieval call binding the contract method 0x45bb8e09.
//
// Solidity: function timeout_1() view returns(uint256)
func (_Swap *SwapSession) Timeout1() (*big.Int, error) {
	return _Swap.Contract.Timeout1(&_Swap.CallOpts)
}

// Timeout1 is a free data retrieval call binding the contract method 0x45bb8e09.
//
// Solidity: function timeout_1() view returns(uint256)
func (_Swap *SwapCallerSession) Timeout1() (*big.Int, error) {
	return _Swap.Contract.Timeout1(&_Swap.CallOpts)
}

// Claim is a paid mutator transaction binding the contract method 0xbd66528a.
//
// Solidity: function claim(bytes32 _s) returns()
func (_Swap *SwapTransactor) Claim(opts *bind.TransactOpts, _s [32]byte) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "claim", _s)
}

// Claim is a paid mutator transaction binding the contract method 0xbd66528a.
//
// Solidity: function claim(bytes32 _s) returns()
func (_Swap *SwapSession) Claim(_s [32]byte) (*types.Transaction, error) {
	return _Swap.Contract.Claim(&_Swap.TransactOpts, _s)
}

// Claim is a paid mutator transaction binding the contract method 0xbd66528a.
//
// Solidity: function claim(bytes32 _s) returns()
func (_Swap *SwapTransactorSession) Claim(_s [32]byte) (*types.Transaction, error) {
	return _Swap.Contract.Claim(&_Swap.TransactOpts, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x7249fbb6.
//
// Solidity: function refund(bytes32 _s) returns()
func (_Swap *SwapTransactor) Refund(opts *bind.TransactOpts, _s [32]byte) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "refund", _s)
}

// Refund is a paid mutator transaction binding the contract method 0x7249fbb6.
//
// Solidity: function refund(bytes32 _s) returns()
func (_Swap *SwapSession) Refund(_s [32]byte) (*types.Transaction, error) {
	return _Swap.Contract.Refund(&_Swap.TransactOpts, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x7249fbb6.
//
// Solidity: function refund(bytes32 _s) returns()
func (_Swap *SwapTransactorSession) Refund(_s [32]byte) (*types.Transaction, error) {
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

// SwapClaimedIterator is returned from FilterClaimed and is used to iterate over the raw logs and unpacked data for Claimed events raised by the Swap contract.
type SwapClaimedIterator struct {
	Event *SwapClaimed // Event containing the contract specifics and raw log

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
func (it *SwapClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapClaimed)
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
		it.Event = new(SwapClaimed)
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
func (it *SwapClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapClaimed represents a Claimed event raised by the Swap contract.
type SwapClaimed struct {
	S   [32]byte
	Raw types.Log // Blockchain specific contextual infos
}

// FilterClaimed is a free log retrieval operation binding the contract event 0xeddf608ef698454af2fb41c1df7b7e5154ff0d46969f895e0f39c7dfe7e6380a.
//
// Solidity: event Claimed(bytes32 s)
func (_Swap *SwapFilterer) FilterClaimed(opts *bind.FilterOpts) (*SwapClaimedIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "Claimed")
	if err != nil {
		return nil, err
	}
	return &SwapClaimedIterator{contract: _Swap.contract, event: "Claimed", logs: logs, sub: sub}, nil
}

// WatchClaimed is a free log subscription operation binding the contract event 0xeddf608ef698454af2fb41c1df7b7e5154ff0d46969f895e0f39c7dfe7e6380a.
//
// Solidity: event Claimed(bytes32 s)
func (_Swap *SwapFilterer) WatchClaimed(opts *bind.WatchOpts, sink chan<- *SwapClaimed) (event.Subscription, error) {

	logs, sub, err := _Swap.contract.WatchLogs(opts, "Claimed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapClaimed)
				if err := _Swap.contract.UnpackLog(event, "Claimed", log); err != nil {
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

// ParseClaimed is a log parse operation binding the contract event 0xeddf608ef698454af2fb41c1df7b7e5154ff0d46969f895e0f39c7dfe7e6380a.
//
// Solidity: event Claimed(bytes32 s)
func (_Swap *SwapFilterer) ParseClaimed(log types.Log) (*SwapClaimed, error) {
	event := new(SwapClaimed)
	if err := _Swap.contract.UnpackLog(event, "Claimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapConstructedIterator is returned from FilterConstructed and is used to iterate over the raw logs and unpacked data for Constructed events raised by the Swap contract.
type SwapConstructedIterator struct {
	Event *SwapConstructed // Event containing the contract specifics and raw log

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
func (it *SwapConstructedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapConstructed)
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
		it.Event = new(SwapConstructed)
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
func (it *SwapConstructedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapConstructedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapConstructed represents a Constructed event raised by the Swap contract.
type SwapConstructed struct {
	ClaimHash  [32]byte
	RefundHash [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterConstructed is a free log retrieval operation binding the contract event 0x8d36aa70807342c3036697a846281194626fd4afa892356ad5979e03831ab080.
//
// Solidity: event Constructed(bytes32 claimHash, bytes32 refundHash)
func (_Swap *SwapFilterer) FilterConstructed(opts *bind.FilterOpts) (*SwapConstructedIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "Constructed")
	if err != nil {
		return nil, err
	}
	return &SwapConstructedIterator{contract: _Swap.contract, event: "Constructed", logs: logs, sub: sub}, nil
}

// WatchConstructed is a free log subscription operation binding the contract event 0x8d36aa70807342c3036697a846281194626fd4afa892356ad5979e03831ab080.
//
// Solidity: event Constructed(bytes32 claimHash, bytes32 refundHash)
func (_Swap *SwapFilterer) WatchConstructed(opts *bind.WatchOpts, sink chan<- *SwapConstructed) (event.Subscription, error) {

	logs, sub, err := _Swap.contract.WatchLogs(opts, "Constructed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapConstructed)
				if err := _Swap.contract.UnpackLog(event, "Constructed", log); err != nil {
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

// ParseConstructed is a log parse operation binding the contract event 0x8d36aa70807342c3036697a846281194626fd4afa892356ad5979e03831ab080.
//
// Solidity: event Constructed(bytes32 claimHash, bytes32 refundHash)
func (_Swap *SwapFilterer) ParseConstructed(log types.Log) (*SwapConstructed, error) {
	event := new(SwapConstructed)
	if err := _Swap.contract.UnpackLog(event, "Constructed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapIsReadyIterator is returned from FilterIsReady and is used to iterate over the raw logs and unpacked data for IsReady events raised by the Swap contract.
type SwapIsReadyIterator struct {
	Event *SwapIsReady // Event containing the contract specifics and raw log

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
func (it *SwapIsReadyIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapIsReady)
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
		it.Event = new(SwapIsReady)
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
func (it *SwapIsReadyIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapIsReadyIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapIsReady represents a IsReady event raised by the Swap contract.
type SwapIsReady struct {
	B   bool
	Raw types.Log // Blockchain specific contextual infos
}

// FilterIsReady is a free log retrieval operation binding the contract event 0x2724cf6c3ad6a3399ad72482e4013d0171794f3ef4c462b7e24790c658cb3cd4.
//
// Solidity: event IsReady(bool b)
func (_Swap *SwapFilterer) FilterIsReady(opts *bind.FilterOpts) (*SwapIsReadyIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "IsReady")
	if err != nil {
		return nil, err
	}
	return &SwapIsReadyIterator{contract: _Swap.contract, event: "IsReady", logs: logs, sub: sub}, nil
}

// WatchIsReady is a free log subscription operation binding the contract event 0x2724cf6c3ad6a3399ad72482e4013d0171794f3ef4c462b7e24790c658cb3cd4.
//
// Solidity: event IsReady(bool b)
func (_Swap *SwapFilterer) WatchIsReady(opts *bind.WatchOpts, sink chan<- *SwapIsReady) (event.Subscription, error) {

	logs, sub, err := _Swap.contract.WatchLogs(opts, "IsReady")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapIsReady)
				if err := _Swap.contract.UnpackLog(event, "IsReady", log); err != nil {
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

// ParseIsReady is a log parse operation binding the contract event 0x2724cf6c3ad6a3399ad72482e4013d0171794f3ef4c462b7e24790c658cb3cd4.
//
// Solidity: event IsReady(bool b)
func (_Swap *SwapFilterer) ParseIsReady(log types.Log) (*SwapIsReady, error) {
	event := new(SwapIsReady)
	if err := _Swap.contract.UnpackLog(event, "IsReady", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapRefundedIterator is returned from FilterRefunded and is used to iterate over the raw logs and unpacked data for Refunded events raised by the Swap contract.
type SwapRefundedIterator struct {
	Event *SwapRefunded // Event containing the contract specifics and raw log

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
func (it *SwapRefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapRefunded)
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
		it.Event = new(SwapRefunded)
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
func (it *SwapRefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapRefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapRefunded represents a Refunded event raised by the Swap contract.
type SwapRefunded struct {
	S   [32]byte
	Raw types.Log // Blockchain specific contextual infos
}

// FilterRefunded is a free log retrieval operation binding the contract event 0xfe509803c09416b28ff3d8f690c8b0c61462a892c46d5430c8fb20abe472daf0.
//
// Solidity: event Refunded(bytes32 s)
func (_Swap *SwapFilterer) FilterRefunded(opts *bind.FilterOpts) (*SwapRefundedIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return &SwapRefundedIterator{contract: _Swap.contract, event: "Refunded", logs: logs, sub: sub}, nil
}

// WatchRefunded is a free log subscription operation binding the contract event 0xfe509803c09416b28ff3d8f690c8b0c61462a892c46d5430c8fb20abe472daf0.
//
// Solidity: event Refunded(bytes32 s)
func (_Swap *SwapFilterer) WatchRefunded(opts *bind.WatchOpts, sink chan<- *SwapRefunded) (event.Subscription, error) {

	logs, sub, err := _Swap.contract.WatchLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapRefunded)
				if err := _Swap.contract.UnpackLog(event, "Refunded", log); err != nil {
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

// ParseRefunded is a log parse operation binding the contract event 0xfe509803c09416b28ff3d8f690c8b0c61462a892c46d5430c8fb20abe472daf0.
//
// Solidity: event Refunded(bytes32 s)
func (_Swap *SwapFilterer) ParseRefunded(log types.Log) (*SwapRefunded, error) {
	event := new(SwapRefunded)
	if err := _Swap.contract.UnpackLog(event, "Refunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
