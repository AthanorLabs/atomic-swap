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

// SwapCreatorRelaySwap is an auto generated low-level Go binding around an user-defined struct.
type SwapCreatorRelaySwap struct {
	Swap        SwapCreatorSwap
	Fee         *big.Int
	Relayer     common.Address
	SwapCreator common.Address
}

// SwapCreatorSwap is an auto generated low-level Go binding around an user-defined struct.
type SwapCreatorSwap struct {
	Owner        common.Address
	Claimer      common.Address
	PubKeyClaim  [32]byte
	PubKeyRefund [32]byte
	Timeout0     *big.Int
	Timeout1     *big.Int
	Asset        common.Address
	Value        *big.Int
	Nonce        *big.Int
}

// SwapCreatorMetaData contains all meta data concerning the SwapCreator contract.
var SwapCreatorMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"InvalidClaimer\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidContractAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSecret\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSwap\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSwapKey\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidTimeout\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidValue\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotTimeToRefund\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OnlySwapClaimer\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OnlySwapOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SwapAlreadyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SwapCompleted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SwapNotPending\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TooEarlyToClaim\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TooLateToClaim\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroValue\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"claimKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"refundKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"New\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"}],\"name\":\"Ready\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapCreator.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_secret\",\"type\":\"bytes32\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapCreator.Swap\",\"name\":\"swap\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"relayer\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"swapCreator\",\"type\":\"address\"}],\"internalType\":\"structSwapCreator.RelaySwap\",\"name\":\"_relaySwap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_secret\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"claimRelayer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"scalar\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"qKeccak\",\"type\":\"uint256\"}],\"name\":\"mulVerify\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"addresspayable\",\"name\":\"_claimer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_timeoutDuration0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_timeoutDuration1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"newSwap\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapCreator.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_secret\",\"type\":\"bytes32\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapCreator.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"}],\"name\":\"setReady\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"swaps\",\"outputs\":[{\"internalType\":\"enumSwapCreator.Stage\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50611243806100206000396000f3fe6080604052600436106100705760003560e01c8063b32d1b4f1161004e578063b32d1b4f146100d7578063c41e46cf1461010c578063eb84e7f21461012d578063fcaf229c1461016a57600080fd5b80631e6c5acc1461007557806356561693146100975780635cb96916146100b7575b600080fd5b34801561008157600080fd5b50610095610090366004610ed6565b61018a565b005b3480156100a357600080fd5b506100956100b2366004610f14565b6103d4565b3480156100c357600080fd5b506100956100d2366004610ed6565b6106b1565b3480156100e357600080fd5b506100f76100f2366004610fe2565b6107d0565b60405190151581526020015b60405180910390f35b61011f61011a366004611004565b6108a0565b604051908152602001610103565b34801561013957600080fd5b5061015d610148366004611075565b60006020819052908152604090205460ff1681565b60405161010391906110a4565b34801561017657600080fd5b506100956101853660046110cc565b610b82565b60008260405160200161019d9190611158565b60408051601f1981840301815291815281516020928301206000818152928390529082205490925060ff16908160038111156101db576101db61108e565b036101f957604051631115766760e01b815260040160405180910390fd5b600381600381111561020d5761020d61108e565b0361022b5760405163066916a960e01b815260040160405180910390fd5b83516001600160a01b031633146102555760405163148ca24360e11b815260040160405180910390fd5b8360a001514210801561028657508360800151421180610286575060028160038111156102845761028461108e565b145b156102a4576040516332a1860f60e11b815260040160405180910390fd5b6102b2838560600151610c60565b604051839083907e7c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f90600090a36000828152602081905260409020805460ff1916600317905560c08401516001600160a01b031661034c57835160e08501516040516001600160a01b039092169181156108fc0291906000818181858888f19350505050158015610346573d6000803e3d6000fd5b506103ce565b60c0840151845160e086015160405163a9059cbb60e01b81526001600160a01b039283166004820152602481019190915291169063a9059cbb906044016020604051808303816000875af11580156103a8573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103cc9190611167565b505b50505050565b60006001866040516020016103e99190611189565b60408051601f198184030181528282528051602091820120600084529083018083525260ff871690820152606081018590526080810184905260a0016020604051602081039080840390855afa158015610447573d6000803e3d6000fd5b5050506020604051035190508560000151602001516001600160a01b0316816001600160a01b03161461048d57604051638baa579f60e01b815260040160405180910390fd5b85606001516001600160a01b0316306001600160a01b0316146104c35760405163a710429d60e01b815260040160405180910390fd5b85516104cf9086610c87565b855160c001516001600160a01b031661057f578560000151602001516001600160a01b03166108fc8760200151886000015160e0015161050f91906111e7565b6040518115909202916000818181858888f19350505050158015610537573d6000803e3d6000fd5b5085604001516001600160a01b03166108fc87602001519081150290604051600060405180830381858888f19350505050158015610579573d6000803e3d6000fd5b506106a9565b855160c08101516020808301519089015160e0909301516001600160a01b039092169263a9059cbb926105b291906111e7565b6040516001600160e01b031960e085901b1681526001600160a01b03909216600483015260248201526044016020604051808303816000875af11580156105fd573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106219190611167565b50855160c001516040808801516020890151915163a9059cbb60e01b81526001600160a01b03918216600482015260248101929092529091169063a9059cbb906044016020604051808303816000875af1158015610683573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106a79190611167565b505b505050505050565b81602001516001600160a01b0316336001600160a01b0316146106e757604051633471640960e11b815260040160405180910390fd5b6106f18282610c87565b60c08201516001600160a01b03166107495781602001516001600160a01b03166108fc8360e001519081150290604051600060405180830381858888f19350505050158015610744573d6000803e3d6000fd5b505050565b60c0820151602083015160e084015160405163a9059cbb60e01b81526001600160a01b039283166004820152602481019190915291169063a9059cbb906044016020604051808303816000875af11580156107a8573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107449190611167565b5050565b600080600181601b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179870014551231950b75fc4402da1732fc9bebe197f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179889096040805160008152602081018083529590955260ff909316928401929092526060830152608082015260a0016020604051602081039080840390855afa15801561087d573d6000803e3d6000fd5b5050604051601f1901516001600160a01b03858116911614925050505b92915050565b6000826000036108c357604051637c946ed760e01b815260040160405180910390fd5b6001600160a01b0384166108f6573483146108f157604051632a9ffab760e21b815260040160405180910390fd5b61096f565b6040516323b872dd60e01b8152336004820152306024820152604481018490526001600160a01b038516906323b872dd906064016020604051808303816000875af1158015610949573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061096d9190611167565b505b88158061097a575087155b1561099857604051631bc61bed60e11b815260040160405180910390fd5b6001600160a01b0387166109be576040516208978560e71b815260040160405180910390fd5b8515806109c9575084155b156109e757604051631ffb86f160e21b815260040160405180910390fd5b6000604051806101200160405280336001600160a01b03168152602001896001600160a01b031681526020018b81526020018a81526020018842610a2b91906111fa565b815260200187610a3b8a426111fa565b610a4591906111fa565b8152602001866001600160a01b03168152602001858152602001848152509050600081604051602001610a789190611158565b60408051601f19818403018152919052805160209091012090506000808281526020819052604090205460ff166003811115610ab657610ab661108e565b14610ad4576040516339a2986760e11b815260040160405180910390fd5b7f91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56818c8c85608001518660a001518760c001518860e00151604051610b53979695949392919096875260208701959095526040860193909352606085019190915260808401526001600160a01b031660a083015260c082015260e00190565b60405180910390a16000818152602081905260409020805460ff191660011790559a9950505050505050505050565b600081604051602001610b959190611158565b60408051601f1981840301815291905280516020909101209050600160008281526020819052604090205460ff166003811115610bd457610bd461108e565b14610bf257604051630fe0fb5160e11b815260040160405180910390fd5b81516001600160a01b03163314610c1c5760405163148ca24360e11b815260040160405180910390fd5b600081815260208190526040808220805460ff191660021790555182917f5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f91a25050565b610c6a82826107d0565b6107cc5760405163abab6bd760e01b815260040160405180910390fd5b600082604051602001610c9a9190611158565b60408051601f1981840301815291815281516020928301206000818152928390529082205490925060ff1690816003811115610cd857610cd861108e565b03610cf657604051631115766760e01b815260040160405180910390fd5b6003816003811115610d0a57610d0a61108e565b03610d285760405163066916a960e01b815260040160405180910390fd5b836080015142108015610d4d57506002816003811115610d4a57610d4a61108e565b14155b15610d6b5760405163d71d60b560e01b815260040160405180910390fd5b8360a001514210610d8f5760405163497df9d160e01b815260040160405180910390fd5b610d9d838560400151610c60565b604051839083907f38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee90600090a3506000908152602081905260409020805460ff191660031790555050565b604051610120810167ffffffffffffffff81118282101715610e1a57634e487b7160e01b600052604160045260246000fd5b60405290565b6001600160a01b0381168114610e3557600080fd5b50565b8035610e4381610e20565b919050565b60006101208284031215610e5b57600080fd5b610e63610de8565b9050610e6e82610e38565b8152610e7c60208301610e38565b602082015260408201356040820152606082013560608201526080820135608082015260a082013560a0820152610eb560c08301610e38565b60c082015260e082013560e082015261010080830135818301525092915050565b6000806101408385031215610eea57600080fd5b610ef48484610e48565b94610120939093013593505050565b803560ff81168114610e4357600080fd5b6000806000806000858703610200811215610f2e57600080fd5b61018080821215610f3e57600080fd5b60405191506080820182811067ffffffffffffffff82111715610f7157634e487b7160e01b600052604160045260246000fd5b604052610f7e8989610e48565b82526101208801356020830152610140880135610f9a81610e20565b6040830152610160880135610fae81610e20565b60608301529095508601359350610fc86101a08701610f03565b949793965093946101c081013594506101e0013592915050565b60008060408385031215610ff557600080fd5b50508035926020909101359150565b600080600080600080600080610100898b03121561102157600080fd5b8835975060208901359650604089013561103a81610e20565b9550606089013594506080890135935060a089013561105881610e20565b979a969950949793969295929450505060c08201359160e0013590565b60006020828403121561108757600080fd5b5035919050565b634e487b7160e01b600052602160045260246000fd5b60208101600483106110c657634e487b7160e01b600052602160045260246000fd5b91905290565b600061012082840312156110df57600080fd5b6110e98383610e48565b9392505050565b60018060a01b0380825116835280602083015116602084015260408201516040840152606082015160608401526080820151608084015260a082015160a08401528060c08301511660c08401525060e081015160e08301526101008082015181840152505050565b610120810161089a82846110f0565b60006020828403121561117957600080fd5b815180151581146110e957600080fd5b60006101808201905061119d8284516110f0565b602083015161012083015260408301516001600160a01b039081166101408401526060909301519092166101609091015290565b634e487b7160e01b600052601160045260246000fd5b8181038181111561089a5761089a6111d1565b8082018082111561089a5761089a6111d156fea26469706673582212206ac5abd4bdeadcd69a91cc110e06a7e6d6b455789d59db647a068af369a4623b64736f6c63430008130033",
}

