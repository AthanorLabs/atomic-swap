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
)

// SwapFactorySwap is an auto generated low-level Go binding around an user-defined struct.
type SwapFactorySwap struct {
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

// SwapFactoryMetaData contains all meta data concerning the SwapFactory contract.
var SwapFactoryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"trustedForwarder\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"claimKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"refundKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"New\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"}],\"name\":\"Ready\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"_trustedForwarder\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapFactory.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapFactory.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"}],\"name\":\"claimRelayer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"forwarder\",\"type\":\"address\"}],\"name\":\"isTrustedForwarder\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"scalar\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"qKeccak\",\"type\":\"uint256\"}],\"name\":\"mulVerify\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"addresspayable\",\"name\":\"_claimer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_timeoutDuration\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"newSwap\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapFactory.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapFactory.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"}],\"name\":\"setReady\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"swaps\",\"outputs\":[{\"internalType\":\"enumSwapFactory.Stage\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60a06040523480156200001157600080fd5b506040516200293a3803806200293a8339818101604052810190620000379190620000de565b808073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff1681525050505062000110565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000620000a68262000079565b9050919050565b620000b88162000099565b8114620000c457600080fd5b50565b600081519050620000d881620000ad565b92915050565b600060208284031215620000f757620000f662000074565b5b60006200010784828501620000c7565b91505092915050565b60805161280762000133600039600081816105ff015261062501526128076000f3fe6080604052600436106100865760003560e01c806373e4771c1161005957806373e4771c14610145578063aa0f87251461016e578063b32d1b4f1461019e578063eb84e7f2146101db578063fcaf229c1461021857610086565b80631e6c5acc1461008b57806356c022bb146100b4578063572b6c05146100df5780635cb969161461011c575b600080fd5b34801561009757600080fd5b506100b260048036038101906100ad91906117be565b610241565b005b3480156100c057600080fd5b506100c96105fd565b6040516100d6919061180f565b60405180910390f35b3480156100eb57600080fd5b506101066004803603810190610101919061182a565b610621565b6040516101139190611872565b60405180910390f35b34801561012857600080fd5b50610143600480360381019061013e91906117be565b610679565b005b34801561015157600080fd5b5061016c6004803603810190610167919061188d565b6107dd565b005b610188600480360381019061018391906118e3565b610aa8565b6040516101959190611994565b60405180910390f35b3480156101aa57600080fd5b506101c560048036038101906101c091906119af565b610e7e565b6040516101d29190611872565b60405180910390f35b3480156101e757600080fd5b5061020260048036038101906101fd91906119ef565b610f83565b60405161020f9190611a93565b60405180910390f35b34801561022457600080fd5b5061023f600480360381019061023a9190611aae565b610fa3565b005b6000826040516020016102549190611bcf565b604051602081830303815290604052805190602001209050600080600083815260200190815260200160002060009054906101000a900460ff1690506003808111156102a3576102a2611a1c565b5b8160038111156102b6576102b5611a1c565b5b141580156102e95750600060038111156102d3576102d2611a1c565b5b8160038111156102e6576102e5611a1c565b5b14155b610328576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161031f90611c48565b60405180910390fd5b836000015173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461039a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161039190611cda565b60405180910390fd5b8360a00151421015806103e157508360800151421080156103e05750600260038111156103ca576103c9611a1c565b5b8160038111156103dd576103dc611a1c565b5b14155b5b610420576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161041790611d6c565b60405180910390fd5b61042e83856060015161113c565b7e7c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f828460405161045e929190611d8c565b60405180910390a1600360008084815260200190815260200160002060006101000a81548160ff0219169083600381111561049c5761049b611a1c565b5b0217905550600073ffffffffffffffffffffffffffffffffffffffff168460c0015173ffffffffffffffffffffffffffffffffffffffff160361052d57836000015173ffffffffffffffffffffffffffffffffffffffff166108fc8560e001519081150290604051600060405180830381858888f19350505050158015610527573d6000803e3d6000fd5b506105f7565b8360c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb85600001518660e001516040518363ffffffff1660e01b8152600401610574929190611e23565b6020604051808303816000875af1158015610593573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105b79190611e78565b6105f6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105ed90611ef1565b60405180910390fd5b5b50505050565b7f000000000000000000000000000000000000000000000000000000000000000081565b60007f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16149050919050565b610683828261118f565b600073ffffffffffffffffffffffffffffffffffffffff168260c0015173ffffffffffffffffffffffffffffffffffffffff160361070f57816020015173ffffffffffffffffffffffffffffffffffffffff166108fc8360e001519081150290604051600060405180830381858888f19350505050158015610709573d6000803e3d6000fd5b506107d9565b8160c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb83602001518460e001516040518363ffffffff1660e01b8152600401610756929190611e23565b6020604051808303816000875af1158015610775573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107999190611e78565b6107d8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107cf90611ef1565b60405180910390fd5b5b5050565b6107e633610621565b610825576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161081c90611f83565b60405180910390fd5b61082f838361118f565b600073ffffffffffffffffffffffffffffffffffffffff168360c0015173ffffffffffffffffffffffffffffffffffffffff160361090d57826020015173ffffffffffffffffffffffffffffffffffffffff166108fc828560e001516108959190611fd2565b9081150290604051600060405180830381858888f193505050501580156108c0573d6000803e3d6000fd5b503273ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015610907573d6000803e3d6000fd5b50610aa3565b8260c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb8460200151838660e001516109429190611fd2565b6040518363ffffffff1660e01b815260040161095f929190611e23565b6020604051808303816000875af115801561097e573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109a29190611e78565b6109e1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109d890612052565b60405180910390fd5b8260c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb32836040518363ffffffff1660e01b8152600401610a20929190612072565b6020604051808303816000875af1158015610a3f573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610a639190611e78565b610aa2576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610a99906120e7565b60405180910390fd5b5b505050565b60008073ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610b2057348314610b1f576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b1690612179565b60405180910390fd5b5b610b286114a2565b33816000019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505086816020019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505088816040018181525050878160600181815250508542610bb89190612199565b816080018181525050600286610bce91906121cd565b42610bd99190612199565b8160a0018181525050848160c0019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff1681525050838160e00181815250508281610100018181525050600081604051602001610c429190611bcf565b60405160208183030381529060405280519060200120905060006003811115610c6e57610c6d611a1c565b5b60008083815260200190815260200160002060009054906101000a900460ff166003811115610ca057610c9f611a1c565b5b14610ce0576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610cd79061225b565b60405180910390fd5b7f91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56818b8b85608001518660a001518760c001518860e00151604051610d2b979695949392919061227b565b60405180910390a1600160008083815260200190815260200160002060006101000a81548160ff02191690836003811115610d6957610d68611a1c565b5b0217905550600073ffffffffffffffffffffffffffffffffffffffff168260c0015173ffffffffffffffffffffffffffffffffffffffff1614610e6e578160c0015173ffffffffffffffffffffffffffffffffffffffff166323b872dd33308560e001516040518463ffffffff1660e01b8152600401610deb939291906122ea565b6020604051808303816000875af1158015610e0a573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610e2e9190611e78565b610e6d576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610e649061236d565b60405180910390fd5b5b8092505050979650505050505050565b60008060016000601b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179860001b7ffffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd036414180610eda57610ed961238d565b5b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798890960001b60405160008152602001604052604051610f1d949392919061244c565b6020604051602081039080840390855afa158015610f3f573d6000803e3d6000fd5b5050506020604051035190508073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161491505092915050565b60006020528060005260406000206000915054906101000a900460ff1681565b600081604051602001610fb69190611bcf565b60405160208183030381529060405280519060200120905060016003811115610fe257610fe1611a1c565b5b60008083815260200190815260200160002060009054906101000a900460ff16600381111561101457611013611a1c565b5b14611054576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161104b906124dd565b60405180910390fd5b3373ffffffffffffffffffffffffffffffffffffffff16826000015173ffffffffffffffffffffffffffffffffffffffff16146110c6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016110bd9061256f565b60405180910390fd5b600260008083815260200190815260200160002060006101000a81548160ff021916908360038111156110fc576110fb611a1c565b5b02179055507f5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f816040516111309190611994565b60405180910390a15050565b61114c8260001c8260001c610e7e565b61118b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161118290612601565b60405180910390fd5b5050565b6000826040516020016111a29190611bcf565b604051602081830303815290604052805190602001209050600080600083815260200190815260200160002060009054906101000a900460ff169050600060038111156111f2576111f1611a1c565b5b81600381111561120557611204611a1c565b5b03611245576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161123c9061266d565b60405180910390fd5b60038081111561125857611257611a1c565b5b81600381111561126b5761126a611a1c565b5b036112ab576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016112a290611c48565b60405180910390fd5b836020015173ffffffffffffffffffffffffffffffffffffffff166112ce611468565b73ffffffffffffffffffffffffffffffffffffffff1614611324576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161131b906126d9565b60405180910390fd5b83608001514210158061135b57506002600381111561134657611345611a1c565b5b81600381111561135957611358611a1c565b5b145b61139a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161139190612745565b60405180910390fd5b8360a0015142106113e0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016113d7906127b1565b60405180910390fd5b6113ee83856040015161113c565b7f38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee828460405161141f929190611d8c565b60405180910390a1600360008084815260200190815260200160002060006101000a81548160ff0219169083600381111561145d5761145c611a1c565b5b021790555050505050565b600061147333610621565b1561148757601436033560601c9050611496565b61148f61149a565b9050611497565b5b90565b600033905090565b604051806101200160405280600073ffffffffffffffffffffffffffffffffffffffff168152602001600073ffffffffffffffffffffffffffffffffffffffff16815260200160008019168152602001600080191681526020016000815260200160008152602001600073ffffffffffffffffffffffffffffffffffffffff16815260200160008152602001600081525090565b6000604051905090565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6115938261154a565b810181811067ffffffffffffffff821117156115b2576115b161155b565b5b80604052505050565b60006115c5611536565b90506115d1828261158a565b919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000611601826115d6565b9050919050565b611611816115f6565b811461161c57600080fd5b50565b60008135905061162e81611608565b92915050565b6000819050919050565b61164781611634565b811461165257600080fd5b50565b6000813590506116648161163e565b92915050565b6000819050919050565b61167d8161166a565b811461168857600080fd5b50565b60008135905061169a81611674565b92915050565b60006116ab826115d6565b9050919050565b6116bb816116a0565b81146116c657600080fd5b50565b6000813590506116d8816116b2565b92915050565b600061012082840312156116f5576116f4611545565b5b6117006101206115bb565b905060006117108482850161161f565b60008301525060206117248482850161161f565b602083015250604061173884828501611655565b604083015250606061174c84828501611655565b60608301525060806117608482850161168b565b60808301525060a06117748482850161168b565b60a08301525060c0611788848285016116c9565b60c08301525060e061179c8482850161168b565b60e0830152506101006117b18482850161168b565b6101008301525092915050565b60008061014083850312156117d6576117d5611540565b5b60006117e4858286016116de565b9250506101206117f685828601611655565b9150509250929050565b611809816116a0565b82525050565b60006020820190506118246000830184611800565b92915050565b6000602082840312156118405761183f611540565b5b600061184e848285016116c9565b91505092915050565b60008115159050919050565b61186c81611857565b82525050565b60006020820190506118876000830184611863565b92915050565b600080600061016084860312156118a7576118a6611540565b5b60006118b5868287016116de565b9350506101206118c786828701611655565b9250506101406118d98682870161168b565b9150509250925092565b600080600080600080600060e0888a03121561190257611901611540565b5b60006119108a828b01611655565b97505060206119218a828b01611655565b96505060406119328a828b0161161f565b95505060606119438a828b0161168b565b94505060806119548a828b016116c9565b93505060a06119658a828b0161168b565b92505060c06119768a828b0161168b565b91505092959891949750929550565b61198e81611634565b82525050565b60006020820190506119a96000830184611985565b92915050565b600080604083850312156119c6576119c5611540565b5b60006119d48582860161168b565b92505060206119e58582860161168b565b9150509250929050565b600060208284031215611a0557611a04611540565b5b6000611a1384828501611655565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b60048110611a5c57611a5b611a1c565b5b50565b6000819050611a6d82611a4b565b919050565b6000611a7d82611a5f565b9050919050565b611a8d81611a72565b82525050565b6000602082019050611aa86000830184611a84565b92915050565b60006101208284031215611ac557611ac4611540565b5b6000611ad3848285016116de565b91505092915050565b611ae5816115f6565b82525050565b611af481611634565b82525050565b611b038161166a565b82525050565b611b12816116a0565b82525050565b61012082016000820151611b2f6000850182611adc565b506020820151611b426020850182611adc565b506040820151611b556040850182611aeb565b506060820151611b686060850182611aeb565b506080820151611b7b6080850182611afa565b5060a0820151611b8e60a0850182611afa565b5060c0820151611ba160c0850182611b09565b5060e0820151611bb460e0850182611afa565b50610100820151611bc9610100850182611afa565b50505050565b600061012082019050611be56000830184611b18565b92915050565b600082825260208201905092915050565b7f7377617020697320616c726561647920636f6d706c6574656400000000000000600082015250565b6000611c32601983611beb565b9150611c3d82611bfc565b602082019050919050565b60006020820190508181036000830152611c6181611c25565b9050919050565b7f726566756e64206d7573742062652063616c6c6564206279207468652073776160008201527f70206f776e657200000000000000000000000000000000000000000000000000602082015250565b6000611cc4602783611beb565b9150611ccf82611c68565b604082019050919050565b60006020820190508181036000830152611cf381611cb7565b9050919050565b7f697427732074686520636f756e74657270617274792773207475726e2c20756e60008201527f61626c6520746f20726566756e642c2074727920616761696e206c6174657200602082015250565b6000611d56603f83611beb565b9150611d6182611cfa565b604082019050919050565b60006020820190508181036000830152611d8581611d49565b9050919050565b6000604082019050611da16000830185611985565b611dae6020830184611985565b9392505050565b6000819050919050565b6000611dda611dd5611dd0846115d6565b611db5565b6115d6565b9050919050565b6000611dec82611dbf565b9050919050565b6000611dfe82611de1565b9050919050565b611e0e81611df3565b82525050565b611e1d8161166a565b82525050565b6000604082019050611e386000830185611e05565b611e456020830184611e14565b9392505050565b611e5581611857565b8114611e6057600080fd5b50565b600081519050611e7281611e4c565b92915050565b600060208284031215611e8e57611e8d611540565b5b6000611e9c84828501611e63565b91505092915050565b7f4552433230207472616e73666572206661696c65640000000000000000000000600082015250565b6000611edb601583611beb565b9150611ee682611ea5565b602082019050919050565b60006020820190508181036000830152611f0a81611ece565b9050919050565b7f636c61696d52656c617965722063616e206f6e6c792062652063616c6c65642060008201527f62792061207472757374656420666f7277617264657200000000000000000000602082015250565b6000611f6d603683611beb565b9150611f7882611f11565b604082019050919050565b60006020820190508181036000830152611f9c81611f60565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000611fdd8261166a565b9150611fe88361166a565b925082820390508181111561200057611fff611fa3565b5b92915050565b7f4552433230207472616e7366657220746f20636c61696d6572206661696c6564600082015250565b600061203c602083611beb565b915061204782612006565b602082019050919050565b6000602082019050818103600083015261206b8161202f565b9050919050565b60006040820190506120876000830185611800565b6120946020830184611e14565b9392505050565b7f4552433230207472616e7366657220746f2072656c61796572206661696c6564600082015250565b60006120d1602083611beb565b91506120dc8261209b565b602082019050919050565b60006020820190508181036000830152612100816120c4565b9050919050565b7f76616c7565206e6f742073616d652061732045544820616d6f756e742073656e60008201527f7400000000000000000000000000000000000000000000000000000000000000602082015250565b6000612163602183611beb565b915061216e82612107565b604082019050919050565b6000602082019050818103600083015261219281612156565b9050919050565b60006121a48261166a565b91506121af8361166a565b92508282019050808211156121c7576121c6611fa3565b5b92915050565b60006121d88261166a565b91506121e38361166a565b92508282026121f18161166a565b9150828204841483151761220857612207611fa3565b5b5092915050565b7f7377617020616c72656164792065786973747300000000000000000000000000600082015250565b6000612245601383611beb565b91506122508261220f565b602082019050919050565b6000602082019050818103600083015261227481612238565b9050919050565b600060e082019050612290600083018a611985565b61229d6020830189611985565b6122aa6040830188611985565b6122b76060830187611e14565b6122c46080830186611e14565b6122d160a0830185611800565b6122de60c0830184611e14565b98975050505050505050565b60006060820190506122ff6000830186611800565b61230c6020830185611800565b6123196040830184611e14565b949350505050565b7f4552433230207472616e7366657246726f6d206661696c656400000000000000600082015250565b6000612357601983611beb565b915061236282612321565b602082019050919050565b600060208201905081810360008301526123868161234a565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000819050919050565b60008160001b9050919050565b60006123ee6123e96123e4846123bc565b6123c6565b611634565b9050919050565b6123fe816123d3565b82525050565b6000819050919050565b600060ff82169050919050565b600061243661243161242c84612404565b611db5565b61240e565b9050919050565b6124468161241b565b82525050565b600060808201905061246160008301876123f5565b61246e602083018661243d565b61247b6040830185611985565b6124886060830184611985565b95945050505050565b7f73776170206973206e6f7420696e2050454e44494e4720737461746500000000600082015250565b60006124c7601c83611beb565b91506124d282612491565b602082019050919050565b600060208201905081810360008301526124f6816124ba565b9050919050565b7f6f6e6c79207468652073776170206f776e65722063616e2063616c6c2073657460008201527f5265616479000000000000000000000000000000000000000000000000000000602082015250565b6000612559602583611beb565b9150612564826124fd565b604082019050919050565b600060208201905081810360008301526125888161254c565b9050919050565b7f70726f76696465642073656372657420646f6573206e6f74206d61746368207460008201527f6865206578706563746564207075626c6963206b657900000000000000000000602082015250565b60006125eb603683611beb565b91506125f68261258f565b604082019050919050565b6000602082019050818103600083015261261a816125de565b9050919050565b7f696e76616c696420737761700000000000000000000000000000000000000000600082015250565b6000612657600c83611beb565b915061266282612621565b602082019050919050565b600060208201905081810360008301526126868161264a565b9050919050565b7f6f6e6c7920636c61696d65722063616e20636c61696d21000000000000000000600082015250565b60006126c3601783611beb565b91506126ce8261268d565b602082019050919050565b600060208201905081810360008301526126f2816126b6565b9050919050565b7f746f6f206561726c7920746f20636c61696d2100000000000000000000000000600082015250565b600061272f601383611beb565b915061273a826126f9565b602082019050919050565b6000602082019050818103600083015261275e81612722565b9050919050565b7f746f6f206c61746520746f20636c61696d210000000000000000000000000000600082015250565b600061279b601283611beb565b91506127a682612765565b602082019050919050565b600060208201905081810360008301526127ca8161278e565b905091905056fea26469706673582212204b5a479afc7173ad7ac388458200e13a8e23ff3077c6b6b29f1fe89d6c126ec064736f6c63430008110033",
}

