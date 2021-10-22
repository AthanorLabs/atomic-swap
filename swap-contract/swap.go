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
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_hashRedeem\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_expectedPublicKey\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_hashRefund\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"x\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"name\":\"DerivedPubKeyRedeem\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"expectedPublicKey\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hashRedeem\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hashRefund\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ready\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_s\",\"type\":\"uint256\"}],\"name\":\"redeem\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_s\",\"type\":\"uint256\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040526000600560146101000a81548160ff0219169083151502179055506040516110e73803806110e78339818101604052606081101561004157600080fd5b8101908080519060200190929190805190602001909291908051906020019092919050505033600560006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508260018190555081600281905550806003819055506201518042016004819055506040516100d490610138565b604051809103906000f0801580156100f0573d6000803e3d6000fd5b506000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550505050610145565b610ac58061062283390190565b6104ce806101546000396000f3fe608060405234801561001057600080fd5b50600436106100625760003560e01c8063278ecde1146100675780636defbf80146100955780638475b9831461009f578063db006a75146100bd578063eb7fd865146100eb578063f911f90814610109575b600080fd5b6100936004803603602081101561007d57600080fd5b8101908080359060200190929190505050610127565b005b61009d6101cd565b005b6100a7610244565b6040518082815260200191505060405180910390f35b6100e9600480360360208110156100d357600080fd5b810190808035906020019092919050505061024a565b005b6100f3610461565b6040518082815260200191505060405180910390f35b610111610467565b6040518082815260200191505060405180910390f35b60008160405160200180828152602001915050604051602081830303815290604052805190602001209050600354811461016057600080fd5b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc479081150290604051600060405180830381858888f193505050501580156101c8573d6000803e3d6000fd5b505050565b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461022757600080fd5b6001600560146101000a81548160ff021916908315150217905550565b60035481565b60011515600560149054906101000a900460ff161515146102d3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f636f6e7472616374206973206e6f74207265616479210000000000000000000081525060200191505060405180910390fd5b60008060008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663bc9e2bcf846040518263ffffffff1660e01b815260040180828152602001915050604080518083038186803b15801561034657600080fd5b505afa15801561035a573d6000803e3d6000fd5b505050506040513d604081101561037057600080fd5b810190808051906020019092919080519060200190929190505050915091507f52f043137a9cbb6eae4c9ef496f9a343d5d5cc6d374e41bb842436c881fb19d78282604051808381526020018281526020019250505060405180910390a1600082826040516020018083815260200182815260200192505050604051602081830303815290604052805190602001209050600254811461045b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602b81526020018061046e602b913960400191505060405180910390fd5b50505050565b60025481565b6001548156fe70726f7669646564207075626c6963206b657920646f6573206e6f74206d61746368206578706563746564a2646970667358221220cacec9f0798cc2efe94fc8b610d0f8e13847cfd33ba84d200b4143fca1a9514b64736f6c634300060c0033608060405234801561001057600080fd5b50610aa5806100206000396000f3fe608060405234801561001057600080fd5b50600436106100625760003560e01c806303a507be146100675780635727dc5c146100855780637a308a4c146100a3578063997da8d4146100c1578063bc9e2bcf146100df578063eeeac01e14610128575b600080fd5b61006f610146565b6040518082815260200191505060405180910390f35b61008d61016a565b6040518082815260200191505060405180910390f35b6100ab61016f565b6040518082815260200191505060405180910390f35b6100c9610193565b6040518082815260200191505060405180910390f35b61010b600480360360208110156100f557600080fd5b8101908080359060200190929190505050610198565b604051808381526020018281526020019250505060405180910390f35b610130610212565b6040518082815260200191505060405180910390f35b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179881565b600781565b7f483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b881565b600081565b600080610209837f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f817987f483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b860007ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f610236565b91509150915091565b7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f81565b600080600080600061024d8a8a8a60018b8b610270565b92509250925061025f8383838961030e565b945094505050509550959350505050565b60008060008089141561028b57878787925092509250610302565b60008990506000806000600190505b600084146102f457600060018516146102c9576102bc8383838f8f8f8e610369565b8093508194508295505050505b600284816102d357fe5b0493506102e38c8c8c8c8c61083a565b809c50819d50829e5050505061029a565b828282965096509650505050505b96509650969350505050565b600080600061031d8585610959565b90506000848061032957fe5b82830990506000858061033857fe5b828a0990506000868061034757fe5b878061034f57fe5b8486098a0990508181955095505050505094509492505050565b6000806000808a14801561037d5750600089145b156103905786868692509250925061082d565b6000871480156103a05750600086145b156103b35789898992509250925061082d565b6103bb610a4d565b84806103c357fe5b898a09816000600481106103d357fe5b60200201818152505084806103e457fe5b816000600481106103f157fe5b60200201518a098160016004811061040557fe5b602002018181525050848061041657fe5b8687098160026004811061042657fe5b602002018181525050848061043757fe5b8160026004811061044457fe5b602002015187098160036004811061045857fe5b6020020181815250506040518060800160405280868061047457fe5b8360026004811061048157fe5b60200201518e098152602001868061049557fe5b836003600481106104a257fe5b60200201518d09815260200186806104b657fe5b836000600481106104c357fe5b60200201518b09815260200186806104d757fe5b836001600481106104e457fe5b60200201518a098152509050806002600481106104fd57fe5b60200201518160006004811061050f57fe5b602002015114158061054357508060036004811061052957fe5b60200201518160016004811061053b57fe5b602002015114155b6105b5576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601e8152602001807f557365206a6163446f75626c652066756e6374696f6e20696e7374656164000081525060200191505060405180910390fd5b6105bd610a4d565b85806105c557fe5b826000600481106105d257fe5b60200201518703836002600481106105e657fe5b602002015108816000600481106105f957fe5b602002018181525050858061060a57fe5b8260016004811061061757fe5b602002015187038360036004811061062b57fe5b6020020151088160016004811061063e57fe5b602002018181525050858061064f57fe5b8160006004811061065c57fe5b60200201518260006004811061066e57fe5b6020020151098160026004811061068157fe5b602002018181525050858061069257fe5b8160006004811061069f57fe5b6020020151826002600481106106b157fe5b602002015109816003600481106106c457fe5b602002018181525050600086806106d757fe5b826003600481106106e457fe5b6020020151880388806106f357fe5b8460016004811061070057fe5b60200201518560016004811061071257fe5b602002015109089050868061072357fe5b878061072b57fe5b888061073357fe5b8460026004811061074057fe5b60200201518660006004811061075257fe5b6020020151096002098803820890506000878061076b57fe5b888061077357fe5b838a038a8061077e57fe5b8660026004811061078b57fe5b60200201518860006004811061079d57fe5b60200201510908846001600481106107b157fe5b602002015109905087806107c157fe5b88806107c957fe5b846003600481106107d657fe5b6020020151866001600481106107e857fe5b602002015109890382089050600088806107fe57fe5b898061080657fe5b8b8f098560006004811061081657fe5b602002015109905082828297509750975050505050505b9750975097945050505050565b6000806000808614156108555787878792509250925061094e565b6000848061085f57fe5b898a0990506000858061086e57fe5b898a0990506000868061087d57fe5b898a0990506000878061088c57fe5b888061089457fe5b848e096004099050600088806108a657fe5b89806108ae57fe5b8a806108b657fe5b8586098c098a806108c357fe5b8760030908905088806108d257fe5b89806108da57fe5b8384088a038a806108e757fe5b83840908945088806108f557fe5b89806108fd57fe5b8a8061090557fe5b8687096008098a038a8061091557fe5b8b8061091d57fe5b888d0386088409089350888061092f57fe5b898061093757fe5b8c8e09600209925084848497509750975050505050505b955095509592505050565b600080831415801561096b5750818314155b8015610978575060008214155b6109ea576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252600e8152602001807f496e76616c6964206e756d62657200000000000000000000000000000000000081525060200191505060405180910390fd5b60008060019050600084905060005b60008714610a4057868281610a0a57fe5b049050828680610a1657fe5b8780610a1e57fe5b85840988038608809450819550505086878202830380985081935050506109f9565b8394505050505092915050565b604051806080016040528060049060208202803683378082019150509050509056fea2646970667358221220e57b7cdaa48a9ccd1b6d21d846c0118922e28182b022f8558bcf8dcc6480b73164736f6c634300060c0033",
}

