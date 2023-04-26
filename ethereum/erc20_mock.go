// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

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

// TestERC20MetaData contains all meta data concerning the TestERC20 contract.
var TestERC20MetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"uint8\",\"name\":\"numDecimals\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"initialAccount\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"initialBalance\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approveInternal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferInternal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405260405162000f5238038062000f528339810160408190526200002691620001fe565b848460036200003683826200033c565b5060046200004582826200033c565b50506005805460ff191660ff8616179055506200006382826200006e565b505050505062000430565b6001600160a01b038216620000c95760405162461bcd60e51b815260206004820152601f60248201527f45524332303a206d696e7420746f20746865207a65726f206164647265737300604482015260640160405180910390fd5b8060026000828254620000dd919062000408565b90915550506001600160a01b038216600081815260208181526040808320805486019055518481527fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef910160405180910390a35050565b505050565b634e487b7160e01b600052604160045260246000fd5b600082601f8301126200016157600080fd5b81516001600160401b03808211156200017e576200017e62000139565b604051601f8301601f19908116603f01168101908282118183101715620001a957620001a962000139565b81604052838152602092508683858801011115620001c657600080fd5b600091505b83821015620001ea5785820183015181830184015290820190620001cb565b600093810190920192909252949350505050565b600080600080600060a086880312156200021757600080fd5b85516001600160401b03808211156200022f57600080fd5b6200023d89838a016200014f565b965060208801519150808211156200025457600080fd5b5062000263888289016200014f565b945050604086015160ff811681146200027b57600080fd5b60608701519093506001600160a01b03811681146200029957600080fd5b80925050608086015190509295509295909350565b600181811c90821680620002c357607f821691505b602082108103620002e457634e487b7160e01b600052602260045260246000fd5b50919050565b601f8211156200013457600081815260208120601f850160051c81016020861015620003135750805b601f850160051c820191505b8181101562000334578281556001016200031f565b505050505050565b81516001600160401b0381111562000358576200035862000139565b6200037081620003698454620002ae565b84620002ea565b602080601f831160018114620003a857600084156200038f5750858301515b600019600386901b1c1916600185901b17855562000334565b600085815260208120601f198616915b82811015620003d957888601518255948401946001909101908401620003b8565b5085821015620003f85787850151600019600388901b60f8161c191681555b5050505050600190811b01905550565b808201808211156200042a57634e487b7160e01b600052601160045260246000fd5b92915050565b610b1280620004406000396000f3fe608060405234801561001057600080fd5b50600436106100f55760003560e01c806340c10f19116100975780639dc29fac116100665780639dc29fac146101f4578063a457c2d714610207578063a9059cbb1461021a578063dd62ed3e1461022d57600080fd5b806340c10f191461019d57806356189cb4146101b057806370a08231146101c357806395d89b41146101ec57600080fd5b8063222f5be0116100d3578063222f5be01461014d57806323b872dd14610162578063313ce56714610175578063395093511461018a57600080fd5b806306fdde03146100fa578063095ea7b31461011857806318160ddd1461013b575b600080fd5b610102610240565b60405161010f919061095c565b60405180910390f35b61012b6101263660046109c6565b6102d2565b604051901515815260200161010f565b6002545b60405190815260200161010f565b61016061015b3660046109f0565b6102ec565b005b61012b6101703660046109f0565b6102fc565b60055460405160ff909116815260200161010f565b61012b6101983660046109c6565b610320565b6101606101ab3660046109c6565b610342565b6101606101be3660046109f0565b610350565b61013f6101d1366004610a2c565b6001600160a01b031660009081526020819052604090205490565b61010261035b565b6101606102023660046109c6565b61036a565b61012b6102153660046109c6565b610374565b61012b6102283660046109c6565b6103f4565b61013f61023b366004610a4e565b610402565b60606003805461024f90610a81565b80601f016020809104026020016040519081016040528092919081815260200182805461027b90610a81565b80156102c85780601f1061029d576101008083540402835291602001916102c8565b820191906000526020600020905b8154815290600101906020018083116102ab57829003601f168201915b5050505050905090565b6000336102e081858561042d565b60019150505b92915050565b6102f7838383610551565b505050565b60003361030a8582856106f7565b610315858585610551565b506001949350505050565b6000336102e08185856103338383610402565b61033d9190610abb565b61042d565b61034c828261076b565b5050565b6102f783838361042d565b60606004805461024f90610a81565b61034c828261082a565b600033816103828286610402565b9050838110156103e75760405162461bcd60e51b815260206004820152602560248201527f45524332303a2064656372656173656420616c6c6f77616e63652062656c6f77604482015264207a65726f60d81b60648201526084015b60405180910390fd5b610315828686840361042d565b6000336102e0818585610551565b6001600160a01b03918216600090815260016020908152604080832093909416825291909152205490565b6001600160a01b03831661048f5760405162461bcd60e51b8152602060048201526024808201527f45524332303a20617070726f76652066726f6d20746865207a65726f206164646044820152637265737360e01b60648201526084016103de565b6001600160a01b0382166104f05760405162461bcd60e51b815260206004820152602260248201527f45524332303a20617070726f766520746f20746865207a65726f206164647265604482015261737360f01b60648201526084016103de565b6001600160a01b0383811660008181526001602090815260408083209487168084529482529182902085905590518481527f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925910160405180910390a3505050565b6001600160a01b0383166105b55760405162461bcd60e51b815260206004820152602560248201527f45524332303a207472616e736665722066726f6d20746865207a65726f206164604482015264647265737360d81b60648201526084016103de565b6001600160a01b0382166106175760405162461bcd60e51b815260206004820152602360248201527f45524332303a207472616e7366657220746f20746865207a65726f206164647260448201526265737360e81b60648201526084016103de565b6001600160a01b0383166000908152602081905260409020548181101561068f5760405162461bcd60e51b815260206004820152602660248201527f45524332303a207472616e7366657220616d6f756e7420657863656564732062604482015265616c616e636560d01b60648201526084016103de565b6001600160a01b03848116600081815260208181526040808320878703905593871680835291849020805487019055925185815290927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef910160405180910390a35b50505050565b60006107038484610402565b905060001981146106f1578181101561075e5760405162461bcd60e51b815260206004820152601d60248201527f45524332303a20696e73756666696369656e7420616c6c6f77616e636500000060448201526064016103de565b6106f1848484840361042d565b6001600160a01b0382166107c15760405162461bcd60e51b815260206004820152601f60248201527f45524332303a206d696e7420746f20746865207a65726f20616464726573730060448201526064016103de565b80600260008282546107d39190610abb565b90915550506001600160a01b038216600081815260208181526040808320805486019055518481527fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef910160405180910390a35050565b6001600160a01b03821661088a5760405162461bcd60e51b815260206004820152602160248201527f45524332303a206275726e2066726f6d20746865207a65726f206164647265736044820152607360f81b60648201526084016103de565b6001600160a01b038216600090815260208190526040902054818110156108fe5760405162461bcd60e51b815260206004820152602260248201527f45524332303a206275726e20616d6f756e7420657863656564732062616c616e604482015261636560f01b60648201526084016103de565b6001600160a01b0383166000818152602081815260408083208686039055600280548790039055518581529192917fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef910160405180910390a3505050565b600060208083528351808285015260005b818110156109895785810183015185820160400152820161096d565b506000604082860101526040601f19601f8301168501019250505092915050565b80356001600160a01b03811681146109c157600080fd5b919050565b600080604083850312156109d957600080fd5b6109e2836109aa565b946020939093013593505050565b600080600060608486031215610a0557600080fd5b610a0e846109aa565b9250610a1c602085016109aa565b9150604084013590509250925092565b600060208284031215610a3e57600080fd5b610a47826109aa565b9392505050565b60008060408385031215610a6157600080fd5b610a6a836109aa565b9150610a78602084016109aa565b90509250929050565b600181811c90821680610a9557607f821691505b602082108103610ab557634e487b7160e01b600052602260045260246000fd5b50919050565b808201808211156102e657634e487b7160e01b600052601160045260246000fdfea2646970667358221220e64cf87a27edc11a020470156015966c538c7ea1d1a8a6879dd867d9ceb9519864736f6c63430008130033",
}