// SwapFactoryABI is the input ABI used to generate the binding from.
// Deprecated: Use SwapFactoryMetaData.ABI instead.
var SwapFactoryABI = SwapFactoryMetaData.ABI

// SwapFactoryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SwapFactoryMetaData.Bin instead.
var SwapFactoryBin = SwapFactoryMetaData.Bin

// DeploySwapFactory deploys a new Ethereum contract, binding an instance of SwapFactory to it.
func DeploySwapFactory(auth *bind.TransactOpts, backend bind.ContractBackend, trustedForwarder common.Address) (common.Address, *types.Transaction, *SwapFactory, error) {
	parsed, err := SwapFactoryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SwapFactoryBin), backend, trustedForwarder)
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

// TrustedForwarder is a free data retrieval call binding the contract method 0x56c022bb.
//
// Solidity: function _trustedForwarder() view returns(address)
func (_SwapFactory *SwapFactoryCaller) TrustedForwarder(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SwapFactory.contract.Call(opts, &out, "_trustedForwarder")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TrustedForwarder is a free data retrieval call binding the contract method 0x56c022bb.
//
// Solidity: function _trustedForwarder() view returns(address)
func (_SwapFactory *SwapFactorySession) TrustedForwarder() (common.Address, error) {
	return _SwapFactory.Contract.TrustedForwarder(&_SwapFactory.CallOpts)
}

// TrustedForwarder is a free data retrieval call binding the contract method 0x56c022bb.
//
// Solidity: function _trustedForwarder() view returns(address)
func (_SwapFactory *SwapFactoryCallerSession) TrustedForwarder() (common.Address, error) {
	return _SwapFactory.Contract.TrustedForwarder(&_SwapFactory.CallOpts)
}

// IsTrustedForwarder is a free data retrieval call binding the contract method 0x572b6c05.
//
// Solidity: function isTrustedForwarder(address forwarder) view returns(bool)
func (_SwapFactory *SwapFactoryCaller) IsTrustedForwarder(opts *bind.CallOpts, forwarder common.Address) (bool, error) {
	var out []interface{}
	err := _SwapFactory.contract.Call(opts, &out, "isTrustedForwarder", forwarder)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTrustedForwarder is a free data retrieval call binding the contract method 0x572b6c05.
//
// Solidity: function isTrustedForwarder(address forwarder) view returns(bool)
func (_SwapFactory *SwapFactorySession) IsTrustedForwarder(forwarder common.Address) (bool, error) {
	return _SwapFactory.Contract.IsTrustedForwarder(&_SwapFactory.CallOpts, forwarder)
}

// IsTrustedForwarder is a free data retrieval call binding the contract method 0x572b6c05.
//
// Solidity: function isTrustedForwarder(address forwarder) view returns(bool)
func (_SwapFactory *SwapFactoryCallerSession) IsTrustedForwarder(forwarder common.Address) (bool, error) {
	return _SwapFactory.Contract.IsTrustedForwarder(&_SwapFactory.CallOpts, forwarder)
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

// Claim is a paid mutator transaction binding the contract method 0x5cb96916.
//
// Solidity: function claim((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapFactory *SwapFactoryTransactor) Claim(opts *bind.TransactOpts, _swap SwapFactorySwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "claim", _swap, _s)
}

// Claim is a paid mutator transaction binding the contract method 0x5cb96916.
//
// Solidity: function claim((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapFactory *SwapFactorySession) Claim(_swap SwapFactorySwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.Contract.Claim(&_SwapFactory.TransactOpts, _swap, _s)
}

// Claim is a paid mutator transaction binding the contract method 0x5cb96916.
//
// Solidity: function claim((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapFactory *SwapFactoryTransactorSession) Claim(_swap SwapFactorySwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.Contract.Claim(&_SwapFactory.TransactOpts, _swap, _s)
}

