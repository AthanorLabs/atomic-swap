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
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_x_alice\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_y_alice\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_x_bob\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_y_bob\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"s\",\"type\":\"uint256\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"x\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"y\",\"type\":\"bytes32\"}],\"name\":\"Constructed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"b\",\"type\":\"bool\"}],\"name\":\"IsReady\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"s\",\"type\":\"uint256\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_s\",\"type\":\"uint256\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_s\",\"type\":\"uint256\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"set_ready\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeout_0\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeout_1\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"x_alice\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"x_bob\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"y_alice\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"y_bob\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60806040526000600860006101000a81548160ff02191690831515021790555060405162001c3138038062001c318339818101604052810190620000449190620001ba565b33600160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550836004819055508260058190555081600281905550806003819055506201518042620000b2919062000265565b600681905550604051620000c6906200016c565b604051809103906000f080158015620000e3573d6000803e3d6000fd5b506000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8d36aa70807342c3036697a846281194626fd4afa892356ad5979e03831ab0806004546005546040516200015a929190620002d3565b60405180910390a15050505062000300565b6110118062000c2083390190565b600080fd5b6000819050919050565b62000194816200017f565b8114620001a057600080fd5b50565b600081519050620001b48162000189565b92915050565b60008060008060808587031215620001d757620001d66200017a565b5b6000620001e787828801620001a3565b9450506020620001fa87828801620001a3565b93505060406200020d87828801620001a3565b92505060606200022087828801620001a3565b91505092959194509250565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600062000272826200022c565b91506200027f836200022c565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03821115620002b757620002b662000236565b5b828201905092915050565b620002cd816200017f565b82525050565b6000604082019050620002ea6000830185620002c2565b620002f96020830184620002c2565b9392505050565b61091080620003106000396000f3fe608060405234801561001057600080fd5b50600436106100935760003560e01c806345bb8e091161006657806345bb8e091461010c5780634ded8d521461012a57806374d7c1381461014857806396edb86214610152578063df9b0c951461017057610093565b8063278ecde1146100985780632c86cd7c146100b45780633181b2a4146100d2578063379607f5146100f0575b600080fd5b6100b260048036038101906100ad919061059e565b61018e565b005b6100bc61025b565b6040516100c991906105e4565b60405180910390f35b6100da610261565b6040516100e791906105e4565b60405180910390f35b61010a6004803603810190610105919061059e565b610267565b005b610114610371565b604051610121919061060e565b60405180910390f35b610132610377565b60405161013f919061060e565b60405180910390f35b61015061037d565b005b61015a61044e565b60405161016791906105e4565b60405180910390f35b610178610454565b60405161018591906105e4565b60405180910390f35b600860009054906101000a900460ff161580156101ac575060065442105b806101d15750600860009054906101000a900460ff1680156101d057506007544210155b5b6101da57600080fd5b6101e98160045460055461045a565b7f3d2a04f53164bedf9a8a46353305d6b2d2261410406df3b41f99ce6489dc003c81604051610218919061060e565b60405180910390a1600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16ff5b60055481565b60045481565b60011515600860009054906101000a900460ff16151514156102cc5760075442106102c7576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102be90610686565b60405180910390fd5b610312565b600654421015610311576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161030890610718565b60405180910390fd5b5b6103218160025460035461045a565b7f7a355715549cfe7c1cba26304350343fbddc4b4f72d3ce3e7c27117dd20b5cb881604051610350919061060e565b60405180910390a13373ffffffffffffffffffffffffffffffffffffffff16ff5b60075481565b60065481565b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161480156103db575060065442105b6103e457600080fd5b6001600860006101000a81548160ff021916908315150217905550620151804261040e9190610767565b6007819055507f2724cf6c3ad6a3399ad72482e4013d0171794f3ef4c462b7e24790c658cb3cd4600160405161044491906107d8565b60405180910390a1565b60035481565b60025481565b60008060008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663bc9e2bcf866040518263ffffffff1660e01b81526004016104b6919061060e565b604080518083038186803b1580156104cd57600080fd5b505afa1580156104e1573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105059190610808565b91509150838260001b14801561051d5750828160001b145b61055c576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610553906108ba565b60405180910390fd5b5050505050565b600080fd5b6000819050919050565b61057b81610568565b811461058657600080fd5b50565b60008135905061059881610572565b92915050565b6000602082840312156105b4576105b3610563565b5b60006105c284828501610589565b91505092915050565b6000819050919050565b6105de816105cb565b82525050565b60006020820190506105f960008301846105d5565b92915050565b61060881610568565b82525050565b600060208201905061062360008301846105ff565b92915050565b600082825260208201905092915050565b7f546f6f206c61746520746f20636c61696d210000000000000000000000000000600082015250565b6000610670601283610629565b915061067b8261063a565b602082019050919050565b6000602082019050818103600083015261069f81610663565b9050919050565b7f2769735265616479203d3d2066616c7365272063616e6e6f7420636c61696d2060008201527f7965742100000000000000000000000000000000000000000000000000000000602082015250565b6000610702602483610629565b915061070d826106a6565b604082019050919050565b60006020820190508181036000830152610731816106f5565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061077282610568565b915061077d83610568565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156107b2576107b1610738565b5b828201905092915050565b60008115159050919050565b6107d2816107bd565b82525050565b60006020820190506107ed60008301846107c9565b92915050565b60008151905061080281610572565b92915050565b6000806040838503121561081f5761081e610563565b5b600061082d858286016107f3565b925050602061083e858286016107f3565b9150509250929050565b7f70726f76696465642073656372657420646f6573206e6f74206d61746368207460008201527f6865206578706563746564207075624b65790000000000000000000000000000602082015250565b60006108a4603283610629565b91506108af82610848565b604082019050919050565b600060208201905081810360008301526108d381610897565b905091905056fea2646970667358221220a38da8cdb7d67672f3e98ced8bb06409ccc1ba71c8382c1cd3b7188812e9349464736f6c63430008090033608060405234801561001057600080fd5b50610ff1806100206000396000f3fe608060405234801561001057600080fd5b50600436106100575760003560e01c806303a507be1461005c5780637a308a4c1461007a578063997da8d414610098578063bc9e2bcf146100b6578063eeeac01e146100e7575b600080fd5b610064610105565b6040516100719190610ce4565b60405180910390f35b610082610129565b60405161008f9190610ce4565b60405180910390f35b6100a061014d565b6040516100ad9190610ce4565b60405180910390f35b6100d060048036038101906100cb9190610d30565b610171565b6040516100de929190610d5d565b60405180910390f35b6100ef61020a565b6040516100fc9190610ce4565b60405180910390f35b7f216936d3cd6e53fec0a4e231fdd6dc5c692cc7609525a7b2c9562d608f25d51a81565b7f666666666666666666666666666666666666666666666666666666666666665881565b7f52036cee2b6ffe738cc740797779e89800700a4d4141d8ab75eb4dca135978a381565b600080610201837f216936d3cd6e53fec0a4e231fdd6dc5c692cc7609525a7b2c9562d608f25d51a7f66666666666666666666666666666666666666666666666666666666666666587f52036cee2b6ffe738cc740797779e89800700a4d4141d8ab75eb4dca135978a37f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed61022e565b91509150915091565b7f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed81565b60008060008060006102458a8a8a60018b8b610268565b92509250925061025783838389610308565b945094505050509550959350505050565b600080600080891415610283578787879250925092506102fc565b60008990506000806000600190505b600084146102ee57600060018516146102c1576102b48383838f8f8f8e61037f565b8093508194508295505050505b6002846102ce9190610de4565b93506102dd8c8c8c8c8c6109fe565b809c50819d50829e50505050610292565b828282965096509650505050505b96509650969350505050565b60008060006103178585610bbd565b90506000848061032a57610329610d86565b5b8283099050600085806103405761033f610d86565b5b828a0990506000868061035657610355610d86565b5b878061036557610364610d86565b5b8486098a0990508181955095505050505094509492505050565b6000806000808a1480156103935750600089145b156103a6578686869250925092506109f1565b6000871480156103b65750600086145b156103c9578989899250925092506109f1565b6103d1610ca9565b84806103e0576103df610d86565b5b898a09816000600481106103f7576103f6610e15565b5b602002018181525050848061040f5761040e610d86565b5b8160006004811061042357610422610e15565b5b60200201518a098160016004811061043e5761043d610e15565b5b602002018181525050848061045657610455610d86565b5b8687098160026004811061046d5761046c610e15565b5b602002018181525050848061048557610484610d86565b5b8160026004811061049957610498610e15565b5b60200201518709816003600481106104b4576104b3610e15565b5b602002018181525050604051806080016040528086806104d7576104d6610d86565b5b836002600481106104eb576104ea610e15565b5b60200201518e098152602001868061050657610505610d86565b5b8360036004811061051a57610519610e15565b5b60200201518d098152602001868061053557610534610d86565b5b8360006004811061054957610548610e15565b5b60200201518b098152602001868061056457610563610d86565b5b8360016004811061057857610577610e15565b5b60200201518a0981525090508060026004811061059857610597610e15565b5b6020020151816000600481106105b1576105b0610e15565b5b60200201511415806105f35750806003600481106105d2576105d1610e15565b5b6020020151816001600481106105eb576105ea610e15565b5b602002015114155b610632576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161062990610ea1565b60405180910390fd5b61063a610ca9565b858061064957610648610d86565b5b8260006004811061065d5761065c610e15565b5b60200201518761066d9190610ec1565b8360026004811061068157610680610e15565b5b6020020151088160006004811061069b5761069a610e15565b5b60200201818152505085806106b3576106b2610d86565b5b826001600481106106c7576106c6610e15565b5b6020020151876106d79190610ec1565b836003600481106106eb576106ea610e15565b5b6020020151088160016004811061070557610704610e15565b5b602002018181525050858061071d5761071c610d86565b5b8160006004811061073157610730610e15565b5b60200201518260006004811061074a57610749610e15565b5b6020020151098160026004811061076457610763610e15565b5b602002018181525050858061077c5761077b610d86565b5b816000600481106107905761078f610e15565b5b6020020151826002600481106107a9576107a8610e15565b5b602002015109816003600481106107c3576107c2610e15565b5b602002018181525050600086806107dd576107dc610d86565b5b826003600481106107f1576107f0610e15565b5b6020020151886108019190610ec1565b88806108105761080f610d86565b5b8460016004811061082457610823610e15565b5b60200201518560016004811061083d5761083c610e15565b5b602002015109089050868061085557610854610d86565b5b878061086457610863610d86565b5b888061087357610872610d86565b5b8460026004811061088757610886610e15565b5b6020020151866000600481106108a05761089f610e15565b5b602002015109600209886108b49190610ec1565b82089050600087806108c9576108c8610d86565b5b88806108d8576108d7610d86565b5b838a6108e49190610ec1565b8a806108f3576108f2610d86565b5b8660026004811061090757610906610e15565b5b6020020151886000600481106109205761091f610e15565b5b602002015109088460016004811061093b5761093a610e15565b5b6020020151099050878061095257610951610d86565b5b888061096157610960610d86565b5b8460036004811061097557610974610e15565b5b60200201518660016004811061098e5761098d610e15565b5b6020020151098961099f9190610ec1565b82089050600088806109b4576109b3610d86565b5b89806109c3576109c2610d86565b5b8b8f09856000600481106109da576109d9610e15565b5b602002015109905082828297509750975050505050505b9750975097945050505050565b600080600080861415610a1957878787925092509250610bb2565b60008480610a2a57610a29610d86565b5b898a09905060008580610a4057610a3f610d86565b5b898a09905060008680610a5657610a55610d86565b5b898a09905060008780610a6c57610a6b610d86565b5b8880610a7b57610a7a610d86565b5b848e09600409905060008880610a9457610a93610d86565b5b8980610aa357610aa2610d86565b5b8a80610ab257610ab1610d86565b5b8586098c098a80610ac657610ac5610d86565b5b876003090890508880610adc57610adb610d86565b5b8980610aeb57610aea610d86565b5b8384088a610af99190610ec1565b8a80610b0857610b07610d86565b5b8384090894508880610b1d57610b1c610d86565b5b8980610b2c57610b2b610d86565b5b8a80610b3b57610b3a610d86565b5b8687096008098a610b4c9190610ec1565b8a80610b5b57610b5a610d86565b5b8b80610b6a57610b69610d86565b5b888d610b769190610ec1565b860884090893508880610b8c57610b8b610d86565b5b8980610b9b57610b9a610d86565b5b8c8e09600209925084848497509750975050505050505b955095509592505050565b6000808314158015610bcf5750818314155b8015610bdc575060008214155b610c1b576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610c1290610f41565b60405180910390fd5b60008060019050600084905060005b60008714610c9c578682610c3e9190610de4565b9050828680610c5057610c4f610d86565b5b8780610c5f57610c5e610d86565b5b85840988610c6d9190610ec1565b86088094508195505050868782610c849190610f61565b83610c8f9190610ec1565b8098508193505050610c2a565b8394505050505092915050565b6040518060800160405280600490602082028036833780820191505090505090565b6000819050919050565b610cde81610ccb565b82525050565b6000602082019050610cf96000830184610cd5565b92915050565b600080fd5b610d0d81610ccb565b8114610d1857600080fd5b50565b600081359050610d2a81610d04565b92915050565b600060208284031215610d4657610d45610cff565b5b6000610d5484828501610d1b565b91505092915050565b6000604082019050610d726000830185610cd5565b610d7f6020830184610cd5565b9392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610def82610ccb565b9150610dfa83610ccb565b925082610e0a57610e09610d86565b5b828204905092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b600082825260208201905092915050565b7f557365206a6163446f75626c652066756e6374696f6e20696e73746561640000600082015250565b6000610e8b601e83610e44565b9150610e9682610e55565b602082019050919050565b60006020820190508181036000830152610eba81610e7e565b9050919050565b6000610ecc82610ccb565b9150610ed783610ccb565b925082821015610eea57610ee9610db5565b5b828203905092915050565b7f496e76616c6964206e756d626572000000000000000000000000000000000000600082015250565b6000610f2b600e83610e44565b9150610f3682610ef5565b602082019050919050565b60006020820190508181036000830152610f5a81610f1e565b9050919050565b6000610f6c82610ccb565b9150610f7783610ccb565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0483118215151615610fb057610faf610db5565b5b82820290509291505056fea26469706673582212200d234a559d972160747aa4d86018ee52b6966df6472b51a6339b5e39b38972ca64736f6c63430008090033",
}

