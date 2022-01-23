// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package swapfactory

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

// SwapFactoryABI is the input ABI used to generate the binding from.
const SwapFactoryABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"swapID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"swapID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"claimKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"refundKey\",\"type\":\"bytes32\"}],\"name\":\"New\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"swapID\",\"type\":\"uint256\"}],\"name\":\"Ready\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"swapID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"addresspayable\",\"name\":\"_claimer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_timeoutDuration\",\"type\":\"uint256\"}],\"name\":\"new_swap\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"set_ready\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"swaps\",\"outputs\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout_0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout_1\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isReady\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"completed\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// SwapFactoryBin is the compiled bytecode used for deploying new contracts.
var SwapFactoryBin = "0x60a060405234801561001057600080fd5b5060405161001d90610072565b604051809103906000f080158015610039573d6000803e3d6000fd5b5073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff168152505061007f565b610393806113fb83390190565b60805161136161009a6000396000610af901526113616000f3fe60806040526004361061004a5760003560e01c80630deeecba1461004f57806331d144571461007f57806337da2ecf146100a857806371eedb88146100d1578063f09c5829146100fa575b600080fd5b61006960048036038101906100649190610d2d565b61013f565b6040516100769190610da3565b60405180910390f35b34801561008b57600080fd5b506100a660048036038101906100a19190610dbe565b61038d565b005b3480156100b457600080fd5b506100cf60048036038101906100ca9190610dfe565b61068b565b005b3480156100dd57600080fd5b506100f860048036038101906100f39190610dbe565b6107bd565b005b34801561010657600080fd5b50610121600480360381019061011c9190610dfe565b610a4f565b60405161013699989796959493929190610e64565b60405180910390f35b600080600054905061014f610bdc565b33816000019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505084816020019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff1681525050868160400181815250508581606001818152505083426101df9190610f20565b8160800181815250506002846101f59190610f76565b426102009190610f20565b8160a001818152505034816101000181815250507f982a99d883f17ecd5797205d5b3674205d7882bb28a9487d736d3799422cd05582888860405161024793929190610fd0565b60405180910390a160016000808282546102619190610f20565b92505081905550806001600084815260200190815260200160002060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160010160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060408201518160020155606082015181600301556080820151816004015560a0820151816005015560c08201518160060160006101000a81548160ff02191690831515021790555060e08201518160060160016101000a81548160ff02191690831515021790555061010082015181600701559050508192505050949350505050565b600060016000848152602001908152602001600020604051806101200160405290816000820160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020016001820160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001600282015481526020016003820154815260200160048201548152602001600582015481526020016006820160009054906101000a900460ff161515151581526020016006820160019054906101000a900460ff1615151515815260200160078201548152505090508060e00151156104d357600080fd5b806020015173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610545576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161053c90611064565b60405180910390fd5b80608001514210158061055957508060c001515b610598576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161058f906110d0565b60405180910390fd5b8060a0015142106105de576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105d59061113c565b60405180910390fd5b6105ec828260400151610af7565b7fd5a2476fc450083bbb092dd3f4be92698ffdc2d213e6f1e730c7f44a52f1ccfc838360405161061d92919061115c565b60405180910390a1806020015173ffffffffffffffffffffffffffffffffffffffff166108fc8261010001519081150290604051600060405180830381858888f19350505050158015610674573d6000803e3d6000fd5b5060018160e0019015159081151581525050505050565b6001600082815260200190815260200160002060060160019054906101000a900460ff16156106b957600080fd5b3373ffffffffffffffffffffffffffffffffffffffff166001600083815260200190815260200160002060000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161461072757600080fd5b6001600082815260200190815260200160002060060160009054906101000a900460ff161561075557600080fd5b600180600083815260200190815260200160002060060160006101000a81548160ff0219169083151502179055507f0b217ad5c70346c7cd952bd2463c6684a56f9ed229f5780947586625781b4770816040516107b29190610da3565b60405180910390a150565b600060016000848152602001908152602001600020604051806101200160405290816000820160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020016001820160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001600282015481526020016003820154815260200160048201548152602001600582015481526020016006820160009054906101000a900460ff161515151581526020016006820160019054906101000a900460ff1615151515815260200160078201548152505090508060e001511561090357600080fd5b806000015173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461093f57600080fd5b8060a0015142101580610963575080608001514210801561096257508060c00151155b5b6109a2576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610999906111f7565b60405180910390fd5b6109b0828260600151610af7565b7f4fd30f3ee0d64f7eaa62d0e005ca64c6a560652156d6c33f23ea8ca4936106e083836040516109e192919061115c565b60405180910390a1806000015173ffffffffffffffffffffffffffffffffffffffff166108fc8261010001519081150290604051600060405180830381858888f19350505050158015610a38573d6000803e3d6000fd5b5060018160e0019015159081151581525050505050565b60016020528060005260406000206000915090508060000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16908060010160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16908060020154908060030154908060040154908060050154908060060160009054906101000a900460ff16908060060160019054906101000a900460ff16908060070154905089565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663b32d1b4f8360001c8360001c6040518363ffffffff1660e01b8152600401610b58929190611217565b602060405180830381865afa158015610b75573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610b99919061126c565b610bd8576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610bcf9061130b565b60405180910390fd5b5050565b604051806101200160405280600073ffffffffffffffffffffffffffffffffffffffff168152602001600073ffffffffffffffffffffffffffffffffffffffff16815260200160008019168152602001600080191681526020016000815260200160008152602001600015158152602001600015158152602001600081525090565b600080fd5b6000819050919050565b610c7681610c63565b8114610c8157600080fd5b50565b600081359050610c9381610c6d565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610cc482610c99565b9050919050565b610cd481610cb9565b8114610cdf57600080fd5b50565b600081359050610cf181610ccb565b92915050565b6000819050919050565b610d0a81610cf7565b8114610d1557600080fd5b50565b600081359050610d2781610d01565b92915050565b60008060008060808587031215610d4757610d46610c5e565b5b6000610d5587828801610c84565b9450506020610d6687828801610c84565b9350506040610d7787828801610ce2565b9250506060610d8887828801610d18565b91505092959194509250565b610d9d81610cf7565b82525050565b6000602082019050610db86000830184610d94565b92915050565b60008060408385031215610dd557610dd4610c5e565b5b6000610de385828601610d18565b9250506020610df485828601610c84565b9150509250929050565b600060208284031215610e1457610e13610c5e565b5b6000610e2284828501610d18565b91505092915050565b610e3481610cb9565b82525050565b610e4381610c63565b82525050565b60008115159050919050565b610e5e81610e49565b82525050565b600061012082019050610e7a600083018c610e2b565b610e87602083018b610e2b565b610e94604083018a610e3a565b610ea16060830189610e3a565b610eae6080830188610d94565b610ebb60a0830187610d94565b610ec860c0830186610e55565b610ed560e0830185610e55565b610ee3610100830184610d94565b9a9950505050505050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610f2b82610cf7565b9150610f3683610cf7565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03821115610f6b57610f6a610ef1565b5b828201905092915050565b6000610f8182610cf7565b9150610f8c83610cf7565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0483118215151615610fc557610fc4610ef1565b5b828202905092915050565b6000606082019050610fe56000830186610d94565b610ff26020830185610e3a565b610fff6040830184610e3a565b949350505050565b600082825260208201905092915050565b7f6f6e6c7920636c61696d65722063616e20636c61696d21000000000000000000600082015250565b600061104e601783611007565b915061105982611018565b602082019050919050565b6000602082019050818103600083015261107d81611041565b9050919050565b7f746f6f206561726c7920746f20636c61696d2100000000000000000000000000600082015250565b60006110ba601383611007565b91506110c582611084565b602082019050919050565b600060208201905081810360008301526110e9816110ad565b9050919050565b7f746f6f206c61746520746f20636c61696d210000000000000000000000000000600082015250565b6000611126601283611007565b9150611131826110f0565b602082019050919050565b6000602082019050818103600083015261115581611119565b9050919050565b60006040820190506111716000830185610d94565b61117e6020830184610e3a565b9392505050565b7f697427732074686520636f756e74657270617274792773207475726e2c20756e60008201527f61626c6520746f20726566756e642c2074727920616761696e206c6174657200602082015250565b60006111e1603f83611007565b91506111ec82611185565b604082019050919050565b60006020820190508181036000830152611210816111d4565b9050919050565b600060408201905061122c6000830185610d94565b6112396020830184610d94565b9392505050565b61124981610e49565b811461125457600080fd5b50565b60008151905061126681611240565b92915050565b60006020828403121561128257611281610c5e565b5b600061129084828501611257565b91505092915050565b7f70726f76696465642073656372657420646f6573206e6f74206d61746368207460008201527f6865206578706563746564207075626c6963206b657900000000000000000000602082015250565b60006112f5603683611007565b915061130082611299565b604082019050919050565b60006020820190508181036000830152611324816112e8565b905091905056fea2646970667358221220afc21ec69157c906ffd7db42328083d416376af2172823b6f0f7c3e94765fdc564736f6c634300080a0033608060405234801561001057600080fd5b50610373806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c8063b32d1b4f14610030575b600080fd5b61004a600480360381019061004591906101a0565b610060565b60405161005791906101fb565b60405180910390f35b60008060016000601b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179860001b7ffffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141806100bc576100bb610216565b5b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798890960001b604051600081526020016040526040516100ff94939291906102f8565b6020604051602081039080840390855afa158015610121573d6000803e3d6000fd5b5050506020604051035190508073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161491505092915050565b600080fd5b6000819050919050565b61017d8161016a565b811461018857600080fd5b50565b60008135905061019a81610174565b92915050565b600080604083850312156101b7576101b6610165565b5b60006101c58582860161018b565b92505060206101d68582860161018b565b9150509250929050565b60008115159050919050565b6101f5816101e0565b82525050565b600060208201905061021060008301846101ec565b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000819050919050565b6000819050919050565b60008160001b9050919050565b600061028161027c61027784610245565b610259565b61024f565b9050919050565b61029181610266565b82525050565b6000819050919050565b600060ff82169050919050565b6000819050919050565b60006102d36102ce6102c984610297565b6102ae565b6102a1565b9050919050565b6102e3816102b8565b82525050565b6102f28161024f565b82525050565b600060808201905061030d6000830187610288565b61031a60208301866102da565b61032760408301856102e9565b61033460608301846102e9565b9594505050505056fea2646970667358221220366f5349c99cc5aedbea5b41a0bba96eef36652cb460171cf7386f8afd621fde64736f6c634300080a0033"