// ClaimRelayer is a paid mutator transaction binding the contract method 0x73e4771c.
//
// Solidity: function claimRelayer((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s, uint256 fee) returns()
func (_SwapFactory *SwapFactoryTransactor) ClaimRelayer(opts *bind.TransactOpts, _swap SwapFactorySwap, _s [32]byte, fee *big.Int) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "claimRelayer", _swap, _s, fee)
}

// ClaimRelayer is a paid mutator transaction binding the contract method 0x73e4771c.
//
// Solidity: function claimRelayer((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s, uint256 fee) returns()
func (_SwapFactory *SwapFactorySession) ClaimRelayer(_swap SwapFactorySwap, _s [32]byte, fee *big.Int) (*types.Transaction, error) {
	return _SwapFactory.Contract.ClaimRelayer(&_SwapFactory.TransactOpts, _swap, _s, fee)
}

// ClaimRelayer is a paid mutator transaction binding the contract method 0x73e4771c.
//
// Solidity: function claimRelayer((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s, uint256 fee) returns()
func (_SwapFactory *SwapFactoryTransactorSession) ClaimRelayer(_swap SwapFactorySwap, _s [32]byte, fee *big.Int) (*types.Transaction, error) {
	return _SwapFactory.Contract.ClaimRelayer(&_SwapFactory.TransactOpts, _swap, _s, fee)
}

