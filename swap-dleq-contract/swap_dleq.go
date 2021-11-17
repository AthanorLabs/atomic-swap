// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package swapdleq

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

// SwapDLEQMetaData contains all meta data concerning the SwapDLEQ contract.
var SwapDLEQMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"claimCtment\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"refundCtment\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"s\",\"type\":\"uint256\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"p\",\"type\":\"bytes32\"}],\"name\":\"Constructed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"IsReady\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"s\",\"type\":\"uint256\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_s\",\"type\":\"uint256\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimCtment\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isReady\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_s\",\"type\":\"uint256\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"refundCtment\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"set_ready\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeout_0\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeout_1\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x6101206040526000600160006101000a81548160ff0219169083151502179055506040516200128e3803806200128e833981810160405281019062000045919062000194565b81813373ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff16815250508160a081815250508060c0818152505062015180426200009c919062000214565b60e081815250507f1ba2159dcdf5aa313440d6540b9acb108e5f7907737e884db04579a584275fbb81604051620000d4919062000282565b60405180910390a15050604051620000ec9062000146565b604051809103906000f08015801562000109573d6000803e3d6000fd5b5073ffffffffffffffffffffffffffffffffffffffff166101008173ffffffffffffffffffffffffffffffffffffffff168152505050506200029f565b6103ca8062000ec483390190565b600080fd5b6000819050919050565b6200016e8162000159565b81146200017a57600080fd5b50565b6000815190506200018e8162000163565b92915050565b60008060408385031215620001ae57620001ad62000154565b5b6000620001be858286016200017d565b9250506020620001d1858286016200017d565b9150509250929050565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60006200022182620001db565b91506200022e83620001db565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03821115620002665762000265620001e5565b5b828201905092915050565b6200027c8162000159565b82525050565b600060208201905062000299600083018462000271565b92915050565b60805160a05160c05160e05161010051610bb66200030e600039600061057d0152600081816101ea015281816103690152818161044d01526104e001526000818161025101526102e701526000818161016701526103d10152600081816102ae01526104890152610bb66000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c806345bb8e091161005b57806345bb8e09146101015780634ded8d521461011f57806374d7c1381461013d578063a094a0311461014757610088565b80631fba1e8a1461008d578063278ecde1146100ab5780632dd0ef7c146100c7578063379607f5146100e5575b600080fd5b610095610165565b6040516100a29190610685565b60405180910390f35b6100c560048036038101906100c091906106db565b610189565b005b6100cf6102e5565b6040516100dc9190610685565b60405180910390f35b6100ff60048036038101906100fa91906106db565b610309565b005b610109610445565b6040516101169190610717565b60405180910390f35b61012761044b565b6040516101349190610717565b60405180910390f35b61014561046f565b005b61014f610568565b60405161015c919061074d565b60405180910390f35b7f000000000000000000000000000000000000000000000000000000000000000081565b600160009054906101000a900460ff16156101e8576000544210156101e3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101da90610811565b60405180910390fd5b61024b565b7f0000000000000000000000000000000000000000000000000000000000000000421061024a576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610241906108a3565b60405180910390fd5b5b610275817f000000000000000000000000000000000000000000000000000000000000000061057b565b7f3d2a04f53164bedf9a8a46353305d6b2d2261410406df3b41f99ce6489dc003c816040516102a49190610717565b60405180910390a17f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16ff5b7f000000000000000000000000000000000000000000000000000000000000000081565b600160009054906101000a900460ff1615610367576000544210610362576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161035990610935565b60405180910390fd5b6103cb565b7f00000000000000000000000000000000000000000000000000000000000000004210156103ca576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103c1906109ed565b60405180910390fd5b5b6103f5817f000000000000000000000000000000000000000000000000000000000000000061057b565b7f7a355715549cfe7c1cba26304350343fbddc4b4f72d3ce3e7c27117dd20b5cb8816040516104249190610717565b60405180910390a13373ffffffffffffffffffffffffffffffffffffffff16ff5b60005481565b7f000000000000000000000000000000000000000000000000000000000000000081565b600160009054906101000a900460ff161580156104d757507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b801561050257507f000000000000000000000000000000000000000000000000000000000000000042105b61050b57600080fd5b60018060006101000a81548160ff02191690831515021790555062015180426105349190610a3c565b6000819055507ff4b95676cf0a9825d44076281c2bb1c3a5c00c3b076e4793838cf071e930a92260405160405180910390a1565b600160009054906101000a900460ff1681565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663b32d1b4f838360001c6040518363ffffffff1660e01b81526004016105d9929190610a92565b60206040518083038186803b1580156105f157600080fd5b505afa158015610605573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106299190610ae7565b610668576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161065f90610b60565b60405180910390fd5b5050565b6000819050919050565b61067f8161066c565b82525050565b600060208201905061069a6000830184610676565b92915050565b600080fd5b6000819050919050565b6106b8816106a5565b81146106c357600080fd5b50565b6000813590506106d5816106af565b92915050565b6000602082840312156106f1576106f06106a0565b5b60006106ff848285016106c6565b91505092915050565b610711816106a5565b82525050565b600060208201905061072c6000830184610708565b92915050565b60008115159050919050565b61074781610732565b82525050565b6000602082019050610762600083018461073e565b92915050565b600082825260208201905092915050565b7f426f622063616e206e6f7720636c61696d207468652066756e647320756e746960008201527f6c20746865207365636f6e642074696d656f75742c20706c656173652077616960208201527f7421000000000000000000000000000000000000000000000000000000000000604082015250565b60006107fb604283610768565b915061080682610779565b606082019050919050565b6000602082019050818103600083015261082a816107ee565b9050919050565b7f546f6f206c61746520666f72206120726566756e64212050726179207468617460008201527f20426f6220636c61696d7320686973204554482e000000000000000000000000602082015250565b600061088d603483610768565b915061089882610831565b604082019050919050565b600060208201905081810360008301526108bc81610880565b9050919050565b7f546f6f206c61746520746f20636c61696d212050726179207468617420416c6960008201527f636520636c61696d73206120726566756e642e00000000000000000000000000602082015250565b600061091f603383610768565b915061092a826108c3565b604082019050919050565b6000602082019050818103600083015261094e81610912565b9050919050565b7f506c65617365207761697420756e74696c20416c696365206861732063616c6c60008201527f6564207365745f7265616479206f72207468652066697273742074696d656f7560208201527f7420697320726561636865642e00000000000000000000000000000000000000604082015250565b60006109d7604d83610768565b91506109e282610955565b606082019050919050565b60006020820190508181036000830152610a06816109ca565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610a47826106a5565b9150610a52836106a5565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03821115610a8757610a86610a0d565b5b828201905092915050565b6000604082019050610aa76000830185610708565b610ab46020830184610708565b9392505050565b610ac481610732565b8114610acf57600080fd5b50565b600081519050610ae181610abb565b92915050565b600060208284031215610afd57610afc6106a0565b5b6000610b0b84828501610ad2565b91505092915050565b7f77726f6e67207365637265740000000000000000000000000000000000000000600082015250565b6000610b4a600c83610768565b9150610b5582610b14565b602082019050919050565b60006020820190508181036000830152610b7981610b3d565b905091905056fea2646970667358221220a8154369f15210a2059835baaa32c1506ad0b6e4003c27c9f011b6d25274829864736f6c63430008090033608060405234801561001057600080fd5b506103aa806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c8063b32d1b4f14610030575b600080fd5b61004a600480360381019061004591906101dc565b610060565b6040516100579190610237565b60405180910390f35b600080600160008060027f483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b86100959190610281565b14156100a257601b6100a5565b601c5b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179860001b7ffffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141806100f8576100f7610252565b5b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798890960001b6040516000815260200160405260405161013b949392919061032f565b6020604051602081039080840390855afa15801561015d573d6000803e3d6000fd5b5050506020604051035190508073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161491505092915050565b600080fd5b6000819050919050565b6101b9816101a6565b81146101c457600080fd5b50565b6000813590506101d6816101b0565b92915050565b600080604083850312156101f3576101f26101a1565b5b6000610201858286016101c7565b9250506020610212858286016101c7565b9150509250929050565b60008115159050919050565b6102318161021c565b82525050565b600060208201905061024c6000830184610228565b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b600061028c826101a6565b9150610297836101a6565b9250826102a7576102a6610252565b5b828206905092915050565b6000819050919050565b6000819050919050565b60008160001b9050919050565b60006102ee6102e96102e4846102b2565b6102c6565b6102bc565b9050919050565b6102fe816102d3565b82525050565b600060ff82169050919050565b61031a81610304565b82525050565b610329816102bc565b82525050565b600060808201905061034460008301876102f5565b6103516020830186610311565b61035e6040830185610320565b61036b6060830184610320565b9594505050505056fea2646970667358221220ea65f48a5c6f3bc62c65a64985d7720e0fe954fe084b0c2965f8310cb9b2e23d64736f6c63430008090033",
}