// SwapABI is the input ABI used to generate the binding from.
// Deprecated: Use SwapMetaData.ABI instead.
var SwapABI = SwapMetaData.ABI

// SwapBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SwapMetaData.Bin instead.
var SwapBin = SwapMetaData.Bin

// DeploySwap deploys a new Ethereum contract, binding an instance of Swap to it.
func DeploySwap(auth *bind.TransactOpts, backend bind.ContractBackend, _hashRedeem [32]byte, _expectedPublicKey [32]byte, _hashRefund [32]byte) (common.Address, *types.Transaction, *Swap, error) {
	parsed, err := SwapMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SwapBin), backend, _hashRedeem, _expectedPublicKey, _hashRefund)
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

// ExpectedPublicKey is a free data retrieval call binding the contract method 0xeb7fd865.
//
// Solidity: function expectedPublicKey() view returns(bytes32)
func (_Swap *SwapCaller) ExpectedPublicKey(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "expectedPublicKey")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ExpectedPublicKey is a free data retrieval call binding the contract method 0xeb7fd865.
//
// Solidity: function expectedPublicKey() view returns(bytes32)
func (_Swap *SwapSession) ExpectedPublicKey() ([32]byte, error) {
	return _Swap.Contract.ExpectedPublicKey(&_Swap.CallOpts)
}