// NewSwap is a paid mutator transaction binding the contract method 0xaa0f8725.
//
// Solidity: function newSwap(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, address _claimer, uint256 _timeoutDuration, address _asset, uint256 _value, uint256 _nonce) payable returns(bytes32)
func (_SwapFactory *SwapFactoryTransactor) NewSwap(opts *bind.TransactOpts, _pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration *big.Int, _asset common.Address, _value *big.Int, _nonce *big.Int) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "newSwap", _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration, _asset, _value, _nonce)
}

// NewSwap is a paid mutator transaction binding the contract method 0xaa0f8725.
//
// Solidity: function newSwap(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, address _claimer, uint256 _timeoutDuration, address _asset, uint256 _value, uint256 _nonce) payable returns(bytes32)
func (_SwapFactory *SwapFactorySession) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration *big.Int, _asset common.Address, _value *big.Int, _nonce *big.Int) (*types.Transaction, error) {
	return _SwapFactory.Contract.NewSwap(&_SwapFactory.TransactOpts, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration, _asset, _value, _nonce)
}

// NewSwap is a paid mutator transaction binding the contract method 0xaa0f8725.
//
// Solidity: function newSwap(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, address _claimer, uint256 _timeoutDuration, address _asset, uint256 _value, uint256 _nonce) payable returns(bytes32)
func (_SwapFactory *SwapFactoryTransactorSession) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer common.Address, _timeoutDuration *big.Int, _asset common.Address, _value *big.Int, _nonce *big.Int) (*types.Transaction, error) {
	return _SwapFactory.Contract.NewSwap(&_SwapFactory.TransactOpts, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration, _asset, _value, _nonce)
}

