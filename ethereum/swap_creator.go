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
	RelayerHash [32]byte
	SwapCreator common.Address
}

// SwapCreatorSwap is an auto generated low-level Go binding around an user-defined struct.
type SwapCreatorSwap struct {
	Owner            common.Address
	Claimer          common.Address
	ClaimCommitment  [32]byte
	RefundCommitment [32]byte
	Timeout1         *big.Int
	Timeout2         *big.Int
	Asset            common.Address
	Value            *big.Int
	Nonce            *big.Int
}

// SwapCreatorMetaData contains all meta data concerning the SwapCreator contract.
var SwapCreatorMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"InvalidClaimer\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidContractAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidRelayerAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSecret\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSwap\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSwapKey\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidTimeout\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidValue\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotTimeToRefund\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OnlySwapClaimer\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OnlySwapOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SwapAlreadyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SwapCompleted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SwapNotPending\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TooEarlyToClaim\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TooLateToClaim\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroValue\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"claimKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"refundKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout2\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"New\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"}],\"name\":\"Ready\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"claimCommitment\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"refundCommitment\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout2\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapCreator.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_secret\",\"type\":\"bytes32\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"claimCommitment\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"refundCommitment\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout2\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapCreator.Swap\",\"name\":\"swap\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"relayerHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"swapCreator\",\"type\":\"address\"}],\"internalType\":\"structSwapCreator.RelaySwap\",\"name\":\"_relaySwap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_secret\",\"type\":\"bytes32\"},{\"internalType\":\"addresspayable\",\"name\":\"_relayer\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"_salt\",\"type\":\"uint32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"claimRelayer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"scalar\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"qKeccak\",\"type\":\"uint256\"}],\"name\":\"mulVerify\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_claimCommitment\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_refundCommitment\",\"type\":\"bytes32\"},{\"internalType\":\"addresspayable\",\"name\":\"_claimer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_timeoutDuration1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_timeoutDuration2\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"newSwap\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"claimCommitment\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"refundCommitment\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout2\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapCreator.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_secret\",\"type\":\"bytes32\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"claimCommitment\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"refundCommitment\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout2\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapCreator.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"}],\"name\":\"setReady\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"swaps\",\"outputs\":[{\"internalType\":\"enumSwapCreator.Stage\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b506113fc8061001d5f395ff3fe60806040526004361061006e575f3560e01c8063b32d1b4f1161004c578063b32d1b4f146100d1578063c41e46cf14610105578063eb84e7f214610126578063fcaf229c14610161575f80fd5b80631e6c5acc146100725780635cb969161461009357806387065c49146100b2575b5f80fd5b34801561007d575f80fd5b5061009161008c366004611040565b610180565b005b34801561009e575f80fd5b506100916100ad366004611040565b610362565b3480156100bd575f80fd5b506100916100cc36600461108e565b610425565b3480156100dc575f80fd5b506100f06100eb366004611146565b610688565b60405190151581526020015b60405180910390f35b610118610113366004611166565b610754565b6040519081526020016100fc565b348015610131575f80fd5b506101546101403660046111d2565b5f6020819052908152604090205460ff1681565b6040516100fc91906111fd565b34801561016c575f80fd5b5061009161017b366004611223565b6109cc565b5f8260405160200161019291906112ad565b60408051601f1981840301815291815281516020928301205f818152928390529082205490925060ff16908160038111156101cf576101cf6111e9565b036101ed57604051631115766760e01b815260040160405180910390fd5b6003816003811115610201576102016111e9565b0361021f5760405163066916a960e01b815260040160405180910390fd5b83516001600160a01b031633146102495760405163148ca24360e11b815260040160405180910390fd5b8360a001514210801561027a5750836080015142118061027a57506002816003811115610278576102786111e9565b145b15610298576040516332a1860f60e11b815260040160405180910390fd5b6102a6838560600151610aa7565b604051839083907e7c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f905f90a35f828152602081905260409020805460ff1916600317905560c08401516001600160a01b031661033b57835160e08501516040516001600160a01b039092169181156108fc0291905f818181858888f19350505050158015610335573d5f803e3d5ffd5b5061035c565b835160e085015160c086015161035c926001600160a01b0390911691610ace565b50505050565b81602001516001600160a01b0316336001600160a01b03161461039857604051633471640960e11b815260040160405180910390fd5b6103a28282610b31565b60c08201516001600160a01b03166103f75781602001516001600160a01b03166108fc8360e0015190811502906040515f60405180830381858888f193505050501580156103f2573d5f803e3d5ffd5b505050565b61042182602001518360e001518460c001516001600160a01b0316610ace9092919063ffffffff16565b5050565b5f60018860405160200161043991906112bc565b60408051601f1981840301815282825280516020918201205f84529083018083525260ff871690820152606081018590526080810184905260a0016020604051602081039080840390855afa158015610494573d5f803e3d5ffd5b505050602060405103519050875f0151602001516001600160a01b0316816001600160a01b0316146104d957604051638baa579f60e01b815260040160405180910390fd5b87606001516001600160a01b0316306001600160a01b03161461050f5760405163a710429d60e01b815260040160405180910390fd5b60408089015190516bffffffffffffffffffffffff19606089901b1660208201526001600160e01b031960e088901b166034820152603801604051602081830303815290604052805190602001201461057b5760405163fe16c3c560e01b815260040160405180910390fd5b87516105879088610b31565b875160c001516001600160a01b031661062757875f0151602001516001600160a01b03166108fc89602001518a5f015160e001516105c59190611312565b6040518115909202915f818181858888f193505050501580156105ea573d5f803e3d5ffd5b5060208801516040516001600160a01b0388169180156108fc02915f818181858888f19350505050158015610621573d5f803e3d5ffd5b5061067e565b8751602080820151908a015160e09092015161065c9261064691611312565b8a5160c001516001600160a01b03169190610ace565b6020880151885160c0015161067e916001600160a01b03909116908890610ace565b5050505050505050565b5f80600181601b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179870014551231950b75fc4402da1732fc9bebe197f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f817988909604080515f8152602081018083529590955260ff909316928401929092526060830152608082015260a0016020604051602081039080840390855afa158015610731573d5f803e3d5ffd5b5050604051601f1901516001600160a01b03858116911614925050505b92915050565b5f825f0361077557604051637c946ed760e01b815260040160405180910390fd5b6001600160a01b0384166107a8573483146107a357604051632a9ffab760e21b815260040160405180910390fd5b6107bd565b6107bd6001600160a01b038516333086610c8e565b8815806107c8575087155b156107e657604051631bc61bed60e11b815260040160405180910390fd5b6001600160a01b03871661080c576040516208978560e71b815260040160405180910390fd5b851580610817575084155b1561083557604051631ffb86f160e21b815260040160405180910390fd5b5f604051806101200160405280336001600160a01b03168152602001896001600160a01b031681526020018b81526020018a815260200188426108789190611325565b8152602001876108888a42611325565b6108929190611325565b8152602001866001600160a01b031681526020018581526020018481525090505f816040516020016108c491906112ad565b60408051601f19818403018152919052805160209091012090505f808281526020819052604090205460ff166003811115610901576109016111e9565b1461091f576040516339a2986760e11b815260040160405180910390fd5b7f91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56818c8c85608001518660a001518760c001518860e0015160405161099e979695949392919096875260208701959095526040860193909352606085019190915260808401526001600160a01b031660a083015260c082015260e00190565b60405180910390a15f818152602081905260409020805460ff191660011790559a9950505050505050505050565b5f816040516020016109de91906112ad565b60408051601f198184030181529190528051602090910120905060015f8281526020819052604090205460ff166003811115610a1c57610a1c6111e9565b14610a3a57604051630fe0fb5160e11b815260040160405180910390fd5b81516001600160a01b03163314610a645760405163148ca24360e11b815260040160405180910390fd5b5f81815260208190526040808220805460ff191660021790555182917f5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f91a25050565b610ab18282610688565b6104215760405163abab6bd760e01b815260040160405180910390fd5b6040516001600160a01b0383166024820152604481018290526103f290849063a9059cbb60e01b906064015b60408051601f198184030181529190526020810180516001600160e01b03166001600160e01b031990931692909217909152610cc6565b5f82604051602001610b4391906112ad565b60408051601f1981840301815291815281516020928301205f818152928390529082205490925060ff1690816003811115610b8057610b806111e9565b03610b9e57604051631115766760e01b815260040160405180910390fd5b6003816003811115610bb257610bb26111e9565b03610bd05760405163066916a960e01b815260040160405180910390fd5b836080015142108015610bf557506002816003811115610bf257610bf26111e9565b14155b15610c135760405163d71d60b560e01b815260040160405180910390fd5b8360a001514210610c375760405163497df9d160e01b815260040160405180910390fd5b610c45838560400151610aa7565b604051839083907f38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee905f90a3505f908152602081905260409020805460ff191660031790555050565b6040516001600160a01b038085166024830152831660448201526064810182905261035c9085906323b872dd60e01b90608401610afa565b5f610d1a826040518060400160405280602081526020017f5361666545524332303a206c6f772d6c6576656c2063616c6c206661696c6564815250856001600160a01b0316610d9e9092919063ffffffff16565b905080515f1480610d3a575080806020019051810190610d3a9190611338565b6103f25760405162461bcd60e51b815260206004820152602a60248201527f5361666545524332303a204552433230206f7065726174696f6e20646964206e6044820152691bdd081cdd58d8d9595960b21b60648201526084015b60405180910390fd5b6060610dac84845f85610db4565b949350505050565b606082471015610e155760405162461bcd60e51b815260206004820152602660248201527f416464726573733a20696e73756666696369656e742062616c616e636520666f6044820152651c8818d85b1b60d21b6064820152608401610d95565b5f80866001600160a01b03168587604051610e309190611379565b5f6040518083038185875af1925050503d805f8114610e6a576040519150601f19603f3d011682016040523d82523d5f602084013e610e6f565b606091505b5091509150610e8087838387610e8b565b979650505050505050565b60608315610ef95782515f03610ef2576001600160a01b0385163b610ef25760405162461bcd60e51b815260206004820152601d60248201527f416464726573733a2063616c6c20746f206e6f6e2d636f6e74726163740000006044820152606401610d95565b5081610dac565b610dac8383815115610f0e5781518083602001fd5b8060405162461bcd60e51b8152600401610d959190611394565b604051610120810167ffffffffffffffff81118282101715610f5857634e487b7160e01b5f52604160045260245ffd5b60405290565b6040516080810167ffffffffffffffff81118282101715610f5857634e487b7160e01b5f52604160045260245ffd5b6001600160a01b0381168114610fa1575f80fd5b50565b8035610faf81610f8d565b919050565b5f6101208284031215610fc5575f80fd5b610fcd610f28565b9050610fd882610fa4565b8152610fe660208301610fa4565b602082015260408201356040820152606082013560608201526080820135608082015260a082013560a082015261101f60c08301610fa4565b60c082015260e082013560e082015261010080830135818301525092915050565b5f806101408385031215611052575f80fd5b61105c8484610fb4565b94610120939093013593505050565b803563ffffffff81168114610faf575f80fd5b803560ff81168114610faf575f80fd5b5f805f805f805f8789036102408112156110a6575f80fd5b610180808212156110b5575f80fd5b6110bd610f5e565b91506110c98b8b610fb4565b82526101208a013560208301526101408a013560408301526101608a01356110f081610f8d565b6060830152909750880135955061110a6101a08901610fa4565b94506111196101c0890161106b565b93506111286101e0890161107e565b92506102008801359150610220880135905092959891949750929550565b5f8060408385031215611157575f80fd5b50508035926020909101359150565b5f805f805f805f80610100898b03121561117e575f80fd5b8835975060208901359650604089013561119781610f8d565b9550606089013594506080890135935060a08901356111b581610f8d565b979a969950949793969295929450505060c08201359160e0013590565b5f602082840312156111e2575f80fd5b5035919050565b634e487b7160e01b5f52602160045260245ffd5b602081016004831061121d57634e487b7160e01b5f52602160045260245ffd5b91905290565b5f6101208284031215611234575f80fd5b61123e8383610fb4565b9392505050565b60018060a01b0380825116835280602083015116602084015260408201516040840152606082015160608401526080820151608084015260a082015160a08401528060c08301511660c08401525060e081015160e08301526101008082015181840152505050565b610120810161074e8284611245565b5f610180820190506112cf828451611245565b602083015161012083015260408301516101408301526060909201516001600160a01b03166101609091015290565b634e487b7160e01b5f52601160045260245ffd5b8181038181111561074e5761074e6112fe565b8082018082111561074e5761074e6112fe565b5f60208284031215611348575f80fd5b8151801515811461123e575f80fd5b5f5b83811015611371578181015183820152602001611359565b50505f910152565b5f825161138a818460208701611357565b9190910192915050565b602081525f82518060208401526113b2816040850160208701611357565b601f01601f1916919091016040019291505056fea264697066735822122058723d6f94b6ae0fd67bfece90eee43db933ef07d0eecef13276b9cf2f7a7b7e64736f6c63430008140033",
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