// TestERC20ABI is the input ABI used to generate the binding from.
// Deprecated: Use TestERC20MetaData.ABI instead.
var TestERC20ABI = TestERC20MetaData.ABI

// TestERC20Bin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use TestERC20MetaData.Bin instead.
var TestERC20Bin = TestERC20MetaData.Bin

// DeployTestERC20 deploys a new Ethereum contract, binding an instance of TestERC20 to it.
func DeployTestERC20(auth *bind.TransactOpts, backend bind.ContractBackend, name string, symbol string, numDecimals uint8, initialAccount common.Address, initialBalance *big.Int) (common.Address, *types.Transaction, *TestERC20, error) {
	parsed, err := TestERC20MetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(TestERC20Bin), backend, name, symbol, numDecimals, initialAccount, initialBalance)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &TestERC20{TestERC20Caller: TestERC20Caller{contract: contract}, TestERC20Transactor: TestERC20Transactor{contract: contract}, TestERC20Filterer: TestERC20Filterer{contract: contract}}, nil
}

// TestERC20 is an auto generated Go binding around an Ethereum contract.
type TestERC20 struct {
	TestERC20Caller     // Read-only binding to the contract
	TestERC20Transactor // Write-only binding to the contract
	TestERC20Filterer   // Log filterer for contract events
}