// SwapCreatorABI is the input ABI used to generate the binding from.
// Deprecated: Use SwapCreatorMetaData.ABI instead.
var SwapCreatorABI = SwapCreatorMetaData.ABI

// SwapCreatorBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SwapCreatorMetaData.Bin instead.
var SwapCreatorBin = SwapCreatorMetaData.Bin

// DeploySwapCreator deploys a new Ethereum contract, binding an instance of SwapCreator to it.
func DeploySwapCreator(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SwapCreator, error) {
	parsed, err := SwapCreatorMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SwapCreatorBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SwapCreator{SwapCreatorCaller: SwapCreatorCaller{contract: contract}, SwapCreatorTransactor: SwapCreatorTransactor{contract: contract}, SwapCreatorFilterer: SwapCreatorFilterer{contract: contract}}, nil
}

// SwapCreator is an auto generated Go binding around an Ethereum contract.
type SwapCreator struct {
	SwapCreatorCaller     // Read-only binding to the contract
	SwapCreatorTransactor // Write-only binding to the contract
	SwapCreatorFilterer   // Log filterer for contract events
}

// SwapCreatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type SwapCreatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapCreatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SwapCreatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapCreatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SwapCreatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapCreatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SwapCreatorSession struct {
	Contract     *SwapCreator      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SwapCreatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SwapCreatorCallerSession struct {
	Contract *SwapCreatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// SwapCreatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SwapCreatorTransactorSession struct {
	Contract     *SwapCreatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// SwapCreatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type SwapCreatorRaw struct {
	Contract *SwapCreator // Generic contract binding to access the raw methods on
}

// SwapCreatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SwapCreatorCallerRaw struct {
	Contract *SwapCreatorCaller // Generic read-only contract binding to access the raw methods on
}

// SwapCreatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SwapCreatorTransactorRaw struct {
	Contract *SwapCreatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSwapCreator creates a new instance of SwapCreator, bound to a specific deployed contract.
func NewSwapCreator(address common.Address, backend bind.ContractBackend) (*SwapCreator, error) {
	contract, err := bindSwapCreator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SwapCreator{SwapCreatorCaller: SwapCreatorCaller{contract: contract}, SwapCreatorTransactor: SwapCreatorTransactor{contract: contract}, SwapCreatorFilterer: SwapCreatorFilterer{contract: contract}}, nil
}

// NewSwapCreatorCaller creates a new read-only instance of SwapCreator, bound to a specific deployed contract.
func NewSwapCreatorCaller(address common.Address, caller bind.ContractCaller) (*SwapCreatorCaller, error) {
	contract, err := bindSwapCreator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SwapCreatorCaller{contract: contract}, nil
}

// NewSwapCreatorTransactor creates a new write-only instance of SwapCreator, bound to a specific deployed contract.
func NewSwapCreatorTransactor(address common.Address, transactor bind.ContractTransactor) (*SwapCreatorTransactor, error) {
	contract, err := bindSwapCreator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SwapCreatorTransactor{contract: contract}, nil
}

