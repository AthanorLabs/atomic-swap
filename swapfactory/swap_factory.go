// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package swapfactory

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

// SwapFactorySwap is an auto generated low-level Go binding around an user-defined struct.
type SwapFactorySwap struct {
	Owner        common.Address
	Claimer      common.Address
	PubKeyClaim  [32]byte
	PubKeyRefund [32]byte
	Timeout0     *big.Int
	Timeout1     *big.Int
	Value        *big.Int
	Nonce        *big.Int
}

// SwapFactoryMetaData contains all meta data concerning the SwapFactory contract.
var SwapFactoryMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"claimKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"refundKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout_0\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout_1\",\"type\":\"uint256\"}],\"name\":\"New\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"}],\"name\":\"Ready\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout_0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout_1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapFactory.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_swapID\",\"type\":\"bytes32\"}],\"name\":\"is_ready\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"scalar\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"qKeccak\",\"type\":\"uint256\"}],\"name\":\"mulVerify\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"addresspayable\",\"name\":\"_claimer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_timeoutDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"new_swap\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout_0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout_1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapFactory.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout_0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout_1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapFactory.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"}],\"name\":\"set_ready\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"swaps\",\"outputs\":[{\"internalType\":\"enumSwapFactory.Stage\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b5061199d806100206000396000f3fe6080604052600436106100705760003560e01c80637069c7f31161004e5780637069c7f314610104578063b32d1b4f1461012d578063d749b6c41461016a578063eb84e7f21461019a57610070565b8063262cd8da14610075578063268a3bd41461009e5780633e7a7b55146100db575b600080fd5b34801561008157600080fd5b5061009c60048036038101906100979190610f67565b6101d7565b005b3480156100aa57600080fd5b506100c560048036038101906100c09190610fa9565b61048c565b6040516100d29190610ff1565b60405180910390f35b3480156100e757600080fd5b5061010260048036038101906100fd919061100c565b6104dc565b005b34801561011057600080fd5b5061012b60048036038101906101269190610f67565b610675565b005b34801561013957600080fd5b50610154600480360381019061014f919061103a565b610961565b6040516101619190610ff1565b60405180910390f35b610184600480360381019061017f919061107a565b610a66565b6040516101919190611104565b60405180910390f35b3480156101a657600080fd5b506101c160048036038101906101bc9190610fa9565b610c49565b6040516101ce9190611196565b60405180910390f35b6000826040516020016101ea9190611280565b604051602081830303815290604052805190602001209050600080600083815260200190815260200160002060009054906101000a900460ff1690506003808111156102395761023861111f565b5b81600381111561024c5761024b61111f565b5b1415801561027f5750600060038111156102695761026861111f565b5b81600381111561027c5761027b61111f565b5b14155b6102be576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102b5906112f9565b60405180910390fd5b836000015173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610330576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103279061138b565b60405180910390fd5b8360a001514210158061037757508360800151421080156103765750600260038111156103605761035f61111f565b5b8160038111156103735761037261111f565b5b14155b5b6103b6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103ad9061141d565b60405180910390fd5b6103c4838560600151610c69565b7e7c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f82846040516103f492919061143d565b60405180910390a1836000015173ffffffffffffffffffffffffffffffffffffffff166108fc8560c001519081150290604051600060405180830381858888f1935050505015801561044a573d6000803e3d6000fd5b50600360008084815260200190815260200160002060006101000a81548160ff021916908360038111156104815761048061111f565b5b021790555050505050565b6000600260038111156104a2576104a161111f565b5b60008084815260200190815260200160002060009054906101000a900460ff1660038111156104d4576104d361111f565b5b149050919050565b6000816040516020016104ef9190611280565b6040516020818303038152906040528051906020012090506001600381111561051b5761051a61111f565b5b60008083815260200190815260200160002060009054906101000a900460ff16600381111561054d5761054c61111f565b5b1461058d576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610584906114b2565b60405180910390fd5b3373ffffffffffffffffffffffffffffffffffffffff16826000015173ffffffffffffffffffffffffffffffffffffffff16146105ff576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105f690611544565b60405180910390fd5b600260008083815260200190815260200160002060006101000a81548160ff021916908360038111156106355761063461111f565b5b02179055507f5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f816040516106699190611104565b60405180910390a15050565b6000826040516020016106889190611280565b604051602081830303815290604052805190602001209050600080600083815260200190815260200160002060009054906101000a900460ff1690506003808111156106d7576106d661111f565b5b8160038111156106ea576106e961111f565b5b1415801561071d5750600060038111156107075761070661111f565b5b81600381111561071a5761071961111f565b5b14155b61075c576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610753906112f9565b60405180910390fd5b836020015173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146107ce576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107c5906115b0565b60405180910390fd5b8360800151421015806108055750600260038111156107f0576107ef61111f565b5b8160038111156108035761080261111f565b5b145b610844576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161083b9061161c565b60405180910390fd5b8360a00151421061088a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161088190611688565b60405180910390fd5b610898838560400151610c69565b7f38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee82846040516108c992919061143d565b60405180910390a1836020015173ffffffffffffffffffffffffffffffffffffffff166108fc8560c001519081150290604051600060405180830381858888f1935050505015801561091f573d6000803e3d6000fd5b50600360008084815260200190815260200160002060006101000a81548160ff021916908360038111156109565761095561111f565b5b021790555050505050565b60008060016000601b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179860001b7ffffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141806109bd576109bc6116a8565b5b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798890960001b60405160008152602001604052604051610a009493929190611771565b6020604051602081039080840390855afa158015610a22573d6000803e3d6000fd5b5050506020604051035190508073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161491505092915050565b6000610a70610cbc565b33816000019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505084816020019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505086816040018181525050858160600181815250508342610b0091906117e5565b816080018181525050600284610b169190611819565b42610b2191906117e5565b8160a0018181525050348160c0018181525050828160e0018181525050600081604051602001610b519190611280565b60405160208183030381529060405280519060200120905060006003811115610b7d57610b7c61111f565b5b60008083815260200190815260200160002060009054906101000a900460ff166003811115610baf57610bae61111f565b5b14610bb957600080fd5b7f8116b8ce401b5f8f3bb3b91fc2ac461b29ffe582eff877d50fb5a9f9e54306be81898985608001518660a00151604051610bf8959493929190611882565b60405180910390a1600160008083815260200190815260200160002060006101000a81548160ff02191690836003811115610c3657610c3561111f565b5b0217905550809250505095945050505050565b60006020528060005260406000206000915054906101000a900460ff1681565b610c798260001c8260001c610961565b610cb8576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610caf90611947565b60405180910390fd5b5050565b604051806101000160405280600073ffffffffffffffffffffffffffffffffffffffff168152602001600073ffffffffffffffffffffffffffffffffffffffff1681526020016000801916815260200160008019168152602001600081526020016000815260200160008152602001600081525090565b6000604051905090565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b610d9082610d47565b810181811067ffffffffffffffff82111715610daf57610dae610d58565b5b80604052505050565b6000610dc2610d33565b9050610dce8282610d87565b919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610dfe82610dd3565b9050919050565b610e0e81610df3565b8114610e1957600080fd5b50565b600081359050610e2b81610e05565b92915050565b6000819050919050565b610e4481610e31565b8114610e4f57600080fd5b50565b600081359050610e6181610e3b565b92915050565b6000819050919050565b610e7a81610e67565b8114610e8557600080fd5b50565b600081359050610e9781610e71565b92915050565b60006101008284031215610eb457610eb3610d42565b5b610ebf610100610db8565b90506000610ecf84828501610e1c565b6000830152506020610ee384828501610e1c565b6020830152506040610ef784828501610e52565b6040830152506060610f0b84828501610e52565b6060830152506080610f1f84828501610e88565b60808301525060a0610f3384828501610e88565b60a08301525060c0610f4784828501610e88565b60c08301525060e0610f5b84828501610e88565b60e08301525092915050565b6000806101208385031215610f7f57610f7e610d3d565b5b6000610f8d85828601610e9d565b925050610100610f9f85828601610e52565b9150509250929050565b600060208284031215610fbf57610fbe610d3d565b5b6000610fcd84828501610e52565b91505092915050565b60008115159050919050565b610feb81610fd6565b82525050565b60006020820190506110066000830184610fe2565b92915050565b6000610100828403121561102357611022610d3d565b5b600061103184828501610e9d565b91505092915050565b6000806040838503121561105157611050610d3d565b5b600061105f85828601610e88565b925050602061107085828601610e88565b9150509250929050565b600080600080600060a0868803121561109657611095610d3d565b5b60006110a488828901610e52565b95505060206110b588828901610e52565b94505060406110c688828901610e1c565b93505060606110d788828901610e88565b92505060806110e888828901610e88565b9150509295509295909350565b6110fe81610e31565b82525050565b600060208201905061111960008301846110f5565b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b6004811061115f5761115e61111f565b5b50565b60008190506111708261114e565b919050565b600061118082611162565b9050919050565b61119081611175565b82525050565b60006020820190506111ab6000830184611187565b92915050565b6111ba81610df3565b82525050565b6111c981610e31565b82525050565b6111d881610e67565b82525050565b610100820160008201516111f560008501826111b1565b50602082015161120860208501826111b1565b50604082015161121b60408501826111c0565b50606082015161122e60608501826111c0565b50608082015161124160808501826111cf565b5060a082015161125460a08501826111cf565b5060c082015161126760c08501826111cf565b5060e082015161127a60e08501826111cf565b50505050565b60006101008201905061129660008301846111de565b92915050565b600082825260208201905092915050565b7f7377617020697320616c726561647920636f6d706c6574656400000000000000600082015250565b60006112e360198361129c565b91506112ee826112ad565b602082019050919050565b60006020820190508181036000830152611312816112d6565b9050919050565b7f726566756e64206d7573742062652063616c6c6564206279207468652073776160008201527f70206f776e657200000000000000000000000000000000000000000000000000602082015250565b600061137560278361129c565b915061138082611319565b604082019050919050565b600060208201905081810360008301526113a481611368565b9050919050565b7f697427732074686520636f756e74657270617274792773207475726e2c20756e60008201527f61626c6520746f20726566756e642c2074727920616761696e206c6174657200602082015250565b6000611407603f8361129c565b9150611412826113ab565b604082019050919050565b60006020820190508181036000830152611436816113fa565b9050919050565b600060408201905061145260008301856110f5565b61145f60208301846110f5565b9392505050565b7f73776170206973206e6f7420696e2050454e44494e4720737461746500000000600082015250565b600061149c601c8361129c565b91506114a782611466565b602082019050919050565b600060208201905081810360008301526114cb8161148f565b9050919050565b7f6f6e6c79207468652073776170206f776e65722063616e2063616c6c2073657460008201527f5f72656164790000000000000000000000000000000000000000000000000000602082015250565b600061152e60268361129c565b9150611539826114d2565b604082019050919050565b6000602082019050818103600083015261155d81611521565b9050919050565b7f6f6e6c7920636c61696d65722063616e20636c61696d21000000000000000000600082015250565b600061159a60178361129c565b91506115a582611564565b602082019050919050565b600060208201905081810360008301526115c98161158d565b9050919050565b7f746f6f206561726c7920746f20636c61696d2100000000000000000000000000600082015250565b600061160660138361129c565b9150611611826115d0565b602082019050919050565b60006020820190508181036000830152611635816115f9565b9050919050565b7f746f6f206c61746520746f20636c61696d210000000000000000000000000000600082015250565b600061167260128361129c565b915061167d8261163c565b602082019050919050565b600060208201905081810360008301526116a181611665565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000819050919050565b60008160001b9050919050565b60006117096117046116ff846116d7565b6116e1565b610e31565b9050919050565b611719816116ee565b82525050565b6000819050919050565b600060ff82169050919050565b6000819050919050565b600061175b6117566117518461171f565b611736565b611729565b9050919050565b61176b81611740565b82525050565b60006080820190506117866000830187611710565b6117936020830186611762565b6117a060408301856110f5565b6117ad60608301846110f5565b95945050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60006117f082610e67565b91506117fb83610e67565b9250828201905080821115611813576118126117b6565b5b92915050565b600061182482610e67565b915061182f83610e67565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0483118215151615611868576118676117b6565b5b828202905092915050565b61187c81610e67565b82525050565b600060a08201905061189760008301886110f5565b6118a460208301876110f5565b6118b160408301866110f5565b6118be6060830185611873565b6118cb6080830184611873565b9695505050505050565b7f70726f76696465642073656372657420646f6573206e6f74206d61746368207460008201527f6865206578706563746564207075626c6963206b657900000000000000000000602082015250565b600061193160368361129c565b915061193c826118d5565b604082019050919050565b6000602082019050818103600083015261196081611924565b905091905056fea2646970667358221220094c42ec41545c6c9b3945dee82de32ddcbf5e339fbab943edea38e681e33cc564736f6c63430008100033",
}

