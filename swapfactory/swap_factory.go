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
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"claimKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"refundKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout_0\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout_1\",\"type\":\"uint256\"}],\"name\":\"New\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"}],\"name\":\"Ready\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout_0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout_1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapFactory.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"scalar\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"qKeccak\",\"type\":\"uint256\"}],\"name\":\"mulVerify\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"addresspayable\",\"name\":\"_claimer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_timeoutDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"new_swap\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout_0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout_1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapFactory.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout_0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout_1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapFactory.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"}],\"name\":\"set_ready\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"swaps\",\"outputs\":[{\"internalType\":\"enumSwapFactory.Stage\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b506118f5806100206000396000f3fe6080604052600436106100555760003560e01c8063262cd8da1461005a5780633e7a7b55146100835780637069c7f3146100ac578063b32d1b4f146100d5578063d749b6c414610112578063eb84e7f214610142575b600080fd5b34801561006657600080fd5b50610081600480360381019061007c9190610ebf565b61017f565b005b34801561008f57600080fd5b506100aa60048036038101906100a59190610f01565b610434565b005b3480156100b857600080fd5b506100d360048036038101906100ce9190610ebf565b6105cd565b005b3480156100e157600080fd5b506100fc60048036038101906100f79190610f2f565b6108b9565b6040516101099190610f8a565b60405180910390f35b61012c60048036038101906101279190610fa5565b6109be565b604051610139919061102f565b60405180910390f35b34801561014e57600080fd5b506101696004803603810190610164919061104a565b610ba1565b60405161017691906110ee565b60405180910390f35b60008260405160200161019291906111d8565b604051602081830303815290604052805190602001209050600080600083815260200190815260200160002060009054906101000a900460ff1690506003808111156101e1576101e0611077565b5b8160038111156101f4576101f3611077565b5b1415801561022757506000600381111561021157610210611077565b5b81600381111561022457610223611077565b5b14155b610266576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161025d90611251565b60405180910390fd5b836000015173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146102d8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102cf906112e3565b60405180910390fd5b8360a001514210158061031f575083608001514210801561031e57506002600381111561030857610307611077565b5b81600381111561031b5761031a611077565b5b14155b5b61035e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161035590611375565b60405180910390fd5b61036c838560600151610bc1565b7e7c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f828460405161039c929190611395565b60405180910390a1836000015173ffffffffffffffffffffffffffffffffffffffff166108fc8560c001519081150290604051600060405180830381858888f193505050501580156103f2573d6000803e3d6000fd5b50600360008084815260200190815260200160002060006101000a81548160ff0219169083600381111561042957610428611077565b5b021790555050505050565b60008160405160200161044791906111d8565b6040516020818303038152906040528051906020012090506001600381111561047357610472611077565b5b60008083815260200190815260200160002060009054906101000a900460ff1660038111156104a5576104a4611077565b5b146104e5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104dc9061140a565b60405180910390fd5b3373ffffffffffffffffffffffffffffffffffffffff16826000015173ffffffffffffffffffffffffffffffffffffffff1614610557576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161054e9061149c565b60405180910390fd5b600260008083815260200190815260200160002060006101000a81548160ff0219169083600381111561058d5761058c611077565b5b02179055507f5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f816040516105c1919061102f565b60405180910390a15050565b6000826040516020016105e091906111d8565b604051602081830303815290604052805190602001209050600080600083815260200190815260200160002060009054906101000a900460ff16905060038081111561062f5761062e611077565b5b81600381111561064257610641611077565b5b1415801561067557506000600381111561065f5761065e611077565b5b81600381111561067257610671611077565b5b14155b6106b4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106ab90611251565b60405180910390fd5b836020015173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610726576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161071d90611508565b60405180910390fd5b83608001514210158061075d57506002600381111561074857610747611077565b5b81600381111561075b5761075a611077565b5b145b61079c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161079390611574565b60405180910390fd5b8360a0015142106107e2576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107d9906115e0565b60405180910390fd5b6107f0838560400151610bc1565b7f38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee8284604051610821929190611395565b60405180910390a1836020015173ffffffffffffffffffffffffffffffffffffffff166108fc8560c001519081150290604051600060405180830381858888f19350505050158015610877573d6000803e3d6000fd5b50600360008084815260200190815260200160002060006101000a81548160ff021916908360038111156108ae576108ad611077565b5b021790555050505050565b60008060016000601b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179860001b7ffffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd03641418061091557610914611600565b5b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798890960001b6040516000815260200160405260405161095894939291906116c9565b6020604051602081039080840390855afa15801561097a573d6000803e3d6000fd5b5050506020604051035190508073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161491505092915050565b60006109c8610c14565b33816000019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505084816020019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505086816040018181525050858160600181815250508342610a58919061173d565b816080018181525050600284610a6e9190611771565b42610a79919061173d565b8160a0018181525050348160c0018181525050828160e0018181525050600081604051602001610aa991906111d8565b60405160208183030381529060405280519060200120905060006003811115610ad557610ad4611077565b5b60008083815260200190815260200160002060009054906101000a900460ff166003811115610b0757610b06611077565b5b14610b1157600080fd5b7f8116b8ce401b5f8f3bb3b91fc2ac461b29ffe582eff877d50fb5a9f9e54306be81898985608001518660a00151604051610b509594939291906117da565b60405180910390a1600160008083815260200190815260200160002060006101000a81548160ff02191690836003811115610b8e57610b8d611077565b5b0217905550809250505095945050505050565b60006020528060005260406000206000915054906101000a900460ff1681565b610bd18260001c8260001c6108b9565b610c10576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610c079061189f565b60405180910390fd5b5050565b604051806101000160405280600073ffffffffffffffffffffffffffffffffffffffff168152602001600073ffffffffffffffffffffffffffffffffffffffff1681526020016000801916815260200160008019168152602001600081526020016000815260200160008152602001600081525090565b6000604051905090565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b610ce882610c9f565b810181811067ffffffffffffffff82111715610d0757610d06610cb0565b5b80604052505050565b6000610d1a610c8b565b9050610d268282610cdf565b919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610d5682610d2b565b9050919050565b610d6681610d4b565b8114610d7157600080fd5b50565b600081359050610d8381610d5d565b92915050565b6000819050919050565b610d9c81610d89565b8114610da757600080fd5b50565b600081359050610db981610d93565b92915050565b6000819050919050565b610dd281610dbf565b8114610ddd57600080fd5b50565b600081359050610def81610dc9565b92915050565b60006101008284031215610e0c57610e0b610c9a565b5b610e17610100610d10565b90506000610e2784828501610d74565b6000830152506020610e3b84828501610d74565b6020830152506040610e4f84828501610daa565b6040830152506060610e6384828501610daa565b6060830152506080610e7784828501610de0565b60808301525060a0610e8b84828501610de0565b60a08301525060c0610e9f84828501610de0565b60c08301525060e0610eb384828501610de0565b60e08301525092915050565b6000806101208385031215610ed757610ed6610c95565b5b6000610ee585828601610df5565b925050610100610ef785828601610daa565b9150509250929050565b60006101008284031215610f1857610f17610c95565b5b6000610f2684828501610df5565b91505092915050565b60008060408385031215610f4657610f45610c95565b5b6000610f5485828601610de0565b9250506020610f6585828601610de0565b9150509250929050565b60008115159050919050565b610f8481610f6f565b82525050565b6000602082019050610f9f6000830184610f7b565b92915050565b600080600080600060a08688031215610fc157610fc0610c95565b5b6000610fcf88828901610daa565b9550506020610fe088828901610daa565b9450506040610ff188828901610d74565b935050606061100288828901610de0565b925050608061101388828901610de0565b9150509295509295909350565b61102981610d89565b82525050565b60006020820190506110446000830184611020565b92915050565b6000602082840312156110605761105f610c95565b5b600061106e84828501610daa565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b600481106110b7576110b6611077565b5b50565b60008190506110c8826110a6565b919050565b60006110d8826110ba565b9050919050565b6110e8816110cd565b82525050565b600060208201905061110360008301846110df565b92915050565b61111281610d4b565b82525050565b61112181610d89565b82525050565b61113081610dbf565b82525050565b6101008201600082015161114d6000850182611109565b5060208201516111606020850182611109565b5060408201516111736040850182611118565b5060608201516111866060850182611118565b5060808201516111996080850182611127565b5060a08201516111ac60a0850182611127565b5060c08201516111bf60c0850182611127565b5060e08201516111d260e0850182611127565b50505050565b6000610100820190506111ee6000830184611136565b92915050565b600082825260208201905092915050565b7f7377617020697320616c726561647920636f6d706c6574656400000000000000600082015250565b600061123b6019836111f4565b915061124682611205565b602082019050919050565b6000602082019050818103600083015261126a8161122e565b9050919050565b7f726566756e64206d7573742062652063616c6c6564206279207468652073776160008201527f70206f776e657200000000000000000000000000000000000000000000000000602082015250565b60006112cd6027836111f4565b91506112d882611271565b604082019050919050565b600060208201905081810360008301526112fc816112c0565b9050919050565b7f697427732074686520636f756e74657270617274792773207475726e2c20756e60008201527f61626c6520746f20726566756e642c2074727920616761696e206c6174657200602082015250565b600061135f603f836111f4565b915061136a82611303565b604082019050919050565b6000602082019050818103600083015261138e81611352565b9050919050565b60006040820190506113aa6000830185611020565b6113b76020830184611020565b9392505050565b7f73776170206973206e6f7420696e2050454e44494e4720737461746500000000600082015250565b60006113f4601c836111f4565b91506113ff826113be565b602082019050919050565b60006020820190508181036000830152611423816113e7565b9050919050565b7f6f6e6c79207468652073776170206f776e65722063616e2063616c6c2073657460008201527f5f72656164790000000000000000000000000000000000000000000000000000602082015250565b60006114866026836111f4565b91506114918261142a565b604082019050919050565b600060208201905081810360008301526114b581611479565b9050919050565b7f6f6e6c7920636c61696d65722063616e20636c61696d21000000000000000000600082015250565b60006114f26017836111f4565b91506114fd826114bc565b602082019050919050565b60006020820190508181036000830152611521816114e5565b9050919050565b7f746f6f206561726c7920746f20636c61696d2100000000000000000000000000600082015250565b600061155e6013836111f4565b915061156982611528565b602082019050919050565b6000602082019050818103600083015261158d81611551565b9050919050565b7f746f6f206c61746520746f20636c61696d210000000000000000000000000000600082015250565b60006115ca6012836111f4565b91506115d582611594565b602082019050919050565b600060208201905081810360008301526115f9816115bd565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000819050919050565b60008160001b9050919050565b600061166161165c6116578461162f565b611639565b610d89565b9050919050565b61167181611646565b82525050565b6000819050919050565b600060ff82169050919050565b6000819050919050565b60006116b36116ae6116a984611677565b61168e565b611681565b9050919050565b6116c381611698565b82525050565b60006080820190506116de6000830187611668565b6116eb60208301866116ba565b6116f86040830185611020565b6117056060830184611020565b95945050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061174882610dbf565b915061175383610dbf565b925082820190508082111561176b5761176a61170e565b5b92915050565b600061177c82610dbf565b915061178783610dbf565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff04831182151516156117c0576117bf61170e565b5b828202905092915050565b6117d481610dbf565b82525050565b600060a0820190506117ef6000830188611020565b6117fc6020830187611020565b6118096040830186611020565b61181660608301856117cb565b61182360808301846117cb565b9695505050505050565b7f70726f76696465642073656372657420646f6573206e6f74206d61746368207460008201527f6865206578706563746564207075626c6963206b657900000000000000000000602082015250565b60006118896036836111f4565b91506118948261182d565b604082019050919050565b600060208201905081810360008301526118b88161187c565b905091905056fea26469706673582212203e7e1a4d0225241f38d53e19113d8f73f6dfa0a394f8e97d2e909300f7433b0764736f6c63430008100033",
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
