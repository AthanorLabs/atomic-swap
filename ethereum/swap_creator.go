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
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"trustedForwarder\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"InvalidClaimer\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSecret\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSwap\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSwapKey\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidTimeout\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidValue\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotTimeToRefund\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OnlySwapClaimer\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OnlySwapOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OnlyTrustedForwarder\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SwapAlreadyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SwapCompleted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SwapNotPending\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TooEarlyToClaim\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TooLateToClaim\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroValue\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"claimKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"refundKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"New\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"}],\"name\":\"Ready\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"swapID\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"_trustedForwarder\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapCreator.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapCreator.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"}],\"name\":\"claimRelayer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"forwarder\",\"type\":\"address\"}],\"name\":\"isTrustedForwarder\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"scalar\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"qKeccak\",\"type\":\"uint256\"}],\"name\":\"mulVerify\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"addresspayable\",\"name\":\"_claimer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_timeoutDuration0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_timeoutDuration1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"newSwap\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapCreator.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"addresspayable\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"claimer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyClaim\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"pubKeyRefund\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timeout0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout1\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structSwapCreator.Swap\",\"name\":\"_swap\",\"type\":\"tuple\"}],\"name\":\"setReady\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"swaps\",\"outputs\":[{\"internalType\":\"enumSwapCreator.Stage\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60a06040523480156200001157600080fd5b506040516200200f3803806200200f8339818101604052810190620000379190620000de565b808073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff1681525050505062000110565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000620000a68262000079565b9050919050565b620000b88162000099565b8114620000c457600080fd5b50565b600081519050620000d881620000ad565b92915050565b600060208284031215620000f757620000f662000074565b5b60006200010784828501620000c7565b91505092915050565b608051611edc62000133600039600081816105c601526105ec0152611edc6000f3fe6080604052600436106100865760003560e01c806373e4771c1161005957806373e4771c14610145578063b32d1b4f1461016e578063c41e46cf146101ab578063eb84e7f2146101db578063fcaf229c1461021857610086565b80631e6c5acc1461008b57806356c022bb146100b4578063572b6c05146100df5780635cb969161461011c575b600080fd5b34801561009757600080fd5b506100b260048036038101906100ad919061170b565b610241565b005b3480156100c057600080fd5b506100c96105c4565b6040516100d6919061175c565b60405180910390f35b3480156100eb57600080fd5b5061010660048036038101906101019190611777565b6105e8565b60405161011391906117bf565b60405180910390f35b34801561012857600080fd5b50610143600480360381019061013e919061170b565b610640565b005b34801561015157600080fd5b5061016c600480360381019061016791906117da565b610766565b005b34801561017a57600080fd5b5061019560048036038101906101909190611830565b6109ac565b6040516101a291906117bf565b60405180910390f35b6101c560048036038101906101c09190611870565b610ab1565b6040516101d29190611935565b60405180910390f35b3480156101e757600080fd5b5061020260048036038101906101fd9190611950565b610f2b565b60405161020f91906119f4565b60405180910390f35b34801561022457600080fd5b5061023f600480360381019061023a9190611a0f565b610f4b565b005b6000826040516020016102549190611b30565b604051602081830303815290604052805190602001209050600080600083815260200190815260200160002060009054906101000a900460ff169050600060038111156102a4576102a361197d565b5b8160038111156102b7576102b661197d565b5b036102ee576040517f1115766700000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6003808111156103015761030061197d565b5b8160038111156103145761031361197d565b5b0361034b576040517f066916a900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b3373ffffffffffffffffffffffffffffffffffffffff16846000015173ffffffffffffffffffffffffffffffffffffffff16146103b4576040517f2919448600000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b8360a00151421080156103f9575083608001514211806103f85750600260038111156103e3576103e261197d565b5b8160038111156103f6576103f561197d565b5b145b5b15610430576040517f65430c1e00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b61043e8385606001516110c8565b82827e7c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f60405160405180910390a3600360008084815260200190815260200160002060006101000a81548160ff021916908360038111156104a1576104a061197d565b5b0217905550600073ffffffffffffffffffffffffffffffffffffffff168460c0015173ffffffffffffffffffffffffffffffffffffffff160361053257836000015173ffffffffffffffffffffffffffffffffffffffff166108fc8560e001519081150290604051600060405180830381858888f1935050505015801561052c573d6000803e3d6000fd5b506105be565b8360c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb85600001518660e001516040518363ffffffff1660e01b8152600401610579929190611bba565b6020604051808303816000875af1158015610598573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105bc9190611c0f565b505b50505050565b7f000000000000000000000000000000000000000000000000000000000000000081565b60007f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16149050919050565b61064a8282611112565b600073ffffffffffffffffffffffffffffffffffffffff168260c0015173ffffffffffffffffffffffffffffffffffffffff16036106d657816020015173ffffffffffffffffffffffffffffffffffffffff166108fc8360e001519081150290604051600060405180830381858888f193505050501580156106d0573d6000803e3d6000fd5b50610762565b8160c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb83602001518460e001516040518363ffffffff1660e01b815260040161071d929190611bba565b6020604051808303816000875af115801561073c573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107609190611c0f565b505b5050565b61076f336105e8565b6107a5576040517ffc5d4daa00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6107af8383611112565b600073ffffffffffffffffffffffffffffffffffffffff168360c0015173ffffffffffffffffffffffffffffffffffffffff160361088d57826020015173ffffffffffffffffffffffffffffffffffffffff166108fc828560e001516108159190611c6b565b9081150290604051600060405180830381858888f19350505050158015610840573d6000803e3d6000fd5b503273ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015610887573d6000803e3d6000fd5b506109a7565b8260c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb8460200151838660e001516108c29190611c6b565b6040518363ffffffff1660e01b81526004016108df929190611bba565b6020604051808303816000875af11580156108fe573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109229190611c0f565b508260c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb32836040518363ffffffff1660e01b8152600401610962929190611c9f565b6020604051808303816000875af1158015610981573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109a59190611c0f565b505b505050565b60008060016000601b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179860001b7ffffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd036414180610a0857610a07611cc8565b5b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798890960001b60405160008152602001604052604051610a4b9493929190611d87565b6020604051602081039080840390855afa158015610a6d573d6000803e3d6000fd5b5050506020604051035190508073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161491505092915050565b6000808303610aec576040517f7c946ed700000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610b5e57348314610b59576040517faa7feadc00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b610be0565b8373ffffffffffffffffffffffffffffffffffffffff166323b872dd3330866040518463ffffffff1660e01b8152600401610b9b93929190611dcc565b6020604051808303816000875af1158015610bba573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610bde9190611c0f565b505b6000801b891480610bf357506000801b88145b15610c2a576040517f378c37da00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff1603610c90576040517f044bc28000000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6000861480610c9f5750600085145b15610cd6576040517f7fee1bc400000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b610cde6113ef565b33816000019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff1681525050898160400181815250508881606001818152505087816020019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508642610d6e9190611e03565b816080018181525050858742610d849190611e03565b610d8e9190611e03565b8160a0018181525050848160c0019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff1681525050838160e00181815250508281610100018181525050600081604051602001610df79190611b30565b60405160208183030381529060405280519060200120905060006003811115610e2357610e2261197d565b5b60008083815260200190815260200160002060009054906101000a900460ff166003811115610e5557610e5461197d565b5b14610e8c576040517f734530ce00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b7f91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56818c8c85608001518660a001518760c001518860e00151604051610ed79796959493929190611e37565b60405180910390a1600160008083815260200190815260200160002060006101000a81548160ff02191690836003811115610f1557610f1461197d565b5b0217905550809250505098975050505050505050565b60006020528060005260406000206000915054906101000a900460ff1681565b600081604051602001610f5e9190611b30565b60405160208183030381529060405280519060200120905060016003811115610f8a57610f8961197d565b5b60008083815260200190815260200160002060009054906101000a900460ff166003811115610fbc57610fbb61197d565b5b14610ff3576040517f1fc1f6a200000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b3373ffffffffffffffffffffffffffffffffffffffff16826000015173ffffffffffffffffffffffffffffffffffffffff161461105c576040517f2919448600000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b600260008083815260200190815260200160002060006101000a81548160ff021916908360038111156110925761109161197d565b5b0217905550807f5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f60405160405180910390a25050565b6110d88260001c8260001c6109ac565b61110e576040517fabab6bd700000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5050565b6000826040516020016111259190611b30565b604051602081830303815290604052805190602001209050600080600083815260200190815260200160002060009054906101000a900460ff169050600060038111156111755761117461197d565b5b8160038111156111885761118761197d565b5b036111bf576040517f1115766700000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6003808111156111d2576111d161197d565b5b8160038111156111e5576111e461197d565b5b0361121c576040517f066916a900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b836020015173ffffffffffffffffffffffffffffffffffffffff1661123f6113b5565b73ffffffffffffffffffffffffffffffffffffffff161461128c576040517f68e2c81200000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b8360800151421080156112c45750600260038111156112ae576112ad61197d565b5b8160038111156112c1576112c061197d565b5b14155b156112fb576040517fd71d60b500000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b8360a001514210611338576040517f497df9d100000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6113468385604001516110c8565b82827f38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee60405160405180910390a3600360008084815260200190815260200160002060006101000a81548160ff021916908360038111156113aa576113a961197d565b5b021790555050505050565b60006113c0336105e8565b156113d457601436033560601c90506113e3565b6113dc6113e7565b90506113e4565b5b90565b600033905090565b604051806101200160405280600073ffffffffffffffffffffffffffffffffffffffff168152602001600073ffffffffffffffffffffffffffffffffffffffff16815260200160008019168152602001600080191681526020016000815260200160008152602001600073ffffffffffffffffffffffffffffffffffffffff16815260200160008152602001600081525090565b6000604051905090565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6114e082611497565b810181811067ffffffffffffffff821117156114ff576114fe6114a8565b5b80604052505050565b6000611512611483565b905061151e82826114d7565b919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600061154e82611523565b9050919050565b61155e81611543565b811461156957600080fd5b50565b60008135905061157b81611555565b92915050565b6000819050919050565b61159481611581565b811461159f57600080fd5b50565b6000813590506115b18161158b565b92915050565b6000819050919050565b6115ca816115b7565b81146115d557600080fd5b50565b6000813590506115e7816115c1565b92915050565b60006115f882611523565b9050919050565b611608816115ed565b811461161357600080fd5b50565b600081359050611625816115ff565b92915050565b6000610120828403121561164257611641611492565b5b61164d610120611508565b9050600061165d8482850161156c565b60008301525060206116718482850161156c565b6020830152506040611685848285016115a2565b6040830152506060611699848285016115a2565b60608301525060806116ad848285016115d8565b60808301525060a06116c1848285016115d8565b60a08301525060c06116d584828501611616565b60c08301525060e06116e9848285016115d8565b60e0830152506101006116fe848285016115d8565b6101008301525092915050565b60008061014083850312156117235761172261148d565b5b60006117318582860161162b565b925050610120611743858286016115a2565b9150509250929050565b611756816115ed565b82525050565b6000602082019050611771600083018461174d565b92915050565b60006020828403121561178d5761178c61148d565b5b600061179b84828501611616565b91505092915050565b60008115159050919050565b6117b9816117a4565b82525050565b60006020820190506117d460008301846117b0565b92915050565b600080600061016084860312156117f4576117f361148d565b5b60006118028682870161162b565b935050610120611814868287016115a2565b925050610140611826868287016115d8565b9150509250925092565b600080604083850312156118475761184661148d565b5b6000611855858286016115d8565b9250506020611866858286016115d8565b9150509250929050565b600080600080600080600080610100898b0312156118915761189061148d565b5b600061189f8b828c016115a2565b98505060206118b08b828c016115a2565b97505060406118c18b828c0161156c565b96505060606118d28b828c016115d8565b95505060806118e38b828c016115d8565b94505060a06118f48b828c01611616565b93505060c06119058b828c016115d8565b92505060e06119168b828c016115d8565b9150509295985092959890939650565b61192f81611581565b82525050565b600060208201905061194a6000830184611926565b92915050565b6000602082840312156119665761196561148d565b5b6000611974848285016115a2565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b600481106119bd576119bc61197d565b5b50565b60008190506119ce826119ac565b919050565b60006119de826119c0565b9050919050565b6119ee816119d3565b82525050565b6000602082019050611a0960008301846119e5565b92915050565b60006101208284031215611a2657611a2561148d565b5b6000611a348482850161162b565b91505092915050565b611a4681611543565b82525050565b611a5581611581565b82525050565b611a64816115b7565b82525050565b611a73816115ed565b82525050565b61012082016000820151611a906000850182611a3d565b506020820151611aa36020850182611a3d565b506040820151611ab66040850182611a4c565b506060820151611ac96060850182611a4c565b506080820151611adc6080850182611a5b565b5060a0820151611aef60a0850182611a5b565b5060c0820151611b0260c0850182611a6a565b5060e0820151611b1560e0850182611a5b565b50610100820151611b2a610100850182611a5b565b50505050565b600061012082019050611b466000830184611a79565b92915050565b6000819050919050565b6000611b71611b6c611b6784611523565b611b4c565b611523565b9050919050565b6000611b8382611b56565b9050919050565b6000611b9582611b78565b9050919050565b611ba581611b8a565b82525050565b611bb4816115b7565b82525050565b6000604082019050611bcf6000830185611b9c565b611bdc6020830184611bab565b9392505050565b611bec816117a4565b8114611bf757600080fd5b50565b600081519050611c0981611be3565b92915050565b600060208284031215611c2557611c2461148d565b5b6000611c3384828501611bfa565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000611c76826115b7565b9150611c81836115b7565b9250828203905081811115611c9957611c98611c3c565b5b92915050565b6000604082019050611cb4600083018561174d565b611cc16020830184611bab565b9392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000819050919050565b60008160001b9050919050565b6000611d29611d24611d1f84611cf7565b611d01565b611581565b9050919050565b611d3981611d0e565b82525050565b6000819050919050565b600060ff82169050919050565b6000611d71611d6c611d6784611d3f565b611b4c565b611d49565b9050919050565b611d8181611d56565b82525050565b6000608082019050611d9c6000830187611d30565b611da96020830186611d78565b611db66040830185611926565b611dc36060830184611926565b95945050505050565b6000606082019050611de1600083018661174d565b611dee602083018561174d565b611dfb6040830184611bab565b949350505050565b6000611e0e826115b7565b9150611e19836115b7565b9250828201905080821115611e3157611e30611c3c565b5b92915050565b600060e082019050611e4c600083018a611926565b611e596020830189611926565b611e666040830188611926565b611e736060830187611bab565b611e806080830186611bab565b611e8d60a083018561174d565b611e9a60c0830184611bab565b9897505050505050505056fea26469706673582212204511ee58616aee393f94e8e7e0b89860a827708dae09ea2476924a6705a6a18164736f6c63430008130033",
}

