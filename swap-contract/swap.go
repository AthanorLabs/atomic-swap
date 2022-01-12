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
const SwapABI = "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"addresspayable\",\"name\":\"_claimer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_timeoutDuration\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"claimKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"refundKey\",\"type\":\"bytes32\"}],\"name\":\"Constructed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"b\",\"type\":\"bool\"}],\"name\":\"Ready\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isReady\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pubKeyClaim\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pubKeyRefund\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"set_ready\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeout_0\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeout_1\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// SwapBin is the compiled bytecode used for deploying new contracts.
var SwapBin = "0x61016060405260008060006101000a81548160ff02191690831515021790555060405162001ef738038062001ef7833981810160405281019062000044919062000289565b3373ffffffffffffffffffffffffffffffffffffffff1660a08173ffffffffffffffffffffffffffffffffffffffff16815250508360e081815250508261010081815250508173ffffffffffffffffffffffffffffffffffffffff1660c08173ffffffffffffffffffffffffffffffffffffffff16815250508042620000cb91906200032a565b6101208181525050600281620000e2919062000387565b42620000ef91906200032a565b610140818152505060405162000105906200019b565b604051809103906000f08015801562000122573d6000803e3d6000fd5b5073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff16815250507f8d36aa70807342c3036697a846281194626fd4afa892356ad5979e03831ab080848460405162000189929190620003f9565b60405180910390a15050505062000426565b610df7806200110083390190565b600080fd5b6000819050919050565b620001c381620001ae565b8114620001cf57600080fd5b50565b600081519050620001e381620001b8565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006200021682620001e9565b9050919050565b620002288162000209565b81146200023457600080fd5b50565b60008151905062000248816200021d565b92915050565b6000819050919050565b62000263816200024e565b81146200026f57600080fd5b50565b600081519050620002838162000258565b92915050565b60008060008060808587031215620002a657620002a5620001a9565b5b6000620002b687828801620001d2565b9450506020620002c987828801620001d2565b9350506040620002dc8782880162000237565b9250506060620002ef8782880162000272565b91505092959194509250565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600062000337826200024e565b915062000344836200024e565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156200037c576200037b620002fb565b5b828201905092915050565b600062000394826200024e565b9150620003a1836200024e565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0483118215151615620003dd57620003dc620002fb565b5b828202905092915050565b620003f381620001ae565b82525050565b6000604082019050620004106000830185620003e8565b6200041f6020830184620003e8565b9392505050565b60805160a05160c05160e051610100516101205161014051610c40620004c06000396000818161018b0152818161022b01526105980152600081816101af01528181610255015261052001526000818161016701526102d301526000818161039a01526105fe015260008181610492015261065b0152600081816101d30152818161033001526103d4015260006106c80152610c406000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c8063736290f81161005b578063736290f81461010357806374d7c13814610121578063a094a0311461012b578063bd66528a1461014957610088565b806303f7e2461461008d57806345bb8e09146100ab5780634ded8d52146100c95780637249fbb6146100e7575b600080fd5b610095610165565b6040516100a291906107e2565b60405180910390f35b6100b3610189565b6040516100c09190610816565b60405180910390f35b6100d16101ad565b6040516100de9190610816565b60405180910390f35b61010160048036038101906100fc9190610862565b6101d1565b005b61010b610398565b60405161011891906107e2565b60405180910390f35b6101296103bc565b005b61013361047f565b60405161014091906108aa565b60405180910390f35b610163600480360381019061015e9190610862565b610490565b005b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461022957600080fd5b7f00000000000000000000000000000000000000000000000000000000000000004210158061028e57507f00000000000000000000000000000000000000000000000000000000000000004210801561028d575060008054906101000a900460ff16155b5b6102cd576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102c490610948565b60405180910390fd5b6102f7817f00000000000000000000000000000000000000000000000000000000000000006106c3565b7ffe509803c09416b28ff3d8f690c8b0c61462a892c46d5430c8fb20abe472daf08160405161032691906107e2565b60405180910390a17f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166108fc479081150290604051600060405180830381858888f19350505050158015610394573d6000803e3d6000fd5b5050565b7f000000000000000000000000000000000000000000000000000000000000000081565b60008054906101000a900460ff1615801561042257507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b61042b57600080fd5b60016000806101000a81548160ff0219169083151502179055507fb54ee60cc7bf27004d4c21b3226232af966dcdb31a046c95533970b3eea24ae9600160405161047591906108aa565b60405180910390a1565b60008054906101000a900460ff1681565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461051e576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610515906109b4565b60405180910390fd5b7f000000000000000000000000000000000000000000000000000000000000000042101580610557575060008054906101000a900460ff165b610596576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161058d90610a20565b60405180910390fd5b7f000000000000000000000000000000000000000000000000000000000000000042106105f8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105ef90610a8c565b60405180910390fd5b610622817f00000000000000000000000000000000000000000000000000000000000000006106c3565b7feddf608ef698454af2fb41c1df7b7e5154ff0d46969f895e0f39c7dfe7e6380a8160405161065191906107e2565b60405180910390a17f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166108fc479081150290604051600060405180830381858888f193505050501580156106bf573d6000803e3d6000fd5b5050565b6000807f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663c4f4912b8560001c6040518263ffffffff1660e01b81526004016107229190610816565b6040805180830381865afa15801561073e573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107629190610ad8565b91509150600060ff6002846107779190610b47565b901b82179050838160001b146107c2576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107b990610bea565b60405180910390fd5b5050505050565b6000819050919050565b6107dc816107c9565b82525050565b60006020820190506107f760008301846107d3565b92915050565b6000819050919050565b610810816107fd565b82525050565b600060208201905061082b6000830184610807565b92915050565b600080fd5b61083f816107c9565b811461084a57600080fd5b50565b60008135905061085c81610836565b92915050565b60006020828403121561087857610877610831565b5b60006108868482850161084d565b91505092915050565b60008115159050919050565b6108a48161088f565b82525050565b60006020820190506108bf600083018461089b565b92915050565b600082825260208201905092915050565b7f4974277320426f622773207475726e206e6f772c20706c65617365207761697460008201527f2100000000000000000000000000000000000000000000000000000000000000602082015250565b60006109326021836108c5565b915061093d826108d6565b604082019050919050565b6000602082019050818103600083015261096181610925565b9050919050565b7f6f6e6c7920636c61696d65722063616e20636c61696d21000000000000000000600082015250565b600061099e6017836108c5565b91506109a982610968565b602082019050919050565b600060208201905081810360008301526109cd81610991565b9050919050565b7f746f6f206561726c7920746f20636c61696d2100000000000000000000000000600082015250565b6000610a0a6013836108c5565b9150610a15826109d4565b602082019050919050565b60006020820190508181036000830152610a39816109fd565b9050919050565b7f746f6f206c61746520746f20636c61696d210000000000000000000000000000600082015250565b6000610a766012836108c5565b9150610a8182610a40565b602082019050919050565b60006020820190508181036000830152610aa581610a69565b9050919050565b610ab5816107fd565b8114610ac057600080fd5b50565b600081519050610ad281610aac565b92915050565b60008060408385031215610aef57610aee610831565b5b6000610afd85828601610ac3565b9250506020610b0e85828601610ac3565b9150509250929050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000610b52826107fd565b9150610b5d836107fd565b925082610b6d57610b6c610b18565b5b828206905092915050565b7f70726f76696465642073656372657420646f6573206e6f74206d61746368207460008201527f6865206578706563746564207075624b65790000000000000000000000000000602082015250565b6000610bd46032836108c5565b9150610bdf82610b78565b604082019050919050565b60006020820190508181036000830152610c0381610bc7565b905091905056fea264697066735822122077d8592a18526f06b159907123cb53dc9bd8cf8345d8b9e4eb88dba809fa503b64736f6c634300080a0033608060405234801561001057600080fd5b50610dd7806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c8063c4f4912b14610030575b600080fd5b61004a60048036038101906100459190610caa565b610061565b604051610058929190610ce6565b60405180910390f35b60008061006c610c09565b610074610c09565b7f216936d3cd6e53fec0a4e231fdd6dc5c692cc7609525a7b2c9562d608f25d51a8260000181815250507f666666666666666666666666666666666666666666666666666666666666665882602001818152505060018260400181815250506000816000018181525050600181602001818152505060018160400181815250505b600085111561012d57600180861614156101165761011381836101d2565b90505b600185901c945061012682610701565b91506100f5565b600061013c8260400151610b6b565b90507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061016d5761016c610d0f565b5b818360000151098260000181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806101ac576101ab610d0f565b5b818360200151098260200181815250508160000151826020015194509450505050915091565b6101da610c09565b6101e2610c2a565b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061021157610210610d0f565b5b83604001518560400151098160000181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061025457610253610d0f565b5b81600001518260000151098160200181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061029757610296610d0f565b5b83600001518560000151098160400181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806102da576102d9610d0f565b5b83602001518560200151098160600181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061031d5761031c610d0f565b5b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061034c5761034b610d0f565b5b82606001518360400151097f52036cee2b6ffe738cc740797779e89800700a4d4141d8ab75eb4dca135978a3098160800181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806103b1576103b0610d0f565b5b81608001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed6103e19190610d6d565b8260200151088160a00181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061041f5761041e610d0f565b5b81608001518260200151088160c00181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061046257610461610d0f565b5b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061049157610490610d0f565b5b82606001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed6104c19190610d6d565b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806104f0576104ef610d0f565b5b84604001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed6105209190610d6d565b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061054f5761054e610d0f565b5b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061057e5761057d610d0f565b5b89602001518a60000151087f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806105b8576105b7610d0f565b5b8b602001518c60000151080908087f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806105f5576105f4610d0f565b5b8360a00151846000015109098260000181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061063957610638610d0f565b5b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061066857610667610d0f565b5b82604001518360600151087f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806106a2576106a1610d0f565b5b8360c00151846000015109098260200181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806106e6576106e5610d0f565b5b8160c001518260a00151098260400181815250505092915050565b610709610c09565b610711610c2a565b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806107405761073f610d0f565b5b83602001518460000151088160000181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061078357610782610d0f565b5b81600001518260000151098160200181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806107c6576107c5610d0f565b5b83600001518460000151098160400181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061080957610808610d0f565b5b836020015184602001510981606001818152505080604001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed61084d9190610d6d565b8160800181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061088557610884610d0f565b5b81606001518260800151088160a00181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806108c8576108c7610d0f565b5b83604001518460400151098160e00181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061090b5761090a610d0f565b5b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed8061093a57610939610d0f565b5b8260e001516002097f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed61096d9190610d6d565b8260a00151088160c00181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806109ab576109aa610d0f565b5b8160c001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed806109df576109de610d0f565b5b83606001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed610a0f9190610d6d565b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed80610a3e57610a3d610d0f565b5b85604001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed610a6e9190610d6d565b86602001510808098260000181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed80610aae57610aad610d0f565b5b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed80610add57610adc610d0f565b5b82606001517f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed610b0d9190610d6d565b8360800151088260a00151098260200181815250507f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed80610b5157610b50610d0f565b5b8160c001518260a001510982604001818152505050919050565b60008060027f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed610b9b9190610d6d565b905060007f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed905060405160208152602080820152602060408201528460608201528260808201528160a082015260208160c0836005600019fa610bfd57600080fd5b80519350505050919050565b60405180606001604052806000815260200160008152602001600081525090565b60405180610100016040528060008152602001600081526020016000815260200160008152602001600081526020016000815260200160008152602001600081525090565b600080fd5b6000819050919050565b610c8781610c74565b8114610c9257600080fd5b50565b600081359050610ca481610c7e565b92915050565b600060208284031215610cc057610cbf610c6f565b5b6000610cce84828501610c95565b91505092915050565b610ce081610c74565b82525050565b6000604082019050610cfb6000830185610cd7565b610d086020830184610cd7565b9392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610d7882610c74565b9150610d8383610c74565b925082821015610d9657610d95610d3e565b5b82820390509291505056fea2646970667358221220228157db5b4e262c7044427ba6e55cfb86010de5e71794b75d39ce32a4cf62d364736f6c634300080a0033"