// DeploySwapFactory deploys a new Ethereum contract, binding an instance of SwapFactory to it.
func DeploySwapFactory(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SwapFactory, error) {
	parsed, err := abi.JSON(strings.NewReader(SwapFactoryABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SwapFactoryBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SwapFactory{SwapFactoryCaller: SwapFactoryCaller{contract: contract}, SwapFactoryTransactor: SwapFactoryTransactor{contract: contract}, SwapFactoryFilterer: SwapFactoryFilterer{contract: contract}}, nil
}

// SwapFactory is an auto generated Go binding around an Ethereum contract.
type SwapFactory struct {
	SwapFactoryCaller     // Read-only binding to the contract
	SwapFactoryTransactor // Write-only binding to the contract
	SwapFactoryFilterer   // Log filterer for contract events
}

// SwapFactoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type SwapFactoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapFactoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SwapFactoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapFactoryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SwapFactoryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapFactorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SwapFactorySession struct {
	Contract     *SwapFactory      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SwapFactoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SwapFactoryCallerSession struct {
	Contract *SwapFactoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// SwapFactoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SwapFactoryTransactorSession struct {
	Contract     *SwapFactoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// SwapFactoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type SwapFactoryRaw struct {
	Contract *SwapFactory // Generic contract binding to access the raw methods on
}

// SwapFactoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SwapFactoryCallerRaw struct {
	Contract *SwapFactoryCaller // Generic read-only contract binding to access the raw methods on
}

// SwapFactoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SwapFactoryTransactorRaw struct {
	Contract *SwapFactoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSwapFactory creates a new instance of SwapFactory, bound to a specific deployed contract.
func NewSwapFactory(address common.Address, backend bind.ContractBackend) (*SwapFactory, error) {
	contract, err := bindSwapFactory(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SwapFactory{SwapFactoryCaller: SwapFactoryCaller{contract: contract}, SwapFactoryTransactor: SwapFactoryTransactor{contract: contract}, SwapFactoryFilterer: SwapFactoryFilterer{contract: contract}}, nil
}

// NewSwapFactoryCaller creates a new read-only instance of SwapFactory, bound to a specific deployed contract.
func NewSwapFactoryCaller(address common.Address, caller bind.ContractCaller) (*SwapFactoryCaller, error) {
	contract, err := bindSwapFactory(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SwapFactoryCaller{contract: contract}, nil
}

// NewSwapFactoryTransactor creates a new write-only instance of SwapFactory, bound to a specific deployed contract.
func NewSwapFactoryTransactor(address common.Address, transactor bind.ContractTransactor) (*SwapFactoryTransactor, error) {
	contract, err := bindSwapFactory(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SwapFactoryTransactor{contract: contract}, nil
}

// NewSwapFactoryFilterer creates a new log filterer instance of SwapFactory, bound to a specific deployed contract.
func NewSwapFactoryFilterer(address common.Address, filterer bind.ContractFilterer) (*SwapFactoryFilterer, error) {
	contract, err := bindSwapFactory(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SwapFactoryFilterer{contract: contract}, nil
}

// bindSwapFactory binds a generic wrapper to an already deployed contract.
func bindSwapFactory(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SwapFactoryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SwapFactory *SwapFactoryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SwapFactory.Contract.SwapFactoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SwapFactory *SwapFactoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwapFactory.Contract.SwapFactoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SwapFactory *SwapFactoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SwapFactory.Contract.SwapFactoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SwapFactory *SwapFactoryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SwapFactory.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SwapFactory *SwapFactoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwapFactory.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SwapFactory *SwapFactoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SwapFactory.Contract.contract.Transact(opts, method, params...)
}

// Swaps is a free data retrieval call binding the contract method 0xf09c5829.
//
// Solidity: function swaps(uint256 ) view returns(address owner, address claimer, bytes32 pubKeyClaim, bytes32 pubKeyRefund, uint256 timeout_0, uint256 timeout_1, bool isReady, bool completed, uint256 value)
func (_SwapFactory *SwapFactoryCaller) Swaps(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Owner        common.Address
	Claimer      common.Address
	PubKeyClaim  [32]byte
	PubKeyRefund [32]byte
	Timeout0     *big.Int
	Timeout1     *big.Int
	IsReady      bool
	Completed    bool
	Value        *big.Int
}, error) {
	var out []interface{}
	err := _SwapFactory.contract.Call(opts, &out, "swaps", arg0)

	outstruct := new(struct {
		Owner        common.Address
		Claimer      common.Address
		PubKeyClaim  [32]byte
		PubKeyRefund [32]byte
		Timeout0     *big.Int
		Timeout1     *big.Int
		IsReady      bool
		Completed    bool
		Value        *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Owner = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Claimer = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.PubKeyClaim = *abi.ConvertType(out[2], new([32]byte)).(*[32]byte)
	outstruct.PubKeyRefund = *abi.ConvertType(out[3], new([32]byte)).(*[32]byte)
	outstruct.Timeout0 = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.Timeout1 = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.IsReady = *abi.ConvertType(out[6], new(bool)).(*bool)
	outstruct.Completed = *abi.ConvertType(out[7], new(bool)).(*bool)
	outstruct.Value = *abi.ConvertType(out[8], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Swaps is a free data retrieval call binding the contract method 0xf09c5829.
//
// Solidity: function swaps(uint256 ) view returns(address owner, address claimer, bytes32 pubKeyClaim, bytes32 pubKeyRefund, uint256 timeout_0, uint256 timeout_1, bool isReady, bool completed, uint256 value)
func (_SwapFactory *SwapFactorySession) Swaps(arg0 *big.Int) (struct {
	Owner        common.Address
	Claimer      common.Address
	PubKeyClaim  [32]byte
	PubKeyRefund [32]byte
	Timeout0     *big.Int
	Timeout1     *big.Int
	IsReady      bool
	Completed    bool
	Value        *big.Int
}, error) {
	return _SwapFactory.Contract.Swaps(&_SwapFactory.CallOpts, arg0)
}

// Swaps is a free data retrieval call binding the contract method 0xf09c5829.
//
// Solidity: function swaps(uint256 ) view returns(address owner, address claimer, bytes32 pubKeyClaim, bytes32 pubKeyRefund, uint256 timeout_0, uint256 timeout_1, bool isReady, bool completed, uint256 value)
func (_SwapFactory *SwapFactoryCallerSession) Swaps(arg0 *big.Int) (struct {
	Owner        common.Address
	Claimer      common.Address
	PubKeyClaim  [32]byte
	PubKeyRefund [32]byte
	Timeout0     *big.Int
	Timeout1     *big.Int
	IsReady      bool
	Completed    bool
	Value        *big.Int
}, error) {
	return _SwapFactory.Contract.Swaps(&_SwapFactory.CallOpts, arg0)
}

// Claim is a paid mutator transaction binding the contract method 0x31d14457.
//
// Solidity: function claim(uint256 id, bytes32 _s) returns()
func (_SwapFactory *SwapFactoryTransactor) Claim(opts *bind.TransactOpts, id *big.Int, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "claim", id, _s)
}

// Claim is a paid mutator transaction binding the contract method 0x31d14457.
//
// Solidity: function claim(uint256 id, bytes32 _s) returns()
func (_SwapFactory *SwapFactorySession) Claim(id *big.Int, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.Contract.Claim(&_SwapFactory.TransactOpts, id, _s)
}

// Claim is a paid mutator transaction binding the contract method 0x31d14457.
//
// Solidity: function claim(uint256 id, bytes32 _s) returns()
func (_SwapFactory *SwapFactoryTransactorSession) Claim(id *big.Int, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.Contract.Claim(&_SwapFactory.TransactOpts, id, _s)
}

// NewSwap is a paid mutator transaction binding the contract method 0x0deeecba.
//
// Solidity: function new_swap(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, address _claimer, uint256 _timeoutDuration) payable returns(uint256)
func (_SwapFactory *SwapFactoryTransactor) NewSwap(opts *bind.TransactOpts, _pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration *big.Int) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "new_swap", _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration)
}

// NewSwap is a paid mutator transaction binding the contract method 0x0deeecba.
//
// Solidity: function new_swap(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, address _claimer, uint256 _timeoutDuration) payable returns(uint256)
func (_SwapFactory *SwapFactorySession) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration *big.Int) (*types.Transaction, error) {
	return _SwapFactory.Contract.NewSwap(&_SwapFactory.TransactOpts, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration)
}

// NewSwap is a paid mutator transaction binding the contract method 0x0deeecba.
//
// Solidity: function new_swap(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, address _claimer, uint256 _timeoutDuration) payable returns(uint256)
func (_SwapFactory *SwapFactoryTransactorSession) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration *big.Int) (*types.Transaction, error) {
	return _SwapFactory.Contract.NewSwap(&_SwapFactory.TransactOpts, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration)
}

// Refund is a paid mutator transaction binding the contract method 0x71eedb88.
//
// Solidity: function refund(uint256 id, bytes32 _s) returns()
func (_SwapFactory *SwapFactoryTransactor) Refund(opts *bind.TransactOpts, id *big.Int, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "refund", id, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x71eedb88.
//
// Solidity: function refund(uint256 id, bytes32 _s) returns()
func (_SwapFactory *SwapFactorySession) Refund(id *big.Int, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.Contract.Refund(&_SwapFactory.TransactOpts, id, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x71eedb88.
//
// Solidity: function refund(uint256 id, bytes32 _s) returns()
func (_SwapFactory *SwapFactoryTransactorSession) Refund(id *big.Int, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.Contract.Refund(&_SwapFactory.TransactOpts, id, _s)
}

// SetReady is a paid mutator transaction binding the contract method 0x37da2ecf.
//
// Solidity: function set_ready(uint256 id) returns()
func (_SwapFactory *SwapFactoryTransactor) SetReady(opts *bind.TransactOpts, id *big.Int) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "set_ready", id)
}

// SetReady is a paid mutator transaction binding the contract method 0x37da2ecf.
//
// Solidity: function set_ready(uint256 id) returns()
func (_SwapFactory *SwapFactorySession) SetReady(id *big.Int) (*types.Transaction, error) {
	return _SwapFactory.Contract.SetReady(&_SwapFactory.TransactOpts, id)
}

// SetReady is a paid mutator transaction binding the contract method 0x37da2ecf.
//
// Solidity: function set_ready(uint256 id) returns()
func (_SwapFactory *SwapFactoryTransactorSession) SetReady(id *big.Int) (*types.Transaction, error) {
	return _SwapFactory.Contract.SetReady(&_SwapFactory.TransactOpts, id)
}

// SwapFactoryClaimedIterator is returned from FilterClaimed and is used to iterate over the raw logs and unpacked data for Claimed events raised by the SwapFactory contract.
type SwapFactoryClaimedIterator struct {
	Event *SwapFactoryClaimed // Event containing the contract specifics and raw log

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
func (it *SwapFactoryClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapFactoryClaimed)
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
		it.Event = new(SwapFactoryClaimed)
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
func (it *SwapFactoryClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapFactoryClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapFactoryClaimed represents a Claimed event raised by the SwapFactory contract.
type SwapFactoryClaimed struct {
	SwapID *big.Int
	S      [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterClaimed is a free log retrieval operation binding the contract event 0xd5a2476fc450083bbb092dd3f4be92698ffdc2d213e6f1e730c7f44a52f1ccfc.
//
// Solidity: event Claimed(uint256 swapID, bytes32 s)
func (_SwapFactory *SwapFactoryFilterer) FilterClaimed(opts *bind.FilterOpts) (*SwapFactoryClaimedIterator, error) {

	logs, sub, err := _SwapFactory.contract.FilterLogs(opts, "Claimed")
	if err != nil {
		return nil, err
	}
	return &SwapFactoryClaimedIterator{contract: _SwapFactory.contract, event: "Claimed", logs: logs, sub: sub}, nil
}

// WatchClaimed is a free log subscription operation binding the contract event 0xd5a2476fc450083bbb092dd3f4be92698ffdc2d213e6f1e730c7f44a52f1ccfc.
//
// Solidity: event Claimed(uint256 swapID, bytes32 s)
func (_SwapFactory *SwapFactoryFilterer) WatchClaimed(opts *bind.WatchOpts, sink chan<- *SwapFactoryClaimed) (event.Subscription, error) {

	logs, sub, err := _SwapFactory.contract.WatchLogs(opts, "Claimed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapFactoryClaimed)
				if err := _SwapFactory.contract.UnpackLog(event, "Claimed", log); err != nil {
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

// ParseClaimed is a log parse operation binding the contract event 0xd5a2476fc450083bbb092dd3f4be92698ffdc2d213e6f1e730c7f44a52f1ccfc.
//
// Solidity: event Claimed(uint256 swapID, bytes32 s)
func (_SwapFactory *SwapFactoryFilterer) ParseClaimed(log types.Log) (*SwapFactoryClaimed, error) {
	event := new(SwapFactoryClaimed)
	if err := _SwapFactory.contract.UnpackLog(event, "Claimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapFactoryNewIterator is returned from FilterNew and is used to iterate over the raw logs and unpacked data for New events raised by the SwapFactory contract.
type SwapFactoryNewIterator struct {
	Event *SwapFactoryNew // Event containing the contract specifics and raw log

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
func (it *SwapFactoryNewIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapFactoryNew)
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
		it.Event = new(SwapFactoryNew)
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
func (it *SwapFactoryNewIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapFactoryNewIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapFactoryNew represents a New event raised by the SwapFactory contract.
type SwapFactoryNew struct {
	SwapID    *big.Int
	ClaimKey  [32]byte
	RefundKey [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterNew is a free log retrieval operation binding the contract event 0x982a99d883f17ecd5797205d5b3674205d7882bb28a9487d736d3799422cd055.
//
// Solidity: event New(uint256 swapID, bytes32 claimKey, bytes32 refundKey)
func (_SwapFactory *SwapFactoryFilterer) FilterNew(opts *bind.FilterOpts) (*SwapFactoryNewIterator, error) {

	logs, sub, err := _SwapFactory.contract.FilterLogs(opts, "New")
	if err != nil {
		return nil, err
	}
	return &SwapFactoryNewIterator{contract: _SwapFactory.contract, event: "New", logs: logs, sub: sub}, nil
}

// WatchNew is a free log subscription operation binding the contract event 0x982a99d883f17ecd5797205d5b3674205d7882bb28a9487d736d3799422cd055.
//
// Solidity: event New(uint256 swapID, bytes32 claimKey, bytes32 refundKey)
func (_SwapFactory *SwapFactoryFilterer) WatchNew(opts *bind.WatchOpts, sink chan<- *SwapFactoryNew) (event.Subscription, error) {

	logs, sub, err := _SwapFactory.contract.WatchLogs(opts, "New")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapFactoryNew)
				if err := _SwapFactory.contract.UnpackLog(event, "New", log); err != nil {
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

// ParseNew is a log parse operation binding the contract event 0x982a99d883f17ecd5797205d5b3674205d7882bb28a9487d736d3799422cd055.
//
// Solidity: event New(uint256 swapID, bytes32 claimKey, bytes32 refundKey)
func (_SwapFactory *SwapFactoryFilterer) ParseNew(log types.Log) (*SwapFactoryNew, error) {
	event := new(SwapFactoryNew)
	if err := _SwapFactory.contract.UnpackLog(event, "New", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapFactoryReadyIterator is returned from FilterReady and is used to iterate over the raw logs and unpacked data for Ready events raised by the SwapFactory contract.
type SwapFactoryReadyIterator struct {
	Event *SwapFactoryReady // Event containing the contract specifics and raw log

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
func (it *SwapFactoryReadyIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapFactoryReady)
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
		it.Event = new(SwapFactoryReady)
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
func (it *SwapFactoryReadyIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapFactoryReadyIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapFactoryReady represents a Ready event raised by the SwapFactory contract.
type SwapFactoryReady struct {
	SwapID *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterReady is a free log retrieval operation binding the contract event 0x0b217ad5c70346c7cd952bd2463c6684a56f9ed229f5780947586625781b4770.
//
// Solidity: event Ready(uint256 swapID)
func (_SwapFactory *SwapFactoryFilterer) FilterReady(opts *bind.FilterOpts) (*SwapFactoryReadyIterator, error) {

	logs, sub, err := _SwapFactory.contract.FilterLogs(opts, "Ready")
	if err != nil {
		return nil, err
	}
	return &SwapFactoryReadyIterator{contract: _SwapFactory.contract, event: "Ready", logs: logs, sub: sub}, nil
}

// WatchReady is a free log subscription operation binding the contract event 0x0b217ad5c70346c7cd952bd2463c6684a56f9ed229f5780947586625781b4770.
//
// Solidity: event Ready(uint256 swapID)
func (_SwapFactory *SwapFactoryFilterer) WatchReady(opts *bind.WatchOpts, sink chan<- *SwapFactoryReady) (event.Subscription, error) {

	logs, sub, err := _SwapFactory.contract.WatchLogs(opts, "Ready")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapFactoryReady)
				if err := _SwapFactory.contract.UnpackLog(event, "Ready", log); err != nil {
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

// ParseReady is a log parse operation binding the contract event 0x0b217ad5c70346c7cd952bd2463c6684a56f9ed229f5780947586625781b4770.
//
// Solidity: event Ready(uint256 swapID)
func (_SwapFactory *SwapFactoryFilterer) ParseReady(log types.Log) (*SwapFactoryReady, error) {
	event := new(SwapFactoryReady)
	if err := _SwapFactory.contract.UnpackLog(event, "Ready", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapFactoryRefundedIterator is returned from FilterRefunded and is used to iterate over the raw logs and unpacked data for Refunded events raised by the SwapFactory contract.
type SwapFactoryRefundedIterator struct {
	Event *SwapFactoryRefunded // Event containing the contract specifics and raw log

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
func (it *SwapFactoryRefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapFactoryRefunded)
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
		it.Event = new(SwapFactoryRefunded)
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
func (it *SwapFactoryRefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapFactoryRefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapFactoryRefunded represents a Refunded event raised by the SwapFactory contract.
type SwapFactoryRefunded struct {
	SwapID *big.Int
	S      [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterRefunded is a free log retrieval operation binding the contract event 0x4fd30f3ee0d64f7eaa62d0e005ca64c6a560652156d6c33f23ea8ca4936106e0.
//
// Solidity: event Refunded(uint256 swapID, bytes32 s)
func (_SwapFactory *SwapFactoryFilterer) FilterRefunded(opts *bind.FilterOpts) (*SwapFactoryRefundedIterator, error) {

	logs, sub, err := _SwapFactory.contract.FilterLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return &SwapFactoryRefundedIterator{contract: _SwapFactory.contract, event: "Refunded", logs: logs, sub: sub}, nil
}

// WatchRefunded is a free log subscription operation binding the contract event 0x4fd30f3ee0d64f7eaa62d0e005ca64c6a560652156d6c33f23ea8ca4936106e0.
//
// Solidity: event Refunded(uint256 swapID, bytes32 s)
func (_SwapFactory *SwapFactoryFilterer) WatchRefunded(opts *bind.WatchOpts, sink chan<- *SwapFactoryRefunded) (event.Subscription, error) {

	logs, sub, err := _SwapFactory.contract.WatchLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapFactoryRefunded)
				if err := _SwapFactory.contract.UnpackLog(event, "Refunded", log); err != nil {
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

// ParseRefunded is a log parse operation binding the contract event 0x4fd30f3ee0d64f7eaa62d0e005ca64c6a560652156d6c33f23ea8ca4936106e0.
//
// Solidity: event Refunded(uint256 swapID, bytes32 s)
func (_SwapFactory *SwapFactoryFilterer) ParseRefunded(log types.Log) (*SwapFactoryRefunded, error) {
	event := new(SwapFactoryRefunded)
	if err := _SwapFactory.contract.UnpackLog(event, "Refunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