// NewSwapCreatorFilterer creates a new log filterer instance of SwapCreator, bound to a specific deployed contract.
func NewSwapCreatorFilterer(address common.Address, filterer bind.ContractFilterer) (*SwapCreatorFilterer, error) {
	contract, err := bindSwapCreator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SwapCreatorFilterer{contract: contract}, nil
}

// bindSwapCreator binds a generic wrapper to an already deployed contract.
func bindSwapCreator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SwapCreatorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SwapCreator *SwapCreatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SwapCreator.Contract.SwapCreatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SwapCreator *SwapCreatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwapCreator.Contract.SwapCreatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SwapCreator *SwapCreatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SwapCreator.Contract.SwapCreatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SwapCreator *SwapCreatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SwapCreator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SwapCreator *SwapCreatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwapCreator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SwapCreator *SwapCreatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SwapCreator.Contract.contract.Transact(opts, method, params...)
}

// MulVerify is a free data retrieval call binding the contract method 0xb32d1b4f.
//
// Solidity: function mulVerify(uint256 scalar, uint256 qKeccak) pure returns(bool)
func (_SwapCreator *SwapCreatorCaller) MulVerify(opts *bind.CallOpts, scalar *big.Int, qKeccak *big.Int) (bool, error) {
	var out []interface{}
	err := _SwapCreator.contract.Call(opts, &out, "mulVerify", scalar, qKeccak)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// MulVerify is a free data retrieval call binding the contract method 0xb32d1b4f.
//
// Solidity: function mulVerify(uint256 scalar, uint256 qKeccak) pure returns(bool)
func (_SwapCreator *SwapCreatorSession) MulVerify(scalar *big.Int, qKeccak *big.Int) (bool, error) {
	return _SwapCreator.Contract.MulVerify(&_SwapCreator.CallOpts, scalar, qKeccak)
}

// MulVerify is a free data retrieval call binding the contract method 0xb32d1b4f.
//
// Solidity: function mulVerify(uint256 scalar, uint256 qKeccak) pure returns(bool)
func (_SwapCreator *SwapCreatorCallerSession) MulVerify(scalar *big.Int, qKeccak *big.Int) (bool, error) {
	return _SwapCreator.Contract.MulVerify(&_SwapCreator.CallOpts, scalar, qKeccak)
}

// Swaps is a free data retrieval call binding the contract method 0xeb84e7f2.
//
// Solidity: function swaps(bytes32 ) view returns(uint8)
func (_SwapCreator *SwapCreatorCaller) Swaps(opts *bind.CallOpts, arg0 [32]byte) (uint8, error) {
	var out []interface{}
	err := _SwapCreator.contract.Call(opts, &out, "swaps", arg0)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Swaps is a free data retrieval call binding the contract method 0xeb84e7f2.
//
// Solidity: function swaps(bytes32 ) view returns(uint8)
func (_SwapCreator *SwapCreatorSession) Swaps(arg0 [32]byte) (uint8, error) {
	return _SwapCreator.Contract.Swaps(&_SwapCreator.CallOpts, arg0)
}

// Swaps is a free data retrieval call binding the contract method 0xeb84e7f2.
//
// Solidity: function swaps(bytes32 ) view returns(uint8)
func (_SwapCreator *SwapCreatorCallerSession) Swaps(arg0 [32]byte) (uint8, error) {
	return _SwapCreator.Contract.Swaps(&_SwapCreator.CallOpts, arg0)
}

// Claim is a paid mutator transaction binding the contract method 0x5cb96916.
//
// Solidity: function claim((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _secret) returns()
func (_SwapCreator *SwapCreatorTransactor) Claim(opts *bind.TransactOpts, _swap SwapCreatorSwap, _secret [32]byte) (*types.Transaction, error) {
	return _SwapCreator.contract.Transact(opts, "claim", _swap, _secret)
}

// Claim is a paid mutator transaction binding the contract method 0x5cb96916.
//
// Solidity: function claim((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _secret) returns()
func (_SwapCreator *SwapCreatorSession) Claim(_swap SwapCreatorSwap, _secret [32]byte) (*types.Transaction, error) {
	return _SwapCreator.Contract.Claim(&_SwapCreator.TransactOpts, _swap, _secret)
}

// Claim is a paid mutator transaction binding the contract method 0x5cb96916.
//
// Solidity: function claim((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _secret) returns()
func (_SwapCreator *SwapCreatorTransactorSession) Claim(_swap SwapCreatorSwap, _secret [32]byte) (*types.Transaction, error) {
	return _SwapCreator.Contract.Claim(&_SwapCreator.TransactOpts, _swap, _secret)
}

// ClaimRelayer is a paid mutator transaction binding the contract method 0x56561693.
//
// Solidity: function claimRelayer(((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256),uint256,address,address) _relaySwap, bytes32 _secret, uint8 v, bytes32 r, bytes32 s) returns()
func (_SwapCreator *SwapCreatorTransactor) ClaimRelayer(opts *bind.TransactOpts, _relaySwap SwapCreatorRelaySwap, _secret [32]byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _SwapCreator.contract.Transact(opts, "claimRelayer", _relaySwap, _secret, v, r, s)
}

// ClaimRelayer is a paid mutator transaction binding the contract method 0x56561693.
//
// Solidity: function claimRelayer(((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256),uint256,address,address) _relaySwap, bytes32 _secret, uint8 v, bytes32 r, bytes32 s) returns()
func (_SwapCreator *SwapCreatorSession) ClaimRelayer(_relaySwap SwapCreatorRelaySwap, _secret [32]byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _SwapCreator.Contract.ClaimRelayer(&_SwapCreator.TransactOpts, _relaySwap, _secret, v, r, s)
}

// ClaimRelayer is a paid mutator transaction binding the contract method 0x56561693.
//
// Solidity: function claimRelayer(((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256),uint256,address,address) _relaySwap, bytes32 _secret, uint8 v, bytes32 r, bytes32 s) returns()
func (_SwapCreator *SwapCreatorTransactorSession) ClaimRelayer(_relaySwap SwapCreatorRelaySwap, _secret [32]byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _SwapCreator.Contract.ClaimRelayer(&_SwapCreator.TransactOpts, _relaySwap, _secret, v, r, s)
}

// NewSwap is a paid mutator transaction binding the contract method 0xc41e46cf.
//
// Solidity: function newSwap(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, address _claimer, uint256 _timeoutDuration0, uint256 _timeoutDuration1, address _asset, uint256 _value, uint256 _nonce) payable returns(bytes32)
func (_SwapCreator *SwapCreatorTransactor) NewSwap(opts *bind.TransactOpts, _pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration0 *big.Int, _timeoutDuration1 *big.Int, _asset common.Address, _value *big.Int, _nonce *big.Int) (*types.Transaction, error) {
	return _SwapCreator.contract.Transact(opts, "newSwap", _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration0, _timeoutDuration1, _asset, _value, _nonce)
}

// NewSwap is a paid mutator transaction binding the contract method 0xc41e46cf.
//
// Solidity: function newSwap(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, address _claimer, uint256 _timeoutDuration0, uint256 _timeoutDuration1, address _asset, uint256 _value, uint256 _nonce) payable returns(bytes32)
func (_SwapCreator *SwapCreatorSession) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration0 *big.Int, _timeoutDuration1 *big.Int, _asset common.Address, _value *big.Int, _nonce *big.Int) (*types.Transaction, error) {
	return _SwapCreator.Contract.NewSwap(&_SwapCreator.TransactOpts, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration0, _timeoutDuration1, _asset, _value, _nonce)
}

// NewSwap is a paid mutator transaction binding the contract method 0xc41e46cf.
//
// Solidity: function newSwap(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, address _claimer, uint256 _timeoutDuration0, uint256 _timeoutDuration1, address _asset, uint256 _value, uint256 _nonce) payable returns(bytes32)
func (_SwapCreator *SwapCreatorTransactorSession) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration0 *big.Int, _timeoutDuration1 *big.Int, _asset common.Address, _value *big.Int, _nonce *big.Int) (*types.Transaction, error) {
	return _SwapCreator.Contract.NewSwap(&_SwapCreator.TransactOpts, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration0, _timeoutDuration1, _asset, _value, _nonce)
}