// TestERC20Caller is an auto generated read-only Go binding around an Ethereum contract.
type TestERC20Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestERC20Transactor is an auto generated write-only Go binding around an Ethereum contract.
type TestERC20Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestERC20Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TestERC20Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestERC20Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TestERC20Session struct {
	Contract     *TestERC20        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TestERC20CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TestERC20CallerSession struct {
	Contract *TestERC20Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// TestERC20TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TestERC20TransactorSession struct {
	Contract     *TestERC20Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// TestERC20Raw is an auto generated low-level Go binding around an Ethereum contract.
type TestERC20Raw struct {
	Contract *TestERC20 // Generic contract binding to access the raw methods on
}

// TestERC20CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TestERC20CallerRaw struct {
	Contract *TestERC20Caller // Generic read-only contract binding to access the raw methods on
}

// TestERC20TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TestERC20TransactorRaw struct {
	Contract *TestERC20Transactor // Generic write-only contract binding to access the raw methods on
}

// NewTestERC20 creates a new instance of TestERC20, bound to a specific deployed contract.
func NewTestERC20(address common.Address, backend bind.ContractBackend) (*TestERC20, error) {
	contract, err := bindTestERC20(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TestERC20{TestERC20Caller: TestERC20Caller{contract: contract}, TestERC20Transactor: TestERC20Transactor{contract: contract}, TestERC20Filterer: TestERC20Filterer{contract: contract}}, nil
}

// NewTestERC20Caller creates a new read-only instance of TestERC20, bound to a specific deployed contract.
func NewTestERC20Caller(address common.Address, caller bind.ContractCaller) (*TestERC20Caller, error) {
	contract, err := bindTestERC20(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TestERC20Caller{contract: contract}, nil
}

// NewTestERC20Transactor creates a new write-only instance of TestERC20, bound to a specific deployed contract.
func NewTestERC20Transactor(address common.Address, transactor bind.ContractTransactor) (*TestERC20Transactor, error) {
	contract, err := bindTestERC20(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TestERC20Transactor{contract: contract}, nil
}

// NewTestERC20Filterer creates a new log filterer instance of TestERC20, bound to a specific deployed contract.
func NewTestERC20Filterer(address common.Address, filterer bind.ContractFilterer) (*TestERC20Filterer, error) {
	contract, err := bindTestERC20(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TestERC20Filterer{contract: contract}, nil
}

// bindTestERC20 binds a generic wrapper to an already deployed contract.
func bindTestERC20(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TestERC20MetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestERC20 *TestERC20Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestERC20.Contract.TestERC20Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestERC20 *TestERC20Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestERC20.Contract.TestERC20Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestERC20 *TestERC20Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestERC20.Contract.TestERC20Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestERC20 *TestERC20CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestERC20.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestERC20 *TestERC20TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestERC20.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestERC20 *TestERC20TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestERC20.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_TestERC20 *TestERC20Caller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _TestERC20.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_TestERC20 *TestERC20Session) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _TestERC20.Contract.Allowance(&_TestERC20.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_TestERC20 *TestERC20CallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _TestERC20.Contract.Allowance(&_TestERC20.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_TestERC20 *TestERC20Caller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _TestERC20.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_TestERC20 *TestERC20Session) BalanceOf(account common.Address) (*big.Int, error) {
	return _TestERC20.Contract.BalanceOf(&_TestERC20.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_TestERC20 *TestERC20CallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _TestERC20.Contract.BalanceOf(&_TestERC20.CallOpts, account)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_TestERC20 *TestERC20Caller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _TestERC20.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_TestERC20 *TestERC20Session) Decimals() (uint8, error) {
	return _TestERC20.Contract.Decimals(&_TestERC20.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_TestERC20 *TestERC20CallerSession) Decimals() (uint8, error) {
	return _TestERC20.Contract.Decimals(&_TestERC20.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_TestERC20 *TestERC20Caller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _TestERC20.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_TestERC20 *TestERC20Session) Name() (string, error) {
	return _TestERC20.Contract.Name(&_TestERC20.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_TestERC20 *TestERC20CallerSession) Name() (string, error) {
	return _TestERC20.Contract.Name(&_TestERC20.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_TestERC20 *TestERC20Caller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _TestERC20.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_TestERC20 *TestERC20Session) Symbol() (string, error) {
	return _TestERC20.Contract.Symbol(&_TestERC20.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_TestERC20 *TestERC20CallerSession) Symbol() (string, error) {
	return _TestERC20.Contract.Symbol(&_TestERC20.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_TestERC20 *TestERC20Caller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _TestERC20.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_TestERC20 *TestERC20Session) TotalSupply() (*big.Int, error) {
	return _TestERC20.Contract.TotalSupply(&_TestERC20.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_TestERC20 *TestERC20CallerSession) TotalSupply() (*big.Int, error) {
	return _TestERC20.Contract.TotalSupply(&_TestERC20.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_TestERC20 *TestERC20Transactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_TestERC20 *TestERC20Session) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.Approve(&_TestERC20.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_TestERC20 *TestERC20TransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.Approve(&_TestERC20.TransactOpts, spender, amount)
}

// ApproveInternal is a paid mutator transaction binding the contract method 0x56189cb4.
//
// Solidity: function approveInternal(address owner, address spender, uint256 value) returns()
func (_TestERC20 *TestERC20Transactor) ApproveInternal(opts *bind.TransactOpts, owner common.Address, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestERC20.contract.Transact(opts, "approveInternal", owner, spender, value)
}

// ApproveInternal is a paid mutator transaction binding the contract method 0x56189cb4.
//
// Solidity: function approveInternal(address owner, address spender, uint256 value) returns()
func (_TestERC20 *TestERC20Session) ApproveInternal(owner common.Address, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.ApproveInternal(&_TestERC20.TransactOpts, owner, spender, value)
}

// ApproveInternal is a paid mutator transaction binding the contract method 0x56189cb4.
//
// Solidity: function approveInternal(address owner, address spender, uint256 value) returns()
func (_TestERC20 *TestERC20TransactorSession) ApproveInternal(owner common.Address, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.ApproveInternal(&_TestERC20.TransactOpts, owner, spender, value)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address account, uint256 amount) returns()
func (_TestERC20 *TestERC20Transactor) Burn(opts *bind.TransactOpts, account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.contract.Transact(opts, "burn", account, amount)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address account, uint256 amount) returns()
func (_TestERC20 *TestERC20Session) Burn(account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.Burn(&_TestERC20.TransactOpts, account, amount)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address account, uint256 amount) returns()
func (_TestERC20 *TestERC20TransactorSession) Burn(account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.Burn(&_TestERC20.TransactOpts, account, amount)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_TestERC20 *TestERC20Transactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _TestERC20.contract.Transact(opts, "decreaseAllowance", spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_TestERC20 *TestERC20Session) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.DecreaseAllowance(&_TestERC20.TransactOpts, spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_TestERC20 *TestERC20TransactorSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.DecreaseAllowance(&_TestERC20.TransactOpts, spender, subtractedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_TestERC20 *TestERC20Transactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _TestERC20.contract.Transact(opts, "increaseAllowance", spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_TestERC20 *TestERC20Session) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.IncreaseAllowance(&_TestERC20.TransactOpts, spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_TestERC20 *TestERC20TransactorSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.IncreaseAllowance(&_TestERC20.TransactOpts, spender, addedValue)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address account, uint256 amount) returns()
func (_TestERC20 *TestERC20Transactor) Mint(opts *bind.TransactOpts, account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.contract.Transact(opts, "mint", account, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address account, uint256 amount) returns()
func (_TestERC20 *TestERC20Session) Mint(account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.Mint(&_TestERC20.TransactOpts, account, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address account, uint256 amount) returns()
func (_TestERC20 *TestERC20TransactorSession) Mint(account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.Mint(&_TestERC20.TransactOpts, account, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_TestERC20 *TestERC20Transactor) Transfer(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.contract.Transact(opts, "transfer", to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_TestERC20 *TestERC20Session) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.Transfer(&_TestERC20.TransactOpts, to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_TestERC20 *TestERC20TransactorSession) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.Transfer(&_TestERC20.TransactOpts, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_TestERC20 *TestERC20Transactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.contract.Transact(opts, "transferFrom", from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_TestERC20 *TestERC20Session) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.TransferFrom(&_TestERC20.TransactOpts, from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_TestERC20 *TestERC20TransactorSession) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.TransferFrom(&_TestERC20.TransactOpts, from, to, amount)
}

// TransferInternal is a paid mutator transaction binding the contract method 0x222f5be0.
//
// Solidity: function transferInternal(address from, address to, uint256 value) returns()
func (_TestERC20 *TestERC20Transactor) TransferInternal(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestERC20.contract.Transact(opts, "transferInternal", from, to, value)
}

// TransferInternal is a paid mutator transaction binding the contract method 0x222f5be0.
//
// Solidity: function transferInternal(address from, address to, uint256 value) returns()
func (_TestERC20 *TestERC20Session) TransferInternal(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.TransferInternal(&_TestERC20.TransactOpts, from, to, value)
}

// TransferInternal is a paid mutator transaction binding the contract method 0x222f5be0.
//
// Solidity: function transferInternal(address from, address to, uint256 value) returns()
func (_TestERC20 *TestERC20TransactorSession) TransferInternal(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestERC20.Contract.TransferInternal(&_TestERC20.TransactOpts, from, to, value)
}

// TestERC20ApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the TestERC20 contract.
type TestERC20ApprovalIterator struct {
	Event *TestERC20Approval // Event containing the contract specifics and raw log

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
func (it *TestERC20ApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TestERC20Approval)
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
		it.Event = new(TestERC20Approval)
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
func (it *TestERC20ApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TestERC20ApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TestERC20Approval represents a Approval event raised by the TestERC20 contract.
type TestERC20Approval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_TestERC20 *TestERC20Filterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*TestERC20ApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _TestERC20.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &TestERC20ApprovalIterator{contract: _TestERC20.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_TestERC20 *TestERC20Filterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *TestERC20Approval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _TestERC20.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TestERC20Approval)
				if err := _TestERC20.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_TestERC20 *TestERC20Filterer) ParseApproval(log types.Log) (*TestERC20Approval, error) {
	event := new(TestERC20Approval)
	if err := _TestERC20.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TestERC20TransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the TestERC20 contract.
type TestERC20TransferIterator struct {
	Event *TestERC20Transfer // Event containing the contract specifics and raw log

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
func (it *TestERC20TransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TestERC20Transfer)
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
		it.Event = new(TestERC20Transfer)
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
func (it *TestERC20TransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TestERC20TransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TestERC20Transfer represents a Transfer event raised by the TestERC20 contract.
type TestERC20Transfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_TestERC20 *TestERC20Filterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*TestERC20TransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _TestERC20.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &TestERC20TransferIterator{contract: _TestERC20.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_TestERC20 *TestERC20Filterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *TestERC20Transfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _TestERC20.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TestERC20Transfer)
				if err := _TestERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_TestERC20 *TestERC20Filterer) ParseTransfer(log types.Log) (*TestERC20Transfer, error) {
	event := new(TestERC20Transfer)
	if err := _TestERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