// DeploySwap deploys a new Ethereum contract, binding an instance of Swap to it.
func DeploySwap(auth *bind.TransactOpts, backend bind.ContractBackend, _pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration *big.Int) (common.Address, *types.Transaction, *Swap, error) {
	parsed, err := abi.JSON(strings.NewReader(SwapABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SwapBin), backend, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration)
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

// IsReady is a free data retrieval call binding the contract method 0xa094a031.
//
// Solidity: function isReady() view returns(bool)
func (_Swap *SwapCaller) IsReady(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "isReady")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsReady is a free data retrieval call binding the contract method 0xa094a031.
//
// Solidity: function isReady() view returns(bool)
func (_Swap *SwapSession) IsReady() (bool, error) {
	return _Swap.Contract.IsReady(&_Swap.CallOpts)
}

// IsReady is a free data retrieval call binding the contract method 0xa094a031.
//
// Solidity: function isReady() view returns(bool)
func (_Swap *SwapCallerSession) IsReady() (bool, error) {
	return _Swap.Contract.IsReady(&_Swap.CallOpts)
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
	ClaimKey  [32]byte
	RefundKey [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterConstructed is a free log retrieval operation binding the contract event 0x8d36aa70807342c3036697a846281194626fd4afa892356ad5979e03831ab080.
//
// Solidity: event Constructed(bytes32 claimKey, bytes32 refundKey)
func (_Swap *SwapFilterer) FilterConstructed(opts *bind.FilterOpts) (*SwapConstructedIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "Constructed")
	if err != nil {
		return nil, err
	}
	return &SwapConstructedIterator{contract: _Swap.contract, event: "Constructed", logs: logs, sub: sub}, nil
}

// WatchConstructed is a free log subscription operation binding the contract event 0x8d36aa70807342c3036697a846281194626fd4afa892356ad5979e03831ab080.
//
// Solidity: event Constructed(bytes32 claimKey, bytes32 refundKey)
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
// Solidity: event Constructed(bytes32 claimKey, bytes32 refundKey)
func (_Swap *SwapFilterer) ParseConstructed(log types.Log) (*SwapConstructed, error) {
	event := new(SwapConstructed)
	if err := _Swap.contract.UnpackLog(event, "Constructed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapReadyIterator is returned from FilterReady and is used to iterate over the raw logs and unpacked data for Ready events raised by the Swap contract.
type SwapReadyIterator struct {
	Event *SwapReady // Event containing the contract specifics and raw log

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
func (it *SwapReadyIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapReady)
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
		it.Event = new(SwapReady)
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
func (it *SwapReadyIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapReadyIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapReady represents a Ready event raised by the Swap contract.
type SwapReady struct {
	B   bool
	Raw types.Log // Blockchain specific contextual infos
}

// FilterReady is a free log retrieval operation binding the contract event 0xb54ee60cc7bf27004d4c21b3226232af966dcdb31a046c95533970b3eea24ae9.
//
// Solidity: event Ready(bool b)
func (_Swap *SwapFilterer) FilterReady(opts *bind.FilterOpts) (*SwapReadyIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "Ready")
	if err != nil {
		return nil, err
	}
	return &SwapReadyIterator{contract: _Swap.contract, event: "Ready", logs: logs, sub: sub}, nil
}

// WatchReady is a free log subscription operation binding the contract event 0xb54ee60cc7bf27004d4c21b3226232af966dcdb31a046c95533970b3eea24ae9.
//
// Solidity: event Ready(bool b)
func (_Swap *SwapFilterer) WatchReady(opts *bind.WatchOpts, sink chan<- *SwapReady) (event.Subscription, error) {

	logs, sub, err := _Swap.contract.WatchLogs(opts, "Ready")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapReady)
				if err := _Swap.contract.UnpackLog(event, "Ready", log); err != nil {
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

// ParseReady is a log parse operation binding the contract event 0xb54ee60cc7bf27004d4c21b3226232af966dcdb31a046c95533970b3eea24ae9.
//
// Solidity: event Ready(bool b)
func (_Swap *SwapFilterer) ParseReady(log types.Log) (*SwapReady, error) {
	event := new(SwapReady)
	if err := _Swap.contract.UnpackLog(event, "Ready", log); err != nil {
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