// ClaimRelayer is a paid mutator transaction binding the contract method 0x87065c49.
//
// Solidity: function claimRelayer(((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256),uint256,bytes32,address) _relaySwap, bytes32 _secret, address _relayer, uint32 _salt, uint8 v, bytes32 r, bytes32 s) returns()
func (_SwapCreator *SwapCreatorTransactor) ClaimRelayer(opts *bind.TransactOpts, _relaySwap SwapCreatorRelaySwap, _secret [32]byte, _relayer common.Address, _salt uint32, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _SwapCreator.contract.Transact(opts, "claimRelayer", _relaySwap, _secret, _relayer, _salt, v, r, s)
}

// ClaimRelayer is a paid mutator transaction binding the contract method 0x87065c49.
//
// Solidity: function claimRelayer(((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256),uint256,bytes32,address) _relaySwap, bytes32 _secret, address _relayer, uint32 _salt, uint8 v, bytes32 r, bytes32 s) returns()
func (_SwapCreator *SwapCreatorSession) ClaimRelayer(_relaySwap SwapCreatorRelaySwap, _secret [32]byte, _relayer common.Address, _salt uint32, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _SwapCreator.Contract.ClaimRelayer(&_SwapCreator.TransactOpts, _relaySwap, _secret, _relayer, _salt, v, r, s)
}