// SwapFactoryABI is the input ABI used to generate the binding from.
// Deprecated: Use SwapFactoryMetaData.ABI instead.
var SwapFactoryABI = SwapFactoryMetaData.ABI

// SwapFactoryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SwapFactoryMetaData.Bin instead.
var SwapFactoryBin = SwapFactoryMetaData.Bin

// DeploySwapFactory deploys a new Ethereum contract, binding an instance of SwapFactory to it.
func DeploySwapFactory(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SwapFactory, error) {
	parsed, err := SwapFactoryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SwapFactoryBin), backend)
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

// IsReady is a free data retrieval call binding the contract method 0x268a3bd4.
//
// Solidity: function is_ready(bytes32 _swapID) view returns(bool)
func (_SwapFactory *SwapFactoryCaller) IsReady(opts *bind.CallOpts, _swapID [32]byte) (bool, error) {
	var out []interface{}
	err := _SwapFactory.contract.Call(opts, &out, "is_ready", _swapID)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsReady is a free data retrieval call binding the contract method 0x268a3bd4.
//
// Solidity: function is_ready(bytes32 _swapID) view returns(bool)
func (_SwapFactory *SwapFactorySession) IsReady(_swapID [32]byte) (bool, error) {
	return _SwapFactory.Contract.IsReady(&_SwapFactory.CallOpts, _swapID)
}

// IsReady is a free data retrieval call binding the contract method 0x268a3bd4.
//
// Solidity: function is_ready(bytes32 _swapID) view returns(bool)
func (_SwapFactory *SwapFactoryCallerSession) IsReady(_swapID [32]byte) (bool, error) {
	return _SwapFactory.Contract.IsReady(&_SwapFactory.CallOpts, _swapID)
}

// MulVerify is a free data retrieval call binding the contract method 0xb32d1b4f.
//
// Solidity: function mulVerify(uint256 scalar, uint256 qKeccak) pure returns(bool)
func (_SwapFactory *SwapFactoryCaller) MulVerify(opts *bind.CallOpts, scalar *big.Int, qKeccak *big.Int) (bool, error) {
	var out []interface{}
	err := _SwapFactory.contract.Call(opts, &out, "mulVerify", scalar, qKeccak)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// MulVerify is a free data retrieval call binding the contract method 0xb32d1b4f.
//
// Solidity: function mulVerify(uint256 scalar, uint256 qKeccak) pure returns(bool)
func (_SwapFactory *SwapFactorySession) MulVerify(scalar *big.Int, qKeccak *big.Int) (bool, error) {
	return _SwapFactory.Contract.MulVerify(&_SwapFactory.CallOpts, scalar, qKeccak)
}

// MulVerify is a free data retrieval call binding the contract method 0xb32d1b4f.
//
// Solidity: function mulVerify(uint256 scalar, uint256 qKeccak) pure returns(bool)
func (_SwapFactory *SwapFactoryCallerSession) MulVerify(scalar *big.Int, qKeccak *big.Int) (bool, error) {
	return _SwapFactory.Contract.MulVerify(&_SwapFactory.CallOpts, scalar, qKeccak)
}

// Swaps is a free data retrieval call binding the contract method 0xeb84e7f2.
//
// Solidity: function swaps(bytes32 ) view returns(uint8)
func (_SwapFactory *SwapFactoryCaller) Swaps(opts *bind.CallOpts, arg0 [32]byte) (uint8, error) {
	var out []interface{}
	err := _SwapFactory.contract.Call(opts, &out, "swaps", arg0)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Swaps is a free data retrieval call binding the contract method 0xeb84e7f2.
//
// Solidity: function swaps(bytes32 ) view returns(uint8)
func (_SwapFactory *SwapFactorySession) Swaps(arg0 [32]byte) (uint8, error) {
	return _SwapFactory.Contract.Swaps(&_SwapFactory.CallOpts, arg0)
}

// Swaps is a free data retrieval call binding the contract method 0xeb84e7f2.
//
// Solidity: function swaps(bytes32 ) view returns(uint8)
func (_SwapFactory *SwapFactoryCallerSession) Swaps(arg0 [32]byte) (uint8, error) {
	return _SwapFactory.Contract.Swaps(&_SwapFactory.CallOpts, arg0)
}

// Claim is a paid mutator transaction binding the contract method 0x7069c7f3.
//
// Solidity: function claim((address,address,bytes32,bytes32,uint256,uint256,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapFactory *SwapFactoryTransactor) Claim(opts *bind.TransactOpts, _swap SwapFactorySwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "claim", _swap, _s)
}

// Claim is a paid mutator transaction binding the contract method 0x7069c7f3.
//
// Solidity: function claim((address,address,bytes32,bytes32,uint256,uint256,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapFactory *SwapFactorySession) Claim(_swap SwapFactorySwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.Contract.Claim(&_SwapFactory.TransactOpts, _swap, _s)
}

// Claim is a paid mutator transaction binding the contract method 0x7069c7f3.
//
// Solidity: function claim((address,address,bytes32,bytes32,uint256,uint256,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapFactory *SwapFactoryTransactorSession) Claim(_swap SwapFactorySwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.Contract.Claim(&_SwapFactory.TransactOpts, _swap, _s)
}