// SwapDLEQABI is the input ABI used to generate the binding from.
// Deprecated: Use SwapDLEQMetaData.ABI instead.
var SwapDLEQABI = SwapDLEQMetaData.ABI

// SwapDLEQBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SwapDLEQMetaData.Bin instead.
var SwapDLEQBin = SwapDLEQMetaData.Bin

// DeploySwapDLEQ deploys a new Ethereum contract, binding an instance of SwapDLEQ to it.
func DeploySwapDLEQ(auth *bind.TransactOpts, backend bind.ContractBackend, claimCtment [32]byte, refundCtment [32]byte) (common.Address, *types.Transaction, *SwapDLEQ, error) {
	parsed, err := SwapDLEQMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SwapDLEQBin), backend, claimCtment, refundCtment)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SwapDLEQ{SwapDLEQCaller: SwapDLEQCaller{contract: contract}, SwapDLEQTransactor: SwapDLEQTransactor{contract: contract}, SwapDLEQFilterer: SwapDLEQFilterer{contract: contract}}, nil
}

// SwapDLEQ is an auto generated Go binding around an Ethereum contract.
type SwapDLEQ struct {
	SwapDLEQCaller     // Read-only binding to the contract
	SwapDLEQTransactor // Write-only binding to the contract
	SwapDLEQFilterer   // Log filterer for contract events
}