// ClaimRelayer is a paid mutator transaction binding the contract method 0x87065c49.
//
// Solidity: function claimRelayer(((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256),uint256,bytes32,address) _relaySwap, bytes32 _secret, address _relayer, uint32 _salt, uint8 v, bytes32 r, bytes32 s) returns()
func (_SwapCreator *SwapCreatorTransactorSession) ClaimRelayer(_relaySwap SwapCreatorRelaySwap, _secret [32]byte, _relayer common.Address, _salt uint32, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _SwapCreator.Contract.ClaimRelayer(&_SwapCreator.TransactOpts, _relaySwap, _secret, _relayer, _salt, v, r, s)
}

// NewSwap is a paid mutator transaction binding the contract method 0xc41e46cf.
//
// Solidity: function newSwap(bytes32 _claimCommitment, bytes32 _refundCommitment, address _claimer, uint256 _timeoutDuration1, uint256 _timeoutDuration2, address _asset, uint256 _value, uint256 _nonce) payable returns(bytes32)
func (_SwapCreator *SwapCreatorTransactor) NewSwap(opts *bind.TransactOpts, _claimCommitment [32]byte, _refundCommitment [32]byte, _claimer common.Address, _timeoutDuration1 *big.Int, _timeoutDuration2 *big.Int, _asset common.Address, _value *big.Int, _nonce *big.Int) (*types.Transaction, error) {
	return _SwapCreator.contract.Transact(opts, "newSwap", _claimCommitment, _refundCommitment, _claimer, _timeoutDuration1, _timeoutDuration2, _asset, _value, _nonce)
}