// Refund is a paid mutator transaction binding the contract method 0x1e6c5acc.
//
// Solidity: function refund((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapFactory *SwapFactoryTransactor) Refund(opts *bind.TransactOpts, _swap SwapFactorySwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "refund", _swap, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x1e6c5acc.
//
// Solidity: function refund((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapFactory *SwapFactorySession) Refund(_swap SwapFactorySwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.Contract.Refund(&_SwapFactory.TransactOpts, _swap, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x1e6c5acc.
//
// Solidity: function refund((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapFactory *SwapFactoryTransactorSession) Refund(_swap SwapFactorySwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapFactory.Contract.Refund(&_SwapFactory.TransactOpts, _swap, _s)
}

// SetReady is a paid mutator transaction binding the contract method 0xfcaf229c.
//
// Solidity: function setReady((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap) returns()
func (_SwapFactory *SwapFactoryTransactor) SetReady(opts *bind.TransactOpts, _swap SwapFactorySwap) (*types.Transaction, error) {
	return _SwapFactory.contract.Transact(opts, "setReady", _swap)
}

// SetReady is a paid mutator transaction binding the contract method 0xfcaf229c.
//
// Solidity: function setReady((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap) returns()
func (_SwapFactory *SwapFactorySession) SetReady(_swap SwapFactorySwap) (*types.Transaction, error) {
	return _SwapFactory.Contract.SetReady(&_SwapFactory.TransactOpts, _swap)
}