// NewSwap is a paid mutator transaction binding the contract method 0xd749b6c4.
//
// Solidity: function new_swap(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, address _claimer, uint256 _timeoutDuration, uint256 _nonce) payable returns(bytes32)
func (_SwapFactory *SwapFactoryTransactor) NewSwap(opts *bind.TransactOpts, _pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration *big.Int, _nonce *big.Int) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "new_swap", _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration, _nonce)
}

// NewSwap is a paid mutator transaction binding the contract method 0xd749b6c4.
//
// Solidity: function new_swap(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, address _claimer, uint256 _timeoutDuration, uint256 _nonce) payable returns(bytes32)
func (_SwapFactory *SwapFactorySession) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration *big.Int, _nonce *big.Int) (*types.Transaction, error) {
	return _SwapFactory.Contract.NewSwap(&_SwapFactory.TransactOpts, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration, _nonce)
}

// NewSwap is a paid mutator transaction binding the contract method 0xd749b6c4.
//
// Solidity: function new_swap(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, address _claimer, uint256 _timeoutDuration, uint256 _nonce) payable returns(bytes32)
func (_SwapFactory *SwapFactoryTransactorSession) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration *big.Int, _nonce *big.Int) (*types.Transaction, error) {
	return _SwapFactory.Contract.NewSwap(&_SwapFactory.TransactOpts, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration, _nonce)
}