// SwapDLEQCaller is an auto generated read-only Go binding around an Ethereum contract.
type SwapDLEQCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapDLEQTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SwapDLEQTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapDLEQFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SwapDLEQFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapDLEQSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SwapDLEQSession struct {
	Contract     *SwapDLEQ         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SwapDLEQCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SwapDLEQCallerSession struct {
	Contract *SwapDLEQCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// SwapDLEQTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SwapDLEQTransactorSession struct {
	Contract     *SwapDLEQTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// SwapDLEQRaw is an auto generated low-level Go binding around an Ethereum contract.
type SwapDLEQRaw struct {
	Contract *SwapDLEQ // Generic contract binding to access the raw methods on
}

// SwapDLEQCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SwapDLEQCallerRaw struct {
	Contract *SwapDLEQCaller // Generic read-only contract binding to access the raw methods on
}

// SwapDLEQTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SwapDLEQTransactorRaw struct {
	Contract *SwapDLEQTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSwapDLEQ creates a new instance of SwapDLEQ, bound to a specific deployed contract.
func NewSwapDLEQ(address common.Address, backend bind.ContractBackend) (*SwapDLEQ, error) {
	contract, err := bindSwapDLEQ(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SwapDLEQ{SwapDLEQCaller: SwapDLEQCaller{contract: contract}, SwapDLEQTransactor: SwapDLEQTransactor{contract: contract}, SwapDLEQFilterer: SwapDLEQFilterer{contract: contract}}, nil
}

// NewSwapDLEQCaller creates a new read-only instance of SwapDLEQ, bound to a specific deployed contract.
func NewSwapDLEQCaller(address common.Address, caller bind.ContractCaller) (*SwapDLEQCaller, error) {
	contract, err := bindSwapDLEQ(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SwapDLEQCaller{contract: contract}, nil
}

// NewSwapDLEQTransactor creates a new write-only instance of SwapDLEQ, bound to a specific deployed contract.
func NewSwapDLEQTransactor(address common.Address, transactor bind.ContractTransactor) (*SwapDLEQTransactor, error) {
	contract, err := bindSwapDLEQ(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SwapDLEQTransactor{contract: contract}, nil
}

// NewSwapDLEQFilterer creates a new log filterer instance of SwapDLEQ, bound to a specific deployed contract.
func NewSwapDLEQFilterer(address common.Address, filterer bind.ContractFilterer) (*SwapDLEQFilterer, error) {
	contract, err := bindSwapDLEQ(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SwapDLEQFilterer{contract: contract}, nil
}

// bindSwapDLEQ binds a generic wrapper to an already deployed contract.
func bindSwapDLEQ(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SwapDLEQABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SwapDLEQ *SwapDLEQRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SwapDLEQ.Contract.SwapDLEQCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SwapDLEQ *SwapDLEQRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.SwapDLEQTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SwapDLEQ *SwapDLEQRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.SwapDLEQTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SwapDLEQ *SwapDLEQCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SwapDLEQ.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SwapDLEQ *SwapDLEQTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SwapDLEQ *SwapDLEQTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.contract.Transact(opts, method, params...)
}

// ClaimCtment is a free data retrieval call binding the contract method 0x1fba1e8a.
//
// Solidity: function claimCtment() view returns(bytes32)
func (_SwapDLEQ *SwapDLEQCaller) ClaimCtment(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SwapDLEQ.contract.Call(opts, &out, "claimCtment")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ClaimCtment is a free data retrieval call binding the contract method 0x1fba1e8a.
//
// Solidity: function claimCtment() view returns(bytes32)
func (_SwapDLEQ *SwapDLEQSession) ClaimCtment() ([32]byte, error) {
	return _SwapDLEQ.Contract.ClaimCtment(&_SwapDLEQ.CallOpts)
}

// ClaimCtment is a free data retrieval call binding the contract method 0x1fba1e8a.
//
// Solidity: function claimCtment() view returns(bytes32)
func (_SwapDLEQ *SwapDLEQCallerSession) ClaimCtment() ([32]byte, error) {
	return _SwapDLEQ.Contract.ClaimCtment(&_SwapDLEQ.CallOpts)
}

// IsReady is a free data retrieval call binding the contract method 0xa094a031.
//
// Solidity: function isReady() view returns(bool)
func (_SwapDLEQ *SwapDLEQCaller) IsReady(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _SwapDLEQ.contract.Call(opts, &out, "isReady")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsReady is a free data retrieval call binding the contract method 0xa094a031.
//
// Solidity: function isReady() view returns(bool)
func (_SwapDLEQ *SwapDLEQSession) IsReady() (bool, error) {
	return _SwapDLEQ.Contract.IsReady(&_SwapDLEQ.CallOpts)
}

// IsReady is a free data retrieval call binding the contract method 0xa094a031.
//
// Solidity: function isReady() view returns(bool)
func (_SwapDLEQ *SwapDLEQCallerSession) IsReady() (bool, error) {
	return _SwapDLEQ.Contract.IsReady(&_SwapDLEQ.CallOpts)
}

// RefundCtment is a free data retrieval call binding the contract method 0x2dd0ef7c.
//
// Solidity: function refundCtment() view returns(bytes32)
func (_SwapDLEQ *SwapDLEQCaller) RefundCtment(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SwapDLEQ.contract.Call(opts, &out, "refundCtment")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RefundCtment is a free data retrieval call binding the contract method 0x2dd0ef7c.
//
// Solidity: function refundCtment() view returns(bytes32)
func (_SwapDLEQ *SwapDLEQSession) RefundCtment() ([32]byte, error) {
	return _SwapDLEQ.Contract.RefundCtment(&_SwapDLEQ.CallOpts)
}

// RefundCtment is a free data retrieval call binding the contract method 0x2dd0ef7c.
//
// Solidity: function refundCtment() view returns(bytes32)
func (_SwapDLEQ *SwapDLEQCallerSession) RefundCtment() ([32]byte, error) {
	return _SwapDLEQ.Contract.RefundCtment(&_SwapDLEQ.CallOpts)
}

// Timeout0 is a free data retrieval call binding the contract method 0x4ded8d52.
//
// Solidity: function timeout_0() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCaller) Timeout0(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwapDLEQ.contract.Call(opts, &out, "timeout_0")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Timeout0 is a free data retrieval call binding the contract method 0x4ded8d52.
//
// Solidity: function timeout_0() view returns(uint256)
func (_SwapDLEQ *SwapDLEQSession) Timeout0() (*big.Int, error) {
	return _SwapDLEQ.Contract.Timeout0(&_SwapDLEQ.CallOpts)
}

// Timeout0 is a free data retrieval call binding the contract method 0x4ded8d52.
//
// Solidity: function timeout_0() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCallerSession) Timeout0() (*big.Int, error) {
	return _SwapDLEQ.Contract.Timeout0(&_SwapDLEQ.CallOpts)
}

// Timeout1 is a free data retrieval call binding the contract method 0x45bb8e09.
//
// Solidity: function timeout_1() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCaller) Timeout1(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwapDLEQ.contract.Call(opts, &out, "timeout_1")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Timeout1 is a free data retrieval call binding the contract method 0x45bb8e09.
//
// Solidity: function timeout_1() view returns(uint256)
func (_SwapDLEQ *SwapDLEQSession) Timeout1() (*big.Int, error) {
	return _SwapDLEQ.Contract.Timeout1(&_SwapDLEQ.CallOpts)
}

// Timeout1 is a free data retrieval call binding the contract method 0x45bb8e09.
//
// Solidity: function timeout_1() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCallerSession) Timeout1() (*big.Int, error) {
	return _SwapDLEQ.Contract.Timeout1(&_SwapDLEQ.CallOpts)
}

// Claim is a paid mutator transaction binding the contract method 0x379607f5.
//
// Solidity: function claim(uint256 _s) returns()
func (_SwapDLEQ *SwapDLEQTransactor) Claim(opts *bind.TransactOpts, _s *big.Int) (*types.Transaction, error) {
	return _SwapDLEQ.contract.Transact(opts, "claim", _s)
}

// Claim is a paid mutator transaction binding the contract method 0x379607f5.
//
// Solidity: function claim(uint256 _s) returns()
func (_SwapDLEQ *SwapDLEQSession) Claim(_s *big.Int) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.Claim(&_SwapDLEQ.TransactOpts, _s)
}

// Claim is a paid mutator transaction binding the contract method 0x379607f5.
//
// Solidity: function claim(uint256 _s) returns()
func (_SwapDLEQ *SwapDLEQTransactorSession) Claim(_s *big.Int) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.Claim(&_SwapDLEQ.TransactOpts, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 _s) returns()
func (_SwapDLEQ *SwapDLEQTransactor) Refund(opts *bind.TransactOpts, _s *big.Int) (*types.Transaction, error) {
	return _SwapDLEQ.contract.Transact(opts, "refund", _s)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 _s) returns()
func (_SwapDLEQ *SwapDLEQSession) Refund(_s *big.Int) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.Refund(&_SwapDLEQ.TransactOpts, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 _s) returns()
func (_SwapDLEQ *SwapDLEQTransactorSession) Refund(_s *big.Int) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.Refund(&_SwapDLEQ.TransactOpts, _s)
}

// SetReady is a paid mutator transaction binding the contract method 0x74d7c138.
//
// Solidity: function set_ready() returns()
func (_SwapDLEQ *SwapDLEQTransactor) SetReady(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwapDLEQ.contract.Transact(opts, "set_ready")
}

// SetReady is a paid mutator transaction binding the contract method 0x74d7c138.
//
// Solidity: function set_ready() returns()
func (_SwapDLEQ *SwapDLEQSession) SetReady() (*types.Transaction, error) {
	return _SwapDLEQ.Contract.SetReady(&_SwapDLEQ.TransactOpts)
}

// SetReady is a paid mutator transaction binding the contract method 0x74d7c138.
//
// Solidity: function set_ready() returns()
func (_SwapDLEQ *SwapDLEQTransactorSession) SetReady() (*types.Transaction, error) {
	return _SwapDLEQ.Contract.SetReady(&_SwapDLEQ.TransactOpts)
}

// SwapDLEQClaimedIterator is returned from FilterClaimed and is used to iterate over the raw logs and unpacked data for Claimed events raised by the SwapDLEQ contract.
type SwapDLEQClaimedIterator struct {
	Event *SwapDLEQClaimed // Event containing the contract specifics and raw log

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
func (it *SwapDLEQClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapDLEQClaimed)
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
		it.Event = new(SwapDLEQClaimed)
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
func (it *SwapDLEQClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapDLEQClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapDLEQClaimed represents a Claimed event raised by the SwapDLEQ contract.
type SwapDLEQClaimed struct {
	S   *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterClaimed is a free log retrieval operation binding the contract event 0x7a355715549cfe7c1cba26304350343fbddc4b4f72d3ce3e7c27117dd20b5cb8.
//
// Solidity: event Claimed(uint256 s)
func (_SwapDLEQ *SwapDLEQFilterer) FilterClaimed(opts *bind.FilterOpts) (*SwapDLEQClaimedIterator, error) {

	logs, sub, err := _SwapDLEQ.contract.FilterLogs(opts, "Claimed")
	if err != nil {
		return nil, err
	}
	return &SwapDLEQClaimedIterator{contract: _SwapDLEQ.contract, event: "Claimed", logs: logs, sub: sub}, nil
}

// WatchClaimed is a free log subscription operation binding the contract event 0x7a355715549cfe7c1cba26304350343fbddc4b4f72d3ce3e7c27117dd20b5cb8.
//
// Solidity: event Claimed(uint256 s)
func (_SwapDLEQ *SwapDLEQFilterer) WatchClaimed(opts *bind.WatchOpts, sink chan<- *SwapDLEQClaimed) (event.Subscription, error) {

	logs, sub, err := _SwapDLEQ.contract.WatchLogs(opts, "Claimed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapDLEQClaimed)
				if err := _SwapDLEQ.contract.UnpackLog(event, "Claimed", log); err != nil {
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

// ParseClaimed is a log parse operation binding the contract event 0x7a355715549cfe7c1cba26304350343fbddc4b4f72d3ce3e7c27117dd20b5cb8.
//
// Solidity: event Claimed(uint256 s)
func (_SwapDLEQ *SwapDLEQFilterer) ParseClaimed(log types.Log) (*SwapDLEQClaimed, error) {
	event := new(SwapDLEQClaimed)
	if err := _SwapDLEQ.contract.UnpackLog(event, "Claimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapDLEQConstructedIterator is returned from FilterConstructed and is used to iterate over the raw logs and unpacked data for Constructed events raised by the SwapDLEQ contract.
type SwapDLEQConstructedIterator struct {
	Event *SwapDLEQConstructed // Event containing the contract specifics and raw log

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
func (it *SwapDLEQConstructedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapDLEQConstructed)
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
		it.Event = new(SwapDLEQConstructed)
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
func (it *SwapDLEQConstructedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapDLEQConstructedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapDLEQConstructed represents a Constructed event raised by the SwapDLEQ contract.
type SwapDLEQConstructed struct {
	P   [32]byte
	Raw types.Log // Blockchain specific contextual infos
}

// FilterConstructed is a free log retrieval operation binding the contract event 0x1ba2159dcdf5aa313440d6540b9acb108e5f7907737e884db04579a584275fbb.
//
// Solidity: event Constructed(bytes32 p)
func (_SwapDLEQ *SwapDLEQFilterer) FilterConstructed(opts *bind.FilterOpts) (*SwapDLEQConstructedIterator, error) {

	logs, sub, err := _SwapDLEQ.contract.FilterLogs(opts, "Constructed")
	if err != nil {
		return nil, err
	}
	return &SwapDLEQConstructedIterator{contract: _SwapDLEQ.contract, event: "Constructed", logs: logs, sub: sub}, nil
}

// WatchConstructed is a free log subscription operation binding the contract event 0x1ba2159dcdf5aa313440d6540b9acb108e5f7907737e884db04579a584275fbb.
//
// Solidity: event Constructed(bytes32 p)
func (_SwapDLEQ *SwapDLEQFilterer) WatchConstructed(opts *bind.WatchOpts, sink chan<- *SwapDLEQConstructed) (event.Subscription, error) {

	logs, sub, err := _SwapDLEQ.contract.WatchLogs(opts, "Constructed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapDLEQConstructed)
				if err := _SwapDLEQ.contract.UnpackLog(event, "Constructed", log); err != nil {
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

// ParseConstructed is a log parse operation binding the contract event 0x1ba2159dcdf5aa313440d6540b9acb108e5f7907737e884db04579a584275fbb.
//
// Solidity: event Constructed(bytes32 p)
func (_SwapDLEQ *SwapDLEQFilterer) ParseConstructed(log types.Log) (*SwapDLEQConstructed, error) {
	event := new(SwapDLEQConstructed)
	if err := _SwapDLEQ.contract.UnpackLog(event, "Constructed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapDLEQIsReadyIterator is returned from FilterIsReady and is used to iterate over the raw logs and unpacked data for IsReady events raised by the SwapDLEQ contract.
type SwapDLEQIsReadyIterator struct {
	Event *SwapDLEQIsReady // Event containing the contract specifics and raw log

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
func (it *SwapDLEQIsReadyIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapDLEQIsReady)
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
		it.Event = new(SwapDLEQIsReady)
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
func (it *SwapDLEQIsReadyIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapDLEQIsReadyIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapDLEQIsReady represents a IsReady event raised by the SwapDLEQ contract.
type SwapDLEQIsReady struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterIsReady is a free log retrieval operation binding the contract event 0xf4b95676cf0a9825d44076281c2bb1c3a5c00c3b076e4793838cf071e930a922.
//
// Solidity: event IsReady()
func (_SwapDLEQ *SwapDLEQFilterer) FilterIsReady(opts *bind.FilterOpts) (*SwapDLEQIsReadyIterator, error) {

	logs, sub, err := _SwapDLEQ.contract.FilterLogs(opts, "IsReady")
	if err != nil {
		return nil, err
	}
	return &SwapDLEQIsReadyIterator{contract: _SwapDLEQ.contract, event: "IsReady", logs: logs, sub: sub}, nil
}

// WatchIsReady is a free log subscription operation binding the contract event 0xf4b95676cf0a9825d44076281c2bb1c3a5c00c3b076e4793838cf071e930a922.
//
// Solidity: event IsReady()
func (_SwapDLEQ *SwapDLEQFilterer) WatchIsReady(opts *bind.WatchOpts, sink chan<- *SwapDLEQIsReady) (event.Subscription, error) {

	logs, sub, err := _SwapDLEQ.contract.WatchLogs(opts, "IsReady")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapDLEQIsReady)
				if err := _SwapDLEQ.contract.UnpackLog(event, "IsReady", log); err != nil {
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

// ParseIsReady is a log parse operation binding the contract event 0xf4b95676cf0a9825d44076281c2bb1c3a5c00c3b076e4793838cf071e930a922.
//
// Solidity: event IsReady()
func (_SwapDLEQ *SwapDLEQFilterer) ParseIsReady(log types.Log) (*SwapDLEQIsReady, error) {
	event := new(SwapDLEQIsReady)
	if err := _SwapDLEQ.contract.UnpackLog(event, "IsReady", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapDLEQRefundedIterator is returned from FilterRefunded and is used to iterate over the raw logs and unpacked data for Refunded events raised by the SwapDLEQ contract.
type SwapDLEQRefundedIterator struct {
	Event *SwapDLEQRefunded // Event containing the contract specifics and raw log

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
func (it *SwapDLEQRefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapDLEQRefunded)
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
		it.Event = new(SwapDLEQRefunded)
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
func (it *SwapDLEQRefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapDLEQRefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapDLEQRefunded represents a Refunded event raised by the SwapDLEQ contract.
type SwapDLEQRefunded struct {
	S   *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterRefunded is a free log retrieval operation binding the contract event 0x3d2a04f53164bedf9a8a46353305d6b2d2261410406df3b41f99ce6489dc003c.
//
// Solidity: event Refunded(uint256 s)
func (_SwapDLEQ *SwapDLEQFilterer) FilterRefunded(opts *bind.FilterOpts) (*SwapDLEQRefundedIterator, error) {

	logs, sub, err := _SwapDLEQ.contract.FilterLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return &SwapDLEQRefundedIterator{contract: _SwapDLEQ.contract, event: "Refunded", logs: logs, sub: sub}, nil
}

// WatchRefunded is a free log subscription operation binding the contract event 0x3d2a04f53164bedf9a8a46353305d6b2d2261410406df3b41f99ce6489dc003c.
//
// Solidity: event Refunded(uint256 s)
func (_SwapDLEQ *SwapDLEQFilterer) WatchRefunded(opts *bind.WatchOpts, sink chan<- *SwapDLEQRefunded) (event.Subscription, error) {

	logs, sub, err := _SwapDLEQ.contract.WatchLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapDLEQRefunded)
				if err := _SwapDLEQ.contract.UnpackLog(event, "Refunded", log); err != nil {
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

// ParseRefunded is a log parse operation binding the contract event 0x3d2a04f53164bedf9a8a46353305d6b2d2261410406df3b41f99ce6489dc003c.
//
// Solidity: event Refunded(uint256 s)
func (_SwapDLEQ *SwapDLEQFilterer) ParseRefunded(log types.Log) (*SwapDLEQRefunded, error) {
	event := new(SwapDLEQRefunded)
	if err := _SwapDLEQ.contract.UnpackLog(event, "Refunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
