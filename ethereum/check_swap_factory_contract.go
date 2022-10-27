package contracts

import (
	"bytes"
	"context"
	"errors"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// expectedSwapFactoryBytecodeHex is generated by deploying an instance of SwapFactory.sol
// with the trustedForwarder address set to all zeros and reading back the bytecode. See
// the unit test TestExpectedSwapFactoryBytecodeHex if you need to update this value.
const (
	expectedSwapFactoryBytecodeHex = "6080604052600436106100865760003560e01c806373e4771c1161005957806373e4771c14610145578063aa0f87251461016e578063b32d1b4f1461019e578063eb84e7f2146101db578063fcaf229c1461021857610086565b80631e6c5acc1461008b57806356c022bb146100b4578063572b6c05146100df5780635cb969161461011c575b600080fd5b34801561009757600080fd5b506100b260048036038101906100ad91906115eb565b610241565b005b3480156100c057600080fd5b506100c96105bf565b6040516100d6919061163c565b60405180910390f35b3480156100eb57600080fd5b5061010660048036038101906101019190611657565b6105e3565b604051610113919061169f565b60405180910390f35b34801561012857600080fd5b50610143600480360381019061013e91906115eb565b61063b565b005b34801561015157600080fd5b5061016c600480360381019061016791906116ba565b610761565b005b61018860048036038101906101839190611710565b610974565b60405161019591906117c1565b60405180910390f35b3480156101aa57600080fd5b506101c560048036038101906101c091906117dc565b610cab565b6040516101d2919061169f565b60405180910390f35b3480156101e757600080fd5b5061020260048036038101906101fd919061181c565b610db0565b60405161020f91906118c0565b60405180910390f35b34801561022457600080fd5b5061023f600480360381019061023a91906118db565b610dd0565b005b60008260405160200161025491906119fc565b604051602081830303815290604052805190602001209050600080600083815260200190815260200160002060009054906101000a900460ff1690506003808111156102a3576102a2611849565b5b8160038111156102b6576102b5611849565b5b141580156102e95750600060038111156102d3576102d2611849565b5b8160038111156102e6576102e5611849565b5b14155b610328576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161031f90611a75565b60405180910390fd5b836000015173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461039a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161039190611b07565b60405180910390fd5b8360a00151421015806103e157508360800151421080156103e05750600260038111156103ca576103c9611849565b5b8160038111156103dd576103dc611849565b5b14155b5b610420576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161041790611b99565b60405180910390fd5b61042e838560600151610f69565b7e7c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f828460405161045e929190611bb9565b60405180910390a1600360008084815260200190815260200160002060006101000a81548160ff0219169083600381111561049c5761049b611849565b5b0217905550600073ffffffffffffffffffffffffffffffffffffffff168460c0015173ffffffffffffffffffffffffffffffffffffffff160361052d57836000015173ffffffffffffffffffffffffffffffffffffffff166108fc8560e001519081150290604051600060405180830381858888f19350505050158015610527573d6000803e3d6000fd5b506105b9565b8360c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb85600001518660e001516040518363ffffffff1660e01b8152600401610574929190611c50565b6020604051808303816000875af1158015610593573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105b79190611ca5565b505b50505050565b7f000000000000000000000000000000000000000000000000000000000000000081565b60007f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16149050919050565b6106458282610fbc565b600073ffffffffffffffffffffffffffffffffffffffff168260c0015173ffffffffffffffffffffffffffffffffffffffff16036106d157816020015173ffffffffffffffffffffffffffffffffffffffff166108fc8360e001519081150290604051600060405180830381858888f193505050501580156106cb573d6000803e3d6000fd5b5061075d565b8160c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb83602001518460e001516040518363ffffffff1660e01b8152600401610718929190611c50565b6020604051808303816000875af1158015610737573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061075b9190611ca5565b505b5050565b61076a336105e3565b6107a9576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107a090611d44565b60405180910390fd5b6107b38383610fbc565b600073ffffffffffffffffffffffffffffffffffffffff168360c0015173ffffffffffffffffffffffffffffffffffffffff160361089157826020015173ffffffffffffffffffffffffffffffffffffffff166108fc828560e001516108199190611d93565b9081150290604051600060405180830381858888f19350505050158015610844573d6000803e3d6000fd5b503273ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f1935050505015801561088b573d6000803e3d6000fd5b5061096f565b8260c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb8460200151838660e001516108c69190611d93565b6040518363ffffffff1660e01b81526004016108e3929190611c50565b6020604051808303816000875af1158015610902573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109269190611ca5565b503273ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f1935050505015801561096d573d6000803e3d6000fd5b505b505050565b600061097e6112cf565b33816000019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff1681525050888160400181815250508781606001818152505086816020019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508542610a0e9190611dc7565b816080018181525050600286610a249190611dfb565b42610a2f9190611dc7565b8160a0018181525050848160c0019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff1681525050838160e0018181525050600073ffffffffffffffffffffffffffffffffffffffff168160c0015173ffffffffffffffffffffffffffffffffffffffff1603610afd57348160e0015114610af8576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610aef90611ec7565b60405180910390fd5b610b87565b8060c0015173ffffffffffffffffffffffffffffffffffffffff166323b872dd33308460e001516040518463ffffffff1660e01b8152600401610b4293929190611ee7565b6020604051808303816000875af1158015610b61573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610b859190611ca5565b505b8281610100018181525050600081604051602001610ba591906119fc565b60405160208183030381529060405280519060200120905060006003811115610bd157610bd0611849565b5b60008083815260200190815260200160002060009054906101000a900460ff166003811115610c0357610c02611849565b5b14610c0d57600080fd5b7f91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56818b8b85608001518660a001518760c001518860e00151604051610c589796959493929190611f1e565b60405180910390a1600160008083815260200190815260200160002060006101000a81548160ff02191690836003811115610c9657610c95611849565b5b02179055508092505050979650505050505050565b60008060016000601b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179860001b7ffffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd036414180610d0757610d06611f8d565b5b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798890960001b60405160008152602001604052604051610d4a949392919061204c565b6020604051602081039080840390855afa158015610d6c573d6000803e3d6000fd5b5050506020604051035190508073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161491505092915050565b60006020528060005260406000206000915054906101000a900460ff1681565b600081604051602001610de391906119fc565b60405160208183030381529060405280519060200120905060016003811115610e0f57610e0e611849565b5b60008083815260200190815260200160002060009054906101000a900460ff166003811115610e4157610e40611849565b5b14610e81576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610e78906120dd565b60405180910390fd5b3373ffffffffffffffffffffffffffffffffffffffff16826000015173ffffffffffffffffffffffffffffffffffffffff1614610ef3576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610eea9061216f565b60405180910390fd5b600260008083815260200190815260200160002060006101000a81548160ff02191690836003811115610f2957610f28611849565b5b02179055507f5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f81604051610f5d91906117c1565b60405180910390a15050565b610f798260001c8260001c610cab565b610fb8576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610faf90612201565b60405180910390fd5b5050565b600082604051602001610fcf91906119fc565b604051602081830303815290604052805190602001209050600080600083815260200190815260200160002060009054906101000a900460ff1690506000600381111561101f5761101e611849565b5b81600381111561103257611031611849565b5b03611072576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016110699061226d565b60405180910390fd5b60038081111561108557611084611849565b5b81600381111561109857611097611849565b5b036110d8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016110cf90611a75565b60405180910390fd5b836020015173ffffffffffffffffffffffffffffffffffffffff166110fb611295565b73ffffffffffffffffffffffffffffffffffffffff1614611151576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611148906122d9565b60405180910390fd5b83608001514210158061118857506002600381111561117357611172611849565b5b81600381111561118657611185611849565b5b145b6111c7576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016111be90612345565b60405180910390fd5b8360a00151421061120d576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611204906123b1565b60405180910390fd5b61121b838560400151610f69565b7f38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee828460405161124c929190611bb9565b60405180910390a1600360008084815260200190815260200160002060006101000a81548160ff0219169083600381111561128a57611289611849565b5b021790555050505050565b60006112a0336105e3565b156112b457601436033560601c90506112c3565b6112bc6112c7565b90506112c4565b5b90565b600033905090565b604051806101200160405280600073ffffffffffffffffffffffffffffffffffffffff168152602001600073ffffffffffffffffffffffffffffffffffffffff16815260200160008019168152602001600080191681526020016000815260200160008152602001600073ffffffffffffffffffffffffffffffffffffffff16815260200160008152602001600081525090565b6000604051905090565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6113c082611377565b810181811067ffffffffffffffff821117156113df576113de611388565b5b80604052505050565b60006113f2611363565b90506113fe82826113b7565b919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600061142e82611403565b9050919050565b61143e81611423565b811461144957600080fd5b50565b60008135905061145b81611435565b92915050565b6000819050919050565b61147481611461565b811461147f57600080fd5b50565b6000813590506114918161146b565b92915050565b6000819050919050565b6114aa81611497565b81146114b557600080fd5b50565b6000813590506114c7816114a1565b92915050565b60006114d882611403565b9050919050565b6114e8816114cd565b81146114f357600080fd5b50565b600081359050611505816114df565b92915050565b6000610120828403121561152257611521611372565b5b61152d6101206113e8565b9050600061153d8482850161144c565b60008301525060206115518482850161144c565b602083015250604061156584828501611482565b604083015250606061157984828501611482565b606083015250608061158d848285016114b8565b60808301525060a06115a1848285016114b8565b60a08301525060c06115b5848285016114f6565b60c08301525060e06115c9848285016114b8565b60e0830152506101006115de848285016114b8565b6101008301525092915050565b60008061014083850312156116035761160261136d565b5b60006116118582860161150b565b92505061012061162385828601611482565b9150509250929050565b611636816114cd565b82525050565b6000602082019050611651600083018461162d565b92915050565b60006020828403121561166d5761166c61136d565b5b600061167b848285016114f6565b91505092915050565b60008115159050919050565b61169981611684565b82525050565b60006020820190506116b46000830184611690565b92915050565b600080600061016084860312156116d4576116d361136d565b5b60006116e28682870161150b565b9350506101206116f486828701611482565b925050610140611706868287016114b8565b9150509250925092565b600080600080600080600060e0888a03121561172f5761172e61136d565b5b600061173d8a828b01611482565b975050602061174e8a828b01611482565b965050604061175f8a828b0161144c565b95505060606117708a828b016114b8565b94505060806117818a828b016114f6565b93505060a06117928a828b016114b8565b92505060c06117a38a828b016114b8565b91505092959891949750929550565b6117bb81611461565b82525050565b60006020820190506117d660008301846117b2565b92915050565b600080604083850312156117f3576117f261136d565b5b6000611801858286016114b8565b9250506020611812858286016114b8565b9150509250929050565b6000602082840312156118325761183161136d565b5b600061184084828501611482565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b6004811061188957611888611849565b5b50565b600081905061189a82611878565b919050565b60006118aa8261188c565b9050919050565b6118ba8161189f565b82525050565b60006020820190506118d560008301846118b1565b92915050565b600061012082840312156118f2576118f161136d565b5b60006119008482850161150b565b91505092915050565b61191281611423565b82525050565b61192181611461565b82525050565b61193081611497565b82525050565b61193f816114cd565b82525050565b6101208201600082015161195c6000850182611909565b50602082015161196f6020850182611909565b5060408201516119826040850182611918565b5060608201516119956060850182611918565b5060808201516119a86080850182611927565b5060a08201516119bb60a0850182611927565b5060c08201516119ce60c0850182611936565b5060e08201516119e160e0850182611927565b506101008201516119f6610100850182611927565b50505050565b600061012082019050611a126000830184611945565b92915050565b600082825260208201905092915050565b7f7377617020697320616c726561647920636f6d706c6574656400000000000000600082015250565b6000611a5f601983611a18565b9150611a6a82611a29565b602082019050919050565b60006020820190508181036000830152611a8e81611a52565b9050919050565b7f726566756e64206d7573742062652063616c6c6564206279207468652073776160008201527f70206f776e657200000000000000000000000000000000000000000000000000602082015250565b6000611af1602783611a18565b9150611afc82611a95565b604082019050919050565b60006020820190508181036000830152611b2081611ae4565b9050919050565b7f697427732074686520636f756e74657270617274792773207475726e2c20756e60008201527f61626c6520746f20726566756e642c2074727920616761696e206c6174657200602082015250565b6000611b83603f83611a18565b9150611b8e82611b27565b604082019050919050565b60006020820190508181036000830152611bb281611b76565b9050919050565b6000604082019050611bce60008301856117b2565b611bdb60208301846117b2565b9392505050565b6000819050919050565b6000611c07611c02611bfd84611403565b611be2565b611403565b9050919050565b6000611c1982611bec565b9050919050565b6000611c2b82611c0e565b9050919050565b611c3b81611c20565b82525050565b611c4a81611497565b82525050565b6000604082019050611c656000830185611c32565b611c726020830184611c41565b9392505050565b611c8281611684565b8114611c8d57600080fd5b50565b600081519050611c9f81611c79565b92915050565b600060208284031215611cbb57611cba61136d565b5b6000611cc984828501611c90565b91505092915050565b7f636c61696d52656c617965722063616e206f6e6c792062652063616c6c65642060008201527f62792061207472757374656420666f7277617264657200000000000000000000602082015250565b6000611d2e603683611a18565b9150611d3982611cd2565b604082019050919050565b60006020820190508181036000830152611d5d81611d21565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000611d9e82611497565b9150611da983611497565b9250828203905081811115611dc157611dc0611d64565b5b92915050565b6000611dd282611497565b9150611ddd83611497565b9250828201905080821115611df557611df4611d64565b5b92915050565b6000611e0682611497565b9150611e1183611497565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0483118215151615611e4a57611e49611d64565b5b828202905092915050565b7f76616c7565206e6f742073616d652061732045544820616d6f756e742073656e60008201527f7400000000000000000000000000000000000000000000000000000000000000602082015250565b6000611eb1602183611a18565b9150611ebc82611e55565b604082019050919050565b60006020820190508181036000830152611ee081611ea4565b9050919050565b6000606082019050611efc600083018661162d565b611f09602083018561162d565b611f166040830184611c41565b949350505050565b600060e082019050611f33600083018a6117b2565b611f4060208301896117b2565b611f4d60408301886117b2565b611f5a6060830187611c41565b611f676080830186611c41565b611f7460a083018561162d565b611f8160c0830184611c41565b98975050505050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000819050919050565b60008160001b9050919050565b6000611fee611fe9611fe484611fbc565b611fc6565b611461565b9050919050565b611ffe81611fd3565b82525050565b6000819050919050565b600060ff82169050919050565b600061203661203161202c84612004565b611be2565b61200e565b9050919050565b6120468161201b565b82525050565b60006080820190506120616000830187611ff5565b61206e602083018661203d565b61207b60408301856117b2565b61208860608301846117b2565b95945050505050565b7f73776170206973206e6f7420696e2050454e44494e4720737461746500000000600082015250565b60006120c7601c83611a18565b91506120d282612091565b602082019050919050565b600060208201905081810360008301526120f6816120ba565b9050919050565b7f6f6e6c79207468652073776170206f776e65722063616e2063616c6c2073657460008201527f5265616479000000000000000000000000000000000000000000000000000000602082015250565b6000612159602583611a18565b9150612164826120fd565b604082019050919050565b600060208201905081810360008301526121888161214c565b9050919050565b7f70726f76696465642073656372657420646f6573206e6f74206d61746368207460008201527f6865206578706563746564207075626c6963206b657900000000000000000000602082015250565b60006121eb603683611a18565b91506121f68261218f565b604082019050919050565b6000602082019050818103600083015261221a816121de565b9050919050565b7f696e76616c696420737761700000000000000000000000000000000000000000600082015250565b6000612257600c83611a18565b915061226282612221565b602082019050919050565b600060208201905081810360008301526122868161224a565b9050919050565b7f6f6e6c7920636c61696d65722063616e20636c61696d21000000000000000000600082015250565b60006122c3601783611a18565b91506122ce8261228d565b602082019050919050565b600060208201905081810360008301526122f2816122b6565b9050919050565b7f746f6f206561726c7920746f20636c61696d2100000000000000000000000000600082015250565b600061232f601383611a18565b915061233a826122f9565b602082019050919050565b6000602082019050818103600083015261235e81612322565b9050919050565b7f746f6f206c61746520746f20636c61696d210000000000000000000000000000600082015250565b600061239b601283611a18565b91506123a682612365565b602082019050919050565b600060208201905081810360008301526123ca8161238e565b905091905056fea2646970667358221220bd1ca644fda0b57c5e313a660f5de8278282cd5e09b739ddb3ede270a025606864736f6c63430008100033" //nolint:lll

	ethAddrByteLen = len(ethcommon.Address{}) // 20 bytes
)

// Inside expectedSwapFactoryBytecodeHex, there are 2 locations where the trusted
// forwarder address, with which the contract was deployed, is embedded. When verifying
// the bytecode of a deployed contract, we need special treatment for these 2, identical
// 20-byte addresses at the start indexes below.
var forwarderAddressIndexes = []int{1485, 1523}

var errInvalidSwapContract = errors.New("given contract address does not contain correct code")

// CheckSwapFactoryContractCode checks that the bytecode at the given address matches that
// of SwapFactory.sol. The trusted forwarder address that the contract was deployed with
// is returned.
func CheckSwapFactoryContractCode(
	ctx context.Context,
	ec *ethclient.Client,
	contractAddr ethcommon.Address,
) (ethcommon.Address, error) {
	code, err := ec.CodeAt(ctx, contractAddr, nil)
	if err != nil {
		return ethcommon.Address{}, err
	}

	expectedCode := ethcommon.FromHex(expectedSwapFactoryBytecodeHex)

	if len(code) != len(expectedCode) {
		return ethcommon.Address{}, errInvalidSwapContract
	}

	allZeroAddr := ethcommon.Address{}

	// we fill this in with the trusted forwarder that the contract was deployed with
	var forwarderAddress ethcommon.Address

	for i, addrIndex := range forwarderAddressIndexes {
		curAddr := code[addrIndex : addrIndex+ethAddrByteLen]
		if i == 0 {
			// initialise the trusted forwarder address on the first index
			copy(forwarderAddress[:], curAddr)
		} else {
			// check that any remaining forwarder addresses match the one we found at the first index
			if !bytes.Equal(curAddr, forwarderAddress[:]) {
				return ethcommon.Address{}, errInvalidSwapContract
			}
		}
		// Zero out the trusted forwarder address in the code, so that we can compare the
		// read in byte code with a copy of the contract code that was deployed using an
		// all-zero trusted forwarder address. curAddr and code have the same backing
		// array, so we are updating expectedCode as well here:
		copy(curAddr, allZeroAddr[:])
	}
	// Now that the trusted forwarder addresses have been zeroed out, the read-in contract code should
	// match the expected code.
	if !bytes.Equal(expectedCode, code) {
		return ethcommon.Address{}, errInvalidSwapContract
	}
	// return the trusted forwarder address that was parsed from the deployed contract byte code
	return forwarderAddress, nil
}
