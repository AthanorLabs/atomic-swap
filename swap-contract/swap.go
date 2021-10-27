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
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"addresspayable\",\"name\":\"_counterparty\",\"type\":\"address\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"s\",\"type\":\"uint256\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"p\",\"type\":\"bytes32\"}],\"name\":\"Constructed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"b\",\"type\":\"bool\"}],\"name\":\"IsReady\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"s\",\"type\":\"uint256\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_s\",\"type\":\"uint256\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pubKeyClaim\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pubKeyRefund\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_s\",\"type\":\"uint256\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"set_ready\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeout_0\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeout_1\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x6101406040526000600160006101000a81548160ff02191690831515021790555060405162001d5238038062001d5283398181016040528101906200004591906200022b565b3373ffffffffffffffffffffffffffffffffffffffff1660a08173ffffffffffffffffffffffffffffffffffffffff16815250508073ffffffffffffffffffffffffffffffffffffffff1660c08173ffffffffffffffffffffffffffffffffffffffff16815250508260e081815250508161010081815250506201518042620000cf9190620002c0565b6101208181525050604051620000e59062000178565b604051809103906000f08015801562000102573d6000803e3d6000fd5b5073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff16815250507f1ba2159dcdf5aa313440d6540b9acb108e5f7907737e884db04579a584275fbb826040516200016791906200032e565b60405180910390a15050506200034b565b610df78062000f5b83390190565b600080fd5b6000819050919050565b620001a0816200018b565b8114620001ac57600080fd5b50565b600081519050620001c08162000195565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000620001f382620001c6565b9050919050565b6200020581620001e6565b81146200021157600080fd5b50565b6000815190506200022581620001fa565b92915050565b60008060006060848603121562000247576200024662000186565b5b60006200025786828701620001af565b93505060206200026a86828701620001af565b92505060406200027d8682870162000214565b9150509250925092565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000620002cd8262000287565b9150620002da8362000287565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0382111562000312576200031162000291565b5b828201905092915050565b62000328816200018b565b82525050565b60006020820190506200034560008301846200031d565b92915050565b60805160a05160c05160e0516101005161012051610b96620003c5600039600081816101c80152818161032a0152818161042e01526104e501526000818161013e015261022f0152600081816103920152610452015260006103ef01526000818161028c015261048e0152600061057e0152610b966000f3fe608060405234801561001057600080fd5b506004361061007d5760003560e01c806345bb8e091161005b57806345bb8e09146100d85780634ded8d52146100f6578063736290f81461011457806374d7c138146101325761007d565b806303f7e24614610082578063278ecde1146100a0578063379607f5146100bc575b600080fd5b61008a61013c565b60405161009791906106a4565b60405180910390f35b6100ba60048036038101906100b591906106fa565b610160565b005b6100d660048036038101906100d191906106fa565b6102c3565b005b6100e0610426565b6040516100ed9190610736565b60405180910390f35b6100fe61042c565b60405161010b9190610736565b60405180910390f35b61011c610450565b60405161012991906106a4565b60405180910390f35b61013a610474565b005b7f000000000000000000000000000000000000000000000000000000000000000081565b60011515600160009054906101000a900460ff16151514156101c6576000544210156101c1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101b8906107d4565b60405180910390fd5b610229565b7f00000000000000000000000000000000000000000000000000000000000000004210610228576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161021f90610840565b60405180910390fd5b5b610253817f0000000000000000000000000000000000000000000000000000000000000000610579565b7f3d2a04f53164bedf9a8a46353305d6b2d2261410406df3b41f99ce6489dc003c816040516102829190610736565b60405180910390a17f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16ff5b60011515600160009054906101000a900460ff1615151415610328576000544210610323576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161031a906108ac565b60405180910390fd5b61038c565b7f000000000000000000000000000000000000000000000000000000000000000042101561038b576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103829061093e565b60405180910390fd5b5b6103b6817f0000000000000000000000000000000000000000000000000000000000000000610579565b7f7a355715549cfe7c1cba26304350343fbddc4b4f72d3ce3e7c27117dd20b5cb8816040516103e59190610736565b60405180910390a17f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16ff5b60005481565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b600160009054906101000a900460ff161580156104dc57507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b801561050757507f000000000000000000000000000000000000000000000000000000000000000042105b61051057600080fd5b60018060006101000a81548160ff0219169083151502179055506201518042610539919061098d565b6000819055507f2724cf6c3ad6a3399ad72482e4013d0171794f3ef4c462b7e24790c658cb3cd4600160405161056f91906109fe565b60405180910390a1565b6000807f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663c4f4912b856040518263ffffffff1660e01b81526004016105d59190610736565b604080518083038186803b1580156105ec57600080fd5b505afa158015610600573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106249190610a2e565b91509150600060ff6002846106399190610a9d565b901b82179050838160001b14610684576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161067b90610b40565b60405180910390fd5b5050505050565b6000819050919050565b61069e8161068b565b82525050565b60006020820190506106b96000830184610695565b92915050565b600080fd5b6000819050919050565b6106d7816106c4565b81146106e257600080fd5b50565b6000813590506106f4816106ce565b92915050565b6000602082840312156107105761070f6106bf565b5b600061071e848285016106e5565b91505092915050565b610730816106c4565b82525050565b600060208201905061074b6000830184610727565b92915050565b600082825260208201905092915050565b7f4974277320426f622773207475726e206e6f772c20706c65617365207761697460008201527f2100000000000000000000000000000000000000000000000000000000000000602082015250565b60006107be602183610751565b91506107c982610762565b604082019050919050565b600060208201905081810360008301526107ed816107b1565b9050919050565b7f4d697373656420796f7572206368616e63652100000000000000000000000000600082015250565b600061082a601383610751565b9150610835826107f4565b602082019050919050565b600060208201905081810360008301526108598161081d565b9050919050565b7f546f6f206c61746520746f20636c61696d210000000000000000000000000000600082015250565b6000610896601283610751565b91506108a182610860565b602082019050919050565b600060208201905081810360008301526108c581610889565b9050919050565b7f2769735265616479203d3d2066616c7365272063616e6e6f7420636c61696d2060008201527f7965742100000000000000000000000000000000000000000000000000000000602082015250565b6000610928602483610751565b9150610933826108cc565b604082019050919050565b600060208201905081810360008301526109578161091b565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610998826106c4565b91506109a3836106c4565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156109d8576109d761095e565b5b828201905092915050565b60008115159050919050565b6109f8816109e3565b82525050565b6000602082019050610a1360008301846109ef565b92915050565b600081519050610a28816106ce565b92915050565b60008060408385031215610a4557610a446106bf565b5b6000610a5385828601610a19565b9250506020610a6485828601610a19565b9150509250929050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000610aa8826106c4565b9150610ab3836106c4565b925082610ac357610ac2610a6e565b5b828206905092915050565b7f70726f76696465642073656372657420646f6573206e6f74206d61746368207460008201527f6865206578706563746564207075624b65790000000000000000000000000000602082015250565b6000610b2a603283610751565b9150610b3582610ace565b604082019050919050565b60006020820190508181036000830152610b5981610b1d565b905091905056fea26469706673582212203070b2d7301342842f5f7c7becb703a3f7b72cf73df2d62c182adaba2c88dfcc64736f6c63430008090033608060405234801561001057600080fd5b50610dd7806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c8063c4f4912b14610030575b600080fd5b61004a60048036038101906100459190610caa565b610061565b604051610058929190610ce6565b60405180910390f35b60008061006c610c09565b610074610c09565b7f216936d3cd6e53fec0a4e231fdd6dc5c692cc7609525a7b2c9562d608f25d51a8260000181815250507f666666666666666666666666666666666666666666666666666666666666665882602001818152505060018260400181815250506000816000018181525050600181602001818152505060018160400181815250505b600085111561012d57600180861614156101165761011381836101d2565b90505b600185901c945061012682610701565b91506100f5565b600061013c8260400151610b6b565b90507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061016d5761016c610d0f565b5b818360000151098260000181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806101ac576101ab610d0f565b5b818360200151098260200181815250508160000151826020015194509450505050915091565b6101da610c09565b6101e2610c2a565b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061021157610210610d0f565b5b83604001518560400151098160000181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061025457610253610d0f565b5b81600001518260000151098160200181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061029757610296610d0f565b5b83600001518560000151098160400181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806102da576102d9610d0f565b5b83602001518560200151098160600181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061031d5761031c610d0f565b5b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061034c5761034b610d0f565b5b82606001518360400151097f52036cee2b6ffe738cc740797779e89800700a4d4141d8ab75eb4dca135978a3098160800181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806103b1576103b0610d0f565b5b81608001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed6103e19190610d6d565b8260200151088160a00181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061041f5761041e610d0f565b5b81608001518260200151088160c00181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061046257610461610d0f565b5b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061049157610490610d0f565b5b82606001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed6104c19190610d6d565b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806104f0576104ef610d0f565b5b84604001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed6105209190610d6d565b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061054f5761054e610d0f565b5b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061057e5761057d610d0f565b5b89602001518a60000151087f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806105b8576105b7610d0f565b5b8b602001518c60000151080908087f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806105f5576105f4610d0f565b5b8360a00151846000015109098260000181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061063957610638610d0f565b5b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061066857610667610d0f565b5b82604001518360600151087f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806106a2576106a1610d0f565b5b8360c00151846000015109098260200181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806106e6576106e5610d0f565b5b8160c001518260a00151098260400181815250505092915050565b610709610c09565b610711610c2a565b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806107405761073f610d0f565b5b83602001518460000151088160000181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061078357610782610d0f565b5b81600001518260000151098160200181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806107c6576107c5610d0f565b5b83600001518460000151098160400181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061080957610808610d0f565b5b836020015184602001510981606001818152505080604001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed61084d9190610d6d565b8160800181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061088557610884610d0f565b5b81606001518260800151088160a00181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806108c8576108c7610d0f565b5b83604001518460400151098160e00181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061090b5761090a610d0f565b5b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061093a57610939610d0f565b5b8260e001516002097f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed61096d9190610d6d565b8260a00151088160c00181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806109ab576109aa610d0f565b5b8160c001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806109df576109de610d0f565b5b83606001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed610a0f9190610d6d565b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed80610a3e57610a3d610d0f565b5b85604001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed610a6e9190610d6d565b86602001510808098260000181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed80610aae57610aad610d0f565b5b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed80610add57610adc610d0f565b5b82606001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed610b0d9190610d6d565b8360800151088260a00151098260200181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed80610b5157610b50610d0f565b5b8160c001518260a001510982604001818152505050919050565b60008060027f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed610b9b9190610d6d565b905060007f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed905060405160208152602080820152602060408201528460608201528260808201528160a082015260208160c0836005600019fa610bfd57600080fd5b80519350505050919050565b60405180606001604052806000815260200160008152602001600081525090565b60405180610100016040528060008152602001600081526020016000815260200160008152602001600081526020016000815260200160008152602001600081525090565b600080fd5b6000819050919050565b610c8781610c74565b8114610c9257600080fd5b50565b600081359050610ca481610c7e565b92915050565b600060208284031215610cc057610cbf610c6f565b5b6000610cce84828501610c95565b91505092915050565b610ce081610c74565b82525050565b6000604082019050610cfb6000830185610cd7565b610d086020830184610cd7565b9392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610d7882610c74565b9150610d8383610c74565b925082821015610d9657610d95610d3e565b5b82820390509291505056fea264697066735822122026763aba14e28dd6f9cbc211a6db8bcee0af925f8e7aef10e72e4164c3feaf8c64736f6c63430008090033",
}