// SetReady is a paid mutator transaction binding the contract method 0xfcaf229c.
//
// Solidity: function setReady((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap) returns()
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
	Asset     common.Address
	Value     *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterNew is a free log retrieval operation binding the contract event 0x91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56.
//
// Solidity: event New(bytes32 swapID, bytes32 claimKey, bytes32 refundKey, uint256 timeout0, uint256 timeout1, address asset, uint256 value)
func (_SwapFactory *SwapFactoryFilterer) FilterNew(opts *bind.FilterOpts) (*SwapFactoryNewIterator, error) {

	logs, sub, err := _SwapFactory.contract.FilterLogs(opts, "New")
	if err != nil {
		return nil, err
	}
	return &SwapFactoryNewIterator{contract: _SwapFactory.contract, event: "New", logs: logs, sub: sub}, nil
}

// WatchNew is a free log subscription operation binding the contract event 0x91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56.
//
// Solidity: event New(bytes32 swapID, bytes32 claimKey, bytes32 refundKey, uint256 timeout0, uint256 timeout1, address asset, uint256 value)
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

// ParseNew is a log parse operation binding the contract event 0x91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56.
//
// Solidity: event New(bytes32 swapID, bytes32 claimKey, bytes32 refundKey, uint256 timeout0, uint256 timeout1, address asset, uint256 value)
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