// Refund is a paid mutator transaction binding the contract method 0x262cd8da.
//
// Solidity: function refund((address,address,bytes32,bytes32,uint256,uint256,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapFactory *SwapFactoryTransactor) Refund(opts *bind.TransactOpts, _swap SwapFactorySwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "refund", _swap, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x262cd8da.
//
// Solidity: function refund((address,address,bytes32,bytes32,uint256,uint256,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapFactory *SwapFactorySession) Refund(_swap SwapFactorySwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.Contract.Refund(&_SwapFactory.TransactOpts, _swap, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x262cd8da.
//
// Solidity: function refund((address,address,bytes32,bytes32,uint256,uint256,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapFactory *SwapFactoryTransactorSession) Refund(_swap SwapFactorySwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.Contract.Refund(&_SwapFactory.TransactOpts, _swap, _s)
}

// SetReady is a paid mutator transaction binding the contract method 0x3e7a7b55.
//
// Solidity: function set_ready((address,address,bytes32,bytes32,uint256,uint256,uint256,uint256) _swap) returns()
func (_SwapFactory *SwapFactoryTransactor) SetReady(opts *bind.TransactOpts, _swap SwapFactorySwap) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "set_ready", _swap)
}

// SetReady is a paid mutator transaction binding the contract method 0x3e7a7b55.
//
// Solidity: function set_ready((address,address,bytes32,bytes32,uint256,uint256,uint256,uint256) _swap) returns()
func (_SwapFactory *SwapFactorySession) SetReady(_swap SwapFactorySwap) (*types.Transaction, error) {
	return _SwapFactory.Contract.SetReady(&_SwapFactory.TransactOpts, _swap)
}