// SwapABI is the input ABI used to generate the binding from.
// Deprecated: Use SwapMetaData.ABI instead.
var SwapABI = SwapMetaData.ABI

// SwapBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SwapMetaData.Bin instead.
var SwapBin = SwapMetaData.Bin

// DeploySwap deploys a new Ethereum contract, binding an instance of Swap to it.
func DeploySwap(auth *bind.TransactOpts, backend bind.ContractBackend, _x_alice [32]byte, _y_alice [32]byte, _x_bob [32]byte, _y_bob [32]byte) (common.Address, *types.Transaction, *Swap, error) {
	parsed, err := SwapMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SwapBin), backend, _x_alice, _y_alice, _x_bob, _y_bob)
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

// XAlice is a free data retrieval call binding the contract method 0x3181b2a4.
//
// Solidity: function x_alice() view returns(bytes32)
func (_Swap *SwapCaller) XAlice(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "x_alice")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// XAlice is a free data retrieval call binding the contract method 0x3181b2a4.
//
// Solidity: function x_alice() view returns(bytes32)
func (_Swap *SwapSession) XAlice() ([32]byte, error) {
	return _Swap.Contract.XAlice(&_Swap.CallOpts)
}

// XAlice is a free data retrieval call binding the contract method 0x3181b2a4.
//
// Solidity: function x_alice() view returns(bytes32)
func (_Swap *SwapCallerSession) XAlice() ([32]byte, error) {
	return _Swap.Contract.XAlice(&_Swap.CallOpts)
}