// ExpectedPublicKey is a free data retrieval call binding the contract method 0xeb7fd865.
//
// Solidity: function expectedPublicKey() view returns(bytes32)
func (_Swap *SwapCallerSession) ExpectedPublicKey() ([32]byte, error) {
	return _Swap.Contract.ExpectedPublicKey(&_Swap.CallOpts)
}

// HashRedeem is a free data retrieval call binding the contract method 0xf911f908.
//
// Solidity: function hashRedeem() view returns(bytes32)
func (_Swap *SwapCaller) HashRedeem(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "hashRedeem")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HashRedeem is a free data retrieval call binding the contract method 0xf911f908.
//
// Solidity: function hashRedeem() view returns(bytes32)
func (_Swap *SwapSession) HashRedeem() ([32]byte, error) {
	return _Swap.Contract.HashRedeem(&_Swap.CallOpts)
}

// HashRedeem is a free data retrieval call binding the contract method 0xf911f908.
//
// Solidity: function hashRedeem() view returns(bytes32)
func (_Swap *SwapCallerSession) HashRedeem() ([32]byte, error) {
	return _Swap.Contract.HashRedeem(&_Swap.CallOpts)
}

// HashRefund is a free data retrieval call binding the contract method 0x8475b983.
//
// Solidity: function hashRefund() view returns(bytes32)
func (_Swap *SwapCaller) HashRefund(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "hashRefund")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HashRefund is a free data retrieval call binding the contract method 0x8475b983.
//
// Solidity: function hashRefund() view returns(bytes32)
func (_Swap *SwapSession) HashRefund() ([32]byte, error) {
	return _Swap.Contract.HashRefund(&_Swap.CallOpts)
}

// HashRefund is a free data retrieval call binding the contract method 0x8475b983.
//
// Solidity: function hashRefund() view returns(bytes32)
func (_Swap *SwapCallerSession) HashRefund() ([32]byte, error) {
	return _Swap.Contract.HashRefund(&_Swap.CallOpts)
}

// Ready is a paid mutator transaction binding the contract method 0x6defbf80.
//
// Solidity: function ready() returns()
func (_Swap *SwapTransactor) Ready(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "ready")
}

// Ready is a paid mutator transaction binding the contract method 0x6defbf80.
//
// Solidity: function ready() returns()
func (_Swap *SwapSession) Ready() (*types.Transaction, error) {
	return _Swap.Contract.Ready(&_Swap.TransactOpts)
}