// SwapABI is the input ABI used to generate the binding from.
// Deprecated: Use SwapMetaData.ABI instead.
var SwapABI = SwapMetaData.ABI

// SwapBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SwapMetaData.Bin instead.
var SwapBin = SwapMetaData.Bin

// DeploySwap deploys a new Ethereum contract, binding an instance of Swap to it.
func DeploySwap(auth *bind.TransactOpts, backend bind.ContractBackend, _pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _counterparty common.Address) (common.Address, *types.Transaction, *Swap, error) {
	parsed, err := SwapMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SwapBin), backend, _pubKeyClaim, _pubKeyRefund, _counterparty)
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
	S   *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterClaimed is a free log retrieval operation binding the contract event 0x7a355715549cfe7c1cba26304350343fbddc4b4f72d3ce3e7c27117dd20b5cb8.
//
// Solidity: event Claimed(uint256 s)
func (_Swap *SwapFilterer) FilterClaimed(opts *bind.FilterOpts) (*SwapClaimedIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "Claimed")
	if err != nil {
		return nil, err
	}
	return &SwapClaimedIterator{contract: _Swap.contract, event: "Claimed", logs: logs, sub: sub}, nil
}

// WatchClaimed is a free log subscription operation binding the contract event 0x7a355715549cfe7c1cba26304350343fbddc4b4f72d3ce3e7c27117dd20b5cb8.
//
// Solidity: event Claimed(uint256 s)
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