// XBob is a free data retrieval call binding the contract method 0xdf9b0c95.
//
// Solidity: function x_bob() view returns(bytes32)
func (_Swap *SwapCaller) XBob(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "x_bob")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// XBob is a free data retrieval call binding the contract method 0xdf9b0c95.
//
// Solidity: function x_bob() view returns(bytes32)
func (_Swap *SwapSession) XBob() ([32]byte, error) {
	return _Swap.Contract.XBob(&_Swap.CallOpts)
}

// XBob is a free data retrieval call binding the contract method 0xdf9b0c95.
//
// Solidity: function x_bob() view returns(bytes32)
func (_Swap *SwapCallerSession) XBob() ([32]byte, error) {
	return _Swap.Contract.XBob(&_Swap.CallOpts)
}

// YAlice is a free data retrieval call binding the contract method 0x2c86cd7c.
//
// Solidity: function y_alice() view returns(bytes32)
func (_Swap *SwapCaller) YAlice(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "y_alice")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// YAlice is a free data retrieval call binding the contract method 0x2c86cd7c.
//
// Solidity: function y_alice() view returns(bytes32)
func (_Swap *SwapSession) YAlice() ([32]byte, error) {
	return _Swap.Contract.YAlice(&_Swap.CallOpts)
}

// YAlice is a free data retrieval call binding the contract method 0x2c86cd7c.
//
// Solidity: function y_alice() view returns(bytes32)
func (_Swap *SwapCallerSession) YAlice() ([32]byte, error) {
	return _Swap.Contract.YAlice(&_Swap.CallOpts)
}