// SetReady is a paid mutator transaction binding the contract method 0x3e7a7b55.
//
// Solidity: function set_ready((address,address,bytes32,bytes32,uint256,uint256,uint256,uint256) _swap) returns()
func (_SwapFactory *SwapFactoryTransactorSession) SetReady(_swap SwapFactorySwap) (*types.Transaction, error) {
	return _SwapFactory.Contract.SetReady(&_SwapFactory.TransactOpts, _swap)
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
	SwapID [32]byte
	S      [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterClaimed is a free log retrieval operation binding the contract event 0x38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee.
//
// Solidity: event Claimed(bytes32 swapID, bytes32 s)
func (_SwapFactory *SwapFactoryFilterer) FilterClaimed(opts *bind.FilterOpts) (*SwapFactoryClaimedIterator, error) {

	logs, sub, err := _SwapFactory.contract.FilterLogs(opts, "Claimed")
	if err != nil {
		return nil, err
	}
	return &SwapFactoryClaimedIterator{contract: _SwapFactory.contract, event: "Claimed", logs: logs, sub: sub}, nil
}

// WatchClaimed is a free log subscription operation binding the contract event 0x38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee.
//
// Solidity: event Claimed(bytes32 swapID, bytes32 s)
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

// ParseClaimed is a log parse operation binding the contract event 0x38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee.
//
// Solidity: event Claimed(bytes32 swapID, bytes32 s)
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
	SwapID    [32]byte
	ClaimKey  [32]byte
	RefundKey [32]byte
	Timeout0  *big.Int
	Timeout1  *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterNew is a free log retrieval operation binding the contract event 0x8116b8ce401b5f8f3bb3b91fc2ac461b29ffe582eff877d50fb5a9f9e54306be.
//
// Solidity: event New(bytes32 swapID, bytes32 claimKey, bytes32 refundKey, uint256 timeout_0, uint256 timeout_1)
func (_SwapFactory *SwapFactoryFilterer) FilterNew(opts *bind.FilterOpts) (*SwapFactoryNewIterator, error) {

	logs, sub, err := _SwapFactory.contract.FilterLogs(opts, "New")
	if err != nil {
		return nil, err
	}
	return &SwapFactoryNewIterator{contract: _SwapFactory.contract, event: "New", logs: logs, sub: sub}, nil
}

// WatchNew is a free log subscription operation binding the contract event 0x8116b8ce401b5f8f3bb3b91fc2ac461b29ffe582eff877d50fb5a9f9e54306be.
//
// Solidity: event New(bytes32 swapID, bytes32 claimKey, bytes32 refundKey, uint256 timeout_0, uint256 timeout_1)
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

// ParseNew is a log parse operation binding the contract event 0x8116b8ce401b5f8f3bb3b91fc2ac461b29ffe582eff877d50fb5a9f9e54306be.
//
// Solidity: event New(bytes32 swapID, bytes32 claimKey, bytes32 refundKey, uint256 timeout_0, uint256 timeout_1)
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
	SwapID [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterReady is a free log retrieval operation binding the contract event 0x5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f.
//
// Solidity: event Ready(bytes32 swapID)
func (_SwapFactory *SwapFactoryFilterer) FilterReady(opts *bind.FilterOpts) (*SwapFactoryReadyIterator, error) {

	logs, sub, err := _SwapFactory.contract.FilterLogs(opts, "Ready")
	if err != nil {
		return nil, err
	}
	return &SwapFactoryReadyIterator{contract: _SwapFactory.contract, event: "Ready", logs: logs, sub: sub}, nil
}

// WatchReady is a free log subscription operation binding the contract event 0x5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f.
//
// Solidity: event Ready(bytes32 swapID)
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

// ParseReady is a log parse operation binding the contract event 0x5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f.
//
// Solidity: event Ready(bytes32 swapID)
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
	SwapID [32]byte
	S      [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterRefunded is a free log retrieval operation binding the contract event 0x007c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f.
//
// Solidity: event Refunded(bytes32 swapID, bytes32 s)
func (_SwapFactory *SwapFactoryFilterer) FilterRefunded(opts *bind.FilterOpts) (*SwapFactoryRefundedIterator, error) {

	logs, sub, err := _SwapFactory.contract.FilterLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return &SwapFactoryRefundedIterator{contract: _SwapFactory.contract, event: "Refunded", logs: logs, sub: sub}, nil
}

// WatchRefunded is a free log subscription operation binding the contract event 0x007c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f.
//
// Solidity: event Refunded(bytes32 swapID, bytes32 s)
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

// ParseRefunded is a log parse operation binding the contract event 0x007c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f.
//
// Solidity: event Refunded(bytes32 swapID, bytes32 s)
func (_SwapFactory *SwapFactoryFilterer) ParseRefunded(log types.Log) (*SwapFactoryRefunded, error) {
	event := new(SwapFactoryRefunded)
	if err := _SwapFactory.contract.UnpackLog(event, "Refunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