// ParseClaimed is a log parse operation binding the contract event 0x7a355715549cfe7c1cba26304350343fbddc4b4f72d3ce3e7c27117dd20b5cb8.
//
// Solidity: event Claimed(uint256 s)
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
	P   [32]byte
	Raw types.Log // Blockchain specific contextual infos
}

// FilterConstructed is a free log retrieval operation binding the contract event 0x1ba2159dcdf5aa313440d6540b9acb108e5f7907737e884db04579a584275fbb.
//
// Solidity: event Constructed(bytes32 p)
func (_Swap *SwapFilterer) FilterConstructed(opts *bind.FilterOpts) (*SwapConstructedIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "Constructed")
	if err != nil {
		return nil, err
	}
	return &SwapConstructedIterator{contract: _Swap.contract, event: "Constructed", logs: logs, sub: sub}, nil
}

// WatchConstructed is a free log subscription operation binding the contract event 0x1ba2159dcdf5aa313440d6540b9acb108e5f7907737e884db04579a584275fbb.
//
// Solidity: event Constructed(bytes32 p)
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

// ParseConstructed is a log parse operation binding the contract event 0x1ba2159dcdf5aa313440d6540b9acb108e5f7907737e884db04579a584275fbb.
//
// Solidity: event Constructed(bytes32 p)
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
	S   *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterRefunded is a free log retrieval operation binding the contract event 0x3d2a04f53164bedf9a8a46353305d6b2d2261410406df3b41f99ce6489dc003c.
//
// Solidity: event Refunded(uint256 s)
func (_Swap *SwapFilterer) FilterRefunded(opts *bind.FilterOpts) (*SwapRefundedIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return &SwapRefundedIterator{contract: _Swap.contract, event: "Refunded", logs: logs, sub: sub}, nil
}

// WatchRefunded is a free log subscription operation binding the contract event 0x3d2a04f53164bedf9a8a46353305d6b2d2261410406df3b41f99ce6489dc003c.
//
// Solidity: event Refunded(uint256 s)
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

// ParseRefunded is a log parse operation binding the contract event 0x3d2a04f53164bedf9a8a46353305d6b2d2261410406df3b41f99ce6489dc003c.
//
// Solidity: event Refunded(uint256 s)
func (_Swap *SwapFilterer) ParseRefunded(log types.Log) (*SwapRefunded, error) {
	event := new(SwapRefunded)
	if err := _Swap.contract.UnpackLog(event, "Refunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