// Ready is a paid mutator transaction binding the contract method 0x6defbf80.
//
// Solidity: function ready() returns()
func (_Swap *SwapTransactorSession) Ready() (*types.Transaction, error) {
	return _Swap.Contract.Ready(&_Swap.TransactOpts)
}

// Redeem is a paid mutator transaction binding the contract method 0xdb006a75.
//
// Solidity: function redeem(uint256 _s) returns()
func (_Swap *SwapTransactor) Redeem(opts *bind.TransactOpts, _s *big.Int) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "redeem", _s)
}

// Redeem is a paid mutator transaction binding the contract method 0xdb006a75.
//
// Solidity: function redeem(uint256 _s) returns()
func (_Swap *SwapSession) Redeem(_s *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.Redeem(&_Swap.TransactOpts, _s)
}

// Redeem is a paid mutator transaction binding the contract method 0xdb006a75.
//
// Solidity: function redeem(uint256 _s) returns()
func (_Swap *SwapTransactorSession) Redeem(_s *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.Redeem(&_Swap.TransactOpts, _s)
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

// SwapDerivedPubKeyRedeemIterator is returned from FilterDerivedPubKeyRedeem and is used to iterate over the raw logs and unpacked data for DerivedPubKeyRedeem events raised by the Swap contract.
type SwapDerivedPubKeyRedeemIterator struct {
	Event *SwapDerivedPubKeyRedeem // Event containing the contract specifics and raw log

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
func (it *SwapDerivedPubKeyRedeemIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapDerivedPubKeyRedeem)
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
		it.Event = new(SwapDerivedPubKeyRedeem)
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
func (it *SwapDerivedPubKeyRedeemIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapDerivedPubKeyRedeemIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapDerivedPubKeyRedeem represents a DerivedPubKeyRedeem event raised by the Swap contract.
type SwapDerivedPubKeyRedeem struct {
	X   *big.Int
	Y   *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterDerivedPubKeyRedeem is a free log retrieval operation binding the contract event 0x52f043137a9cbb6eae4c9ef496f9a343d5d5cc6d374e41bb842436c881fb19d7.
//
// Solidity: event DerivedPubKeyRedeem(uint256 x, uint256 y)
func (_Swap *SwapFilterer) FilterDerivedPubKeyRedeem(opts *bind.FilterOpts) (*SwapDerivedPubKeyRedeemIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "DerivedPubKeyRedeem")
	if err != nil {
		return nil, err
	}
	return &SwapDerivedPubKeyRedeemIterator{contract: _Swap.contract, event: "DerivedPubKeyRedeem", logs: logs, sub: sub}, nil
}

// WatchDerivedPubKeyRedeem is a free log subscription operation binding the contract event 0x52f043137a9cbb6eae4c9ef496f9a343d5d5cc6d374e41bb842436c881fb19d7.
//
// Solidity: event DerivedPubKeyRedeem(uint256 x, uint256 y)
func (_Swap *SwapFilterer) WatchDerivedPubKeyRedeem(opts *bind.WatchOpts, sink chan<- *SwapDerivedPubKeyRedeem) (event.Subscription, error) {

	logs, sub, err := _Swap.contract.WatchLogs(opts, "DerivedPubKeyRedeem")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapDerivedPubKeyRedeem)
				if err := _Swap.contract.UnpackLog(event, "DerivedPubKeyRedeem", log); err != nil {
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

// ParseDerivedPubKeyRedeem is a log parse operation binding the contract event 0x52f043137a9cbb6eae4c9ef496f9a343d5d5cc6d374e41bb842436c881fb19d7.
//
// Solidity: event DerivedPubKeyRedeem(uint256 x, uint256 y)
func (_Swap *SwapFilterer) ParseDerivedPubKeyRedeem(log types.Log) (*SwapDerivedPubKeyRedeem, error) {
	event := new(SwapDerivedPubKeyRedeem)
	if err := _Swap.contract.UnpackLog(event, "DerivedPubKeyRedeem", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