// SwapCreatorABI is the input ABI used to generate the binding from.
// Deprecated: Use SwapCreatorMetaData.ABI instead.
var SwapCreatorABI = SwapCreatorMetaData.ABI

// SwapCreatorBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SwapCreatorMetaData.Bin instead.
var SwapCreatorBin = SwapCreatorMetaData.Bin

// DeploySwapCreator deploys a new Ethereum contract, binding an instance of SwapCreator to it.
func DeploySwapCreator(auth *bind.TransactOpts, backend bind.ContractBackend, trustedForwarder common.Address) (common.Address, *types.Transaction, *SwapCreator, error) {
	parsed, err := SwapCreatorMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SwapCreatorBin), backend, trustedForwarder)
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

// TrustedForwarder is a free data retrieval call binding the contract method 0x56c022bb.
//
// Solidity: function _trustedForwarder() view returns(address)
func (_SwapCreator *SwapCreatorCaller) TrustedForwarder(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SwapCreator.contract.Call(opts, &out, "_trustedForwarder")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TrustedForwarder is a free data retrieval call binding the contract method 0x56c022bb.
//
// Solidity: function _trustedForwarder() view returns(address)
func (_SwapCreator *SwapCreatorSession) TrustedForwarder() (common.Address, error) {
	return _SwapCreator.Contract.TrustedForwarder(&_SwapCreator.CallOpts)
}

// TrustedForwarder is a free data retrieval call binding the contract method 0x56c022bb.
//
// Solidity: function _trustedForwarder() view returns(address)
func (_SwapCreator *SwapCreatorCallerSession) TrustedForwarder() (common.Address, error) {
	return _SwapCreator.Contract.TrustedForwarder(&_SwapCreator.CallOpts)
}

// IsTrustedForwarder is a free data retrieval call binding the contract method 0x572b6c05.
//
// Solidity: function isTrustedForwarder(address forwarder) view returns(bool)
func (_SwapCreator *SwapCreatorCaller) IsTrustedForwarder(opts *bind.CallOpts, forwarder common.Address) (bool, error) {
	var out []interface{}
	err := _SwapCreator.contract.Call(opts, &out, "isTrustedForwarder", forwarder)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTrustedForwarder is a free data retrieval call binding the contract method 0x572b6c05.
//
// Solidity: function isTrustedForwarder(address forwarder) view returns(bool)
func (_SwapCreator *SwapCreatorSession) IsTrustedForwarder(forwarder common.Address) (bool, error) {
	return _SwapCreator.Contract.IsTrustedForwarder(&_SwapCreator.CallOpts, forwarder)
}

// IsTrustedForwarder is a free data retrieval call binding the contract method 0x572b6c05.
//
// Solidity: function isTrustedForwarder(address forwarder) view returns(bool)
func (_SwapCreator *SwapCreatorCallerSession) IsTrustedForwarder(forwarder common.Address) (bool, error) {
	return _SwapCreator.Contract.IsTrustedForwarder(&_SwapCreator.CallOpts, forwarder)
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
// Solidity: function claim((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapCreator *SwapCreatorTransactor) Claim(opts *bind.TransactOpts, _swap SwapCreatorSwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapCreator.contract.Transact(opts, "claim", _swap, _s)
}

// Claim is a paid mutator transaction binding the contract method 0x5cb96916.
//
// Solidity: function claim((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapCreator *SwapCreatorSession) Claim(_swap SwapCreatorSwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapCreator.Contract.Claim(&_SwapCreator.TransactOpts, _swap, _s)
}

// Claim is a paid mutator transaction binding the contract method 0x5cb96916.
//
// Solidity: function claim((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapCreator *SwapCreatorTransactorSession) Claim(_swap SwapCreatorSwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapCreator.Contract.Claim(&_SwapCreator.TransactOpts, _swap, _s)
}