// Refund is a paid mutator transaction binding the contract method 0x1e6c5acc.
//
// Solidity: function refund((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _secret) returns()
func (_SwapCreator *SwapCreatorTransactor) Refund(opts *bind.TransactOpts, _swap SwapCreatorSwap, _secret [32]byte) (*types.Transaction, error) {
	return _SwapCreator.contract.Transact(opts, "refund", _swap, _secret)
}

// Refund is a paid mutator transaction binding the contract method 0x1e6c5acc.
//
// Solidity: function refund((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _secret) returns()
func (_SwapCreator *SwapCreatorSession) Refund(_swap SwapCreatorSwap, _secret [32]byte) (*types.Transaction, error) {
	return _SwapCreator.Contract.Refund(&_SwapCreator.TransactOpts, _swap, _secret)
}

// Refund is a paid mutator transaction binding the contract method 0x1e6c5acc.
//
// Solidity: function refund((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _secret) returns()
func (_SwapCreator *SwapCreatorTransactorSession) Refund(_swap SwapCreatorSwap, _secret [32]byte) (*types.Transaction, error) {
	return _SwapCreator.Contract.Refund(&_SwapCreator.TransactOpts, _swap, _secret)
}

// SetReady is a paid mutator transaction binding the contract method 0xfcaf229c.
//
// Solidity: function setReady((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap) returns()
func (_SwapCreator *SwapCreatorTransactor) SetReady(opts *bind.TransactOpts, _swap SwapCreatorSwap) (*types.Transaction, error) {
	return _SwapCreator.contract.Transact(opts, "setReady", _swap)
}