// NewSwap is a paid mutator transaction binding the contract method 0xc41e46cf.
//
// Solidity: function newSwap(bytes32 _claimCommitment, bytes32 _refundCommitment, address _claimer, uint256 _timeoutDuration1, uint256 _timeoutDuration2, address _asset, uint256 _value, uint256 _nonce) payable returns(bytes32)
func (_SwapCreator *SwapCreatorSession) NewSwap(_claimCommitment [32]byte, _refundCommitment [32]byte, _claimer common.Address, _timeoutDuration1 *big.Int, _timeoutDuration2 *big.Int, _asset common.Address, _value *big.Int, _nonce *big.Int) (*types.Transaction, error) {
	return _SwapCreator.Contract.NewSwap(&_SwapCreator.TransactOpts, _claimCommitment, _refundCommitment, _claimer, _timeoutDuration1, _timeoutDuration2, _asset, _value, _nonce)
}

// NewSwap is a paid mutator transaction binding the contract method 0xc41e46cf.
//
// Solidity: function newSwap(bytes32 _claimCommitment, bytes32 _refundCommitment, address _claimer, uint256 _timeoutDuration1, uint256 _timeoutDuration2, address _asset, uint256 _value, uint256 _nonce) payable returns(bytes32)
func (_SwapCreator *SwapCreatorTransactorSession) NewSwap(_claimCommitment [32]byte, _refundCommitment [32]byte, _claimer common.Address, _timeoutDuration1 *big.Int, _timeoutDuration2 *big.Int, _asset common.Address, _value *big.Int, _nonce *big.Int) (*types.Transaction, error) {
	return _SwapCreator.Contract.NewSwap(&_SwapCreator.TransactOpts, _claimCommitment, _refundCommitment, _claimer, _timeoutDuration1, _timeoutDuration2, _asset, _value, _nonce)
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
	Timeout1  *big.Int
	Timeout2  *big.Int
	Asset     common.Address
	Value     *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterNew is a free log retrieval operation binding the contract event 0x91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56.
//
// Solidity: event New(bytes32 swapID, bytes32 claimKey, bytes32 refundKey, uint256 timeout1, uint256 timeout2, address asset, uint256 value)
func (_SwapCreator *SwapCreatorFilterer) FilterNew(opts *bind.FilterOpts) (*SwapCreatorNewIterator, error) {

	logs, sub, err := _SwapCreator.contract.FilterLogs(opts, "New")
	if err != nil {
		return nil, err
	}
	return &SwapCreatorNewIterator{contract: _SwapCreator.contract, event: "New", logs: logs, sub: sub}, nil
}

// WatchNew is a free log subscription operation binding the contract event 0x91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56.
//
// Solidity: event New(bytes32 swapID, bytes32 claimKey, bytes32 refundKey, uint256 timeout1, uint256 timeout2, address asset, uint256 value)
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
// Solidity: event New(bytes32 swapID, bytes32 claimKey, bytes32 refundKey, uint256 timeout1, uint256 timeout2, address asset, uint256 value)
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