// ClaimRelayer is a paid mutator transaction binding the contract method 0x73e4771c.
//
// Solidity: function claimRelayer((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s, uint256 fee) returns()
func (_SwapCreator *SwapCreatorTransactor) ClaimRelayer(opts *bind.TransactOpts, _swap SwapCreatorSwap, _s [32]byte, fee *big.Int) (*types.Transaction, error) {
	return _SwapCreator.contract.Transact(opts, "claimRelayer", _swap, _s, fee)
}

// ClaimRelayer is a paid mutator transaction binding the contract method 0x73e4771c.
//
// Solidity: function claimRelayer((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s, uint256 fee) returns()
func (_SwapCreator *SwapCreatorSession) ClaimRelayer(_swap SwapCreatorSwap, _s [32]byte, fee *big.Int) (*types.Transaction, error) {
	return _SwapCreator.Contract.ClaimRelayer(&_SwapCreator.TransactOpts, _swap, _s, fee)
}

// ClaimRelayer is a paid mutator transaction binding the contract method 0x73e4771c.
//
// Solidity: function claimRelayer((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s, uint256 fee) returns()
func (_SwapCreator *SwapCreatorTransactorSession) ClaimRelayer(_swap SwapCreatorSwap, _s [32]byte, fee *big.Int) (*types.Transaction, error) {
	return _SwapCreator.Contract.ClaimRelayer(&_SwapCreator.TransactOpts, _swap, _s, fee)
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
// Solidity: function refund((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapCreator *SwapCreatorTransactor) Refund(opts *bind.TransactOpts, _swap SwapCreatorSwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapCreator.contract.Transact(opts, "refund", _swap, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x1e6c5acc.
//
// Solidity: function refund((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapCreator *SwapCreatorSession) Refund(_swap SwapCreatorSwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapCreator.Contract.Refund(&_SwapCreator.TransactOpts, _swap, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x1e6c5acc.
//
// Solidity: function refund((address,address,bytes32,bytes32,uint256,uint256,address,uint256,uint256) _swap, bytes32 _s) returns()
func (_SwapCreator *SwapCreatorTransactorSession) Refund(_swap SwapCreatorSwap, _s [32]byte) (*types.Transaction, error) {
	return _SwapCreator.Contract.Refund(&_SwapCreator.TransactOpts, _swap, _s)
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