// SetReady is a paid mutator transaction binding the contract method 0xfcaf229c.
//
// Solidity: function setReady((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap) returns()
func (_SwapCreator *SwapCreatorSession) SetReady(_swap SwapCreatorSwap) (*types.Transaction, error) {
	return _SwapCreator.Contract.SetReady(&_SwapCreator.TransactOpts, _swap)
}

// SetReady is a paid mutator transaction binding the contract method 0xfcaf229c.
//
// Solidity: function setReady((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap) returns()
func (_SwapCreator *SwapCreatorTransactorSession) SetReady(_swap SwapCreatorSwap) (*types.Transaction, error) {
	return _SwapCreator.Contract.SetReady(&_SwapCreator.TransactOpts, _swap)
}

// SwapCreatorClaimedIterator is returned from FilterClaimed and is used to iterate over the raw logs and unpacked data for Claimed events raised by the SwapCreator contract.
type SwapCreatorClaimedIterator struct {
	Event *SwapCreatorClaimed // Event containing the contract specifics and raw log

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
func (it *SwapCreatorClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapCreatorClaimed)
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
		it.Event = new(SwapCreatorClaimed)
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
func (it *SwapCreatorClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapCreatorClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapCreatorClaimed represents a Claimed event raised by the SwapCreator contract.
type SwapCreatorClaimed struct {
	SwapID [32]byte
	S      [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterClaimed is a free log retrieval operation binding the contract event 0x38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee.
//
// Solidity: event Claimed(bytes32 indexed swapID, bytes32 indexed s)
func (_SwapCreator *SwapCreatorFilterer) FilterClaimed(opts *bind.FilterOpts, swapID [][32]byte, s [][32]byte) (*SwapCreatorClaimedIterator, error) {

	var swapIDRule []interface{}
	for _, swapIDItem := range swapID {
		swapIDRule = append(swapIDRule, swapIDItem)
	}
	var sRule []interface{}
	for _, sItem := range s {
		sRule = append(sRule, sItem)
	}

	logs, sub, err := _SwapCreator.contract.FilterLogs(opts, "Claimed", swapIDRule, sRule)
	if err != nil {
		return nil, err
	}
	return &SwapCreatorClaimedIterator{contract: _SwapCreator.contract, event: "Claimed", logs: logs, sub: sub}, nil
}

// WatchClaimed is a free log subscription operation binding the contract event 0x38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee.
//
// Solidity: event Claimed(bytes32 indexed swapID, bytes32 indexed s)
func (_SwapCreator *SwapCreatorFilterer) WatchClaimed(opts *bind.WatchOpts, sink chan<- *SwapCreatorClaimed, swapID [][32]byte, s [][32]byte) (event.Subscription, error) {

	var swapIDRule []interface{}
	for _, swapIDItem := range swapID {
		swapIDRule = append(swapIDRule, swapIDItem)
	}
	var sRule []interface{}
	for _, sItem := range s {
		sRule = append(sRule, sItem)
	}

	logs, sub, err := _SwapCreator.contract.WatchLogs(opts, "Claimed", swapIDRule, sRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapCreatorClaimed)
				if err := _SwapCreator.contract.UnpackLog(event, "Claimed", log); err != nil {
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
// Solidity: event Claimed(bytes32 indexed swapID, bytes32 indexed s)
func (_SwapCreator *SwapCreatorFilterer) ParseClaimed(log types.Log) (*SwapCreatorClaimed, error) {
	event := new(SwapCreatorClaimed)
	if err := _SwapCreator.contract.UnpackLog(event, "Claimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapCreatorNewIterator is returned from FilterNew and is used to iterate over the raw logs and unpacked data for New events raised by the SwapCreator contract.
type SwapCreatorNewIterator struct {
	Event *SwapCreatorNew // Event containing the contract specifics and raw log

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
func (it *SwapCreatorNewIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapCreatorNew)
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
		it.Event = new(SwapCreatorNew)
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
func (it *SwapCreatorNewIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapCreatorNewIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapCreatorNew represents a New event raised by the SwapCreator contract.
type SwapCreatorNew struct {
	SwapID    [32]byte
	ClaimKey  [32]byte
	RefundKey [32]byte
	Timeout0  *big.Int
	Timeout1  *big.Int
	Asset     common.Address
	Value     *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterNew is a free log retrieval operation binding the contract event 0x91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56.
//
// Solidity: event New(bytes32 swapID, bytes32 claimKey, bytes32 refundKey, uint256 timeout0, uint256 timeout1, address asset, uint256 value)
func (_SwapCreator *SwapCreatorFilterer) FilterNew(opts *bind.FilterOpts) (*SwapCreatorNewIterator, error) {

	logs, sub, err := _SwapCreator.contract.FilterLogs(opts, "New")
	if err != nil {
		return nil, err
	}
	return &SwapCreatorNewIterator{contract: _SwapCreator.contract, event: "New", logs: logs, sub: sub}, nil
}

// WatchNew is a free log subscription operation binding the contract event 0x91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56.
//
// Solidity: event New(bytes32 swapID, bytes32 claimKey, bytes32 refundKey, uint256 timeout0, uint256 timeout1, address asset, uint256 value)
func (_SwapCreator *SwapCreatorFilterer) WatchNew(opts *bind.WatchOpts, sink chan<- *SwapCreatorNew) (event.Subscription, error) {

	logs, sub, err := _SwapCreator.contract.WatchLogs(opts, "New")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapCreatorNew)
				if err := _SwapCreator.contract.UnpackLog(event, "New", log); err != nil {
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

// ParseNew is a log parse operation binding the contract event 0x91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56.
//
// Solidity: event New(bytes32 swapID, bytes32 claimKey, bytes32 refundKey, uint256 timeout0, uint256 timeout1, address asset, uint256 value)
func (_SwapCreator *SwapCreatorFilterer) ParseNew(log types.Log) (*SwapCreatorNew, error) {
	event := new(SwapCreatorNew)
	if err := _SwapCreator.contract.UnpackLog(event, "New", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapCreatorReadyIterator is returned from FilterReady and is used to iterate over the raw logs and unpacked data for Ready events raised by the SwapCreator contract.
type SwapCreatorReadyIterator struct {
	Event *SwapCreatorReady // Event containing the contract specifics and raw log

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
func (it *SwapCreatorReadyIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapCreatorReady)
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
		it.Event = new(SwapCreatorReady)
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
func (it *SwapCreatorReadyIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapCreatorReadyIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapCreatorReady represents a Ready event raised by the SwapCreator contract.
type SwapCreatorReady struct {
	SwapID [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterReady is a free log retrieval operation binding the contract event 0x5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f.
//
// Solidity: event Ready(bytes32 indexed swapID)
func (_SwapCreator *SwapCreatorFilterer) FilterReady(opts *bind.FilterOpts, swapID [][32]byte) (*SwapCreatorReadyIterator, error) {

	var swapIDRule []interface{}
	for _, swapIDItem := range swapID {
		swapIDRule = append(swapIDRule, swapIDItem)
	}

	logs, sub, err := _SwapCreator.contract.FilterLogs(opts, "Ready", swapIDRule)
	if err != nil {
		return nil, err
	}
	return &SwapCreatorReadyIterator{contract: _SwapCreator.contract, event: "Ready", logs: logs, sub: sub}, nil
}

// WatchReady is a free log subscription operation binding the contract event 0x5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f.
//
// Solidity: event Ready(bytes32 indexed swapID)
func (_SwapCreator *SwapCreatorFilterer) WatchReady(opts *bind.WatchOpts, sink chan<- *SwapCreatorReady, swapID [][32]byte) (event.Subscription, error) {

	var swapIDRule []interface{}
	for _, swapIDItem := range swapID {
		swapIDRule = append(swapIDRule, swapIDItem)
	}

	logs, sub, err := _SwapCreator.contract.WatchLogs(opts, "Ready", swapIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapCreatorReady)
				if err := _SwapCreator.contract.UnpackLog(event, "Ready", log); err != nil {
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
// Solidity: event Ready(bytes32 indexed swapID)
func (_SwapCreator *SwapCreatorFilterer) ParseReady(log types.Log) (*SwapCreatorReady, error) {
	event := new(SwapCreatorReady)
	if err := _SwapCreator.contract.UnpackLog(event, "Ready", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapCreatorRefundedIterator is returned from FilterRefunded and is used to iterate over the raw logs and unpacked data for Refunded events raised by the SwapCreator contract.
type SwapCreatorRefundedIterator struct {
	Event *SwapCreatorRefunded // Event containing the contract specifics and raw log

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
func (it *SwapCreatorRefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapCreatorRefunded)
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
		it.Event = new(SwapCreatorRefunded)
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
func (it *SwapCreatorRefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapCreatorRefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapCreatorRefunded represents a Refunded event raised by the SwapCreator contract.
type SwapCreatorRefunded struct {
	SwapID [32]byte
	S      [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterRefunded is a free log retrieval operation binding the contract event 0x007c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f.
//
// Solidity: event Refunded(bytes32 indexed swapID, bytes32 indexed s)
func (_SwapCreator *SwapCreatorFilterer) FilterRefunded(opts *bind.FilterOpts, swapID [][32]byte, s [][32]byte) (*SwapCreatorRefundedIterator, error) {

	var swapIDRule []interface{}
	for _, swapIDItem := range swapID {
		swapIDRule = append(swapIDRule, swapIDItem)
	}
	var sRule []interface{}
	for _, sItem := range s {
		sRule = append(sRule, sItem)
	}

	logs, sub, err := _SwapCreator.contract.FilterLogs(opts, "Refunded", swapIDRule, sRule)
	if err != nil {
		return nil, err
	}
	return &SwapCreatorRefundedIterator{contract: _SwapCreator.contract, event: "Refunded", logs: logs, sub: sub}, nil
}

// WatchRefunded is a free log subscription operation binding the contract event 0x007c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f.
//
// Solidity: event Refunded(bytes32 indexed swapID, bytes32 indexed s)
func (_SwapCreator *SwapCreatorFilterer) WatchRefunded(opts *bind.WatchOpts, sink chan<- *SwapCreatorRefunded, swapID [][32]byte, s [][32]byte) (event.Subscription, error) {

	var swapIDRule []interface{}
	for _, swapIDItem := range swapID {
		swapIDRule = append(swapIDRule, swapIDItem)
	}
	var sRule []interface{}
	for _, sItem := range s {
		sRule = append(sRule, sItem)
	}

	logs, sub, err := _SwapCreator.contract.WatchLogs(opts, "Refunded", swapIDRule, sRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapCreatorRefunded)
				if err := _SwapCreator.contract.UnpackLog(event, "Refunded", log); err != nil {
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
// Solidity: event Refunded(bytes32 indexed swapID, bytes32 indexed s)
func (_SwapCreator *SwapCreatorFilterer) ParseRefunded(log types.Log) (*SwapCreatorRefunded, error) {
	event := new(SwapCreatorRefunded)
	if err := _SwapCreator.contract.UnpackLog(event, "Refunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