// YBob is a free data retrieval call binding the contract method 0x96edb862.
//
// Solidity: function y_bob() view returns(bytes32)
func (_Swap *SwapCaller) YBob(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Swap.contract.Call(opts, &out, "y_bob")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// YBob is a free data retrieval call binding the contract method 0x96edb862.
//
// Solidity: function y_bob() view returns(bytes32)
func (_Swap *SwapSession) YBob() ([32]byte, error) {
	return _Swap.Contract.YBob(&_Swap.CallOpts)
}

// YBob is a free data retrieval call binding the contract method 0x96edb862.
//
// Solidity: function y_bob() view returns(bytes32)
func (_Swap *SwapCallerSession) YBob() ([32]byte, error) {
	return _Swap.Contract.YBob(&_Swap.CallOpts)
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
	X   [32]byte
	Y   [32]byte
	Raw types.Log // Blockchain specific contextual infos
}

// FilterConstructed is a free log retrieval operation binding the contract event 0x8d36aa70807342c3036697a846281194626fd4afa892356ad5979e03831ab080.
//
// Solidity: event Constructed(bytes32 x, bytes32 y)
func (_Swap *SwapFilterer) FilterConstructed(opts *bind.FilterOpts) (*SwapConstructedIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "Constructed")
	if err != nil {
		return nil, err
	}
	return &SwapConstructedIterator{contract: _Swap.contract, event: "Constructed", logs: logs, sub: sub}, nil
}

// WatchConstructed is a free log subscription operation binding the contract event 0x8d36aa70807342c3036697a846281194626fd4afa892356ad5979e03831ab080.
//
// Solidity: event Constructed(bytes32 x, bytes32 y)
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
// Solidity: event Constructed(bytes32 x, bytes32 y)
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
