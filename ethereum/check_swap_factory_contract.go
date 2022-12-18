package contracts

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
)

// expectedSwapFactoryBytecodeHex is generated by deploying an instance of SwapFactory.sol
// with the trustedForwarder address set to all zeros and reading back the bytecode. See
// the unit test TestExpectedSwapFactoryBytecodeHex if you need to update this value.
const (
	expectedSwapFactoryBytecodeHex = "6080604052600436106100865760003560e01c806373e4771c1161005957806373e4771c14610145578063aa0f87251461016e578063b32d1b4f1461019e578063eb84e7f2146101db578063fcaf229c1461021857610086565b80631e6c5acc1461008b57806356c022bb146100b4578063572b6c05146100df5780635cb969161461011c575b600080fd5b34801561009757600080fd5b506100b260048036038101906100ad919061165d565b610241565b005b3480156100c057600080fd5b506100c96105bf565b6040516100d691906116ae565b60405180910390f35b3480156100eb57600080fd5b50610106600480360381019061010191906116c9565b6105e3565b6040516101139190611711565b60405180910390f35b34801561012857600080fd5b50610143600480360381019061013e919061165d565b61063b565b005b34801561015157600080fd5b5061016c6004803603810190610167919061172c565b610761565b005b61018860048036038101906101839190611782565b6109b0565b6040516101959190611833565b60405180910390f35b3480156101aa57600080fd5b506101c560048036038101906101c0919061184e565b610d1d565b6040516101d29190611711565b60405180910390f35b3480156101e757600080fd5b5061020260048036038101906101fd919061188e565b610e22565b60405161020f9190611932565b60405180910390f35b34801561022457600080fd5b5061023f600480360381019061023a919061194d565b610e42565b005b6000826040516020016102549190611a6e565b604051602081830303815290604052805190602001209050600080600083815260200190815260200160002060009054906101000a900460ff1690506003808111156102a3576102a26118bb565b5b8160038111156102b6576102b56118bb565b5b141580156102e95750600060038111156102d3576102d26118bb565b5b8160038111156102e6576102e56118bb565b5b14155b610328576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161031f90611ae7565b60405180910390fd5b836000015173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461039a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161039190611b79565b60405180910390fd5b8360a00151421015806103e157508360800151421080156103e05750600260038111156103ca576103c96118bb565b5b8160038111156103dd576103dc6118bb565b5b14155b5b610420576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161041790611c0b565b60405180910390fd5b61042e838560600151610fdb565b7e7c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f828460405161045e929190611c2b565b60405180910390a1600360008084815260200190815260200160002060006101000a81548160ff0219169083600381111561049c5761049b6118bb565b5b0217905550600073ffffffffffffffffffffffffffffffffffffffff168460c0015173ffffffffffffffffffffffffffffffffffffffff160361052d57836000015173ffffffffffffffffffffffffffffffffffffffff166108fc8560e001519081150290604051600060405180830381858888f19350505050158015610527573d6000803e3d6000fd5b506105b9565b8360c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb85600001518660e001516040518363ffffffff1660e01b8152600401610574929190611cc2565b6020604051808303816000875af1158015610593573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105b79190611d17565b505b50505050565b7f000000000000000000000000000000000000000000000000000000000000000081565b60007f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16149050919050565b610645828261102e565b600073ffffffffffffffffffffffffffffffffffffffff168260c0015173ffffffffffffffffffffffffffffffffffffffff16036106d157816020015173ffffffffffffffffffffffffffffffffffffffff166108fc8360e001519081150290604051600060405180830381858888f193505050501580156106cb573d6000803e3d6000fd5b5061075d565b8160c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb83602001518460e001516040518363ffffffff1660e01b8152600401610718929190611cc2565b6020604051808303816000875af1158015610737573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061075b9190611d17565b505b5050565b61076a336105e3565b6107a9576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107a090611db6565b60405180910390fd5b6107b3838361102e565b600073ffffffffffffffffffffffffffffffffffffffff168360c0015173ffffffffffffffffffffffffffffffffffffffff160361089157826020015173ffffffffffffffffffffffffffffffffffffffff166108fc828560e001516108199190611e05565b9081150290604051600060405180830381858888f19350505050158015610844573d6000803e3d6000fd5b503273ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f1935050505015801561088b573d6000803e3d6000fd5b506109ab565b8260c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb8460200151838660e001516108c69190611e05565b6040518363ffffffff1660e01b81526004016108e3929190611cc2565b6020604051808303816000875af1158015610902573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109269190611d17565b508260c0015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb32836040518363ffffffff1660e01b8152600401610966929190611e39565b6020604051808303816000875af1158015610985573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109a99190611d17565b505b505050565b60006109ba611341565b33816000019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff1681525050888160400181815250508781606001818152505086816020019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508542610a4a9190611e62565b816080018181525050600286610a609190611e96565b42610a6b9190611e62565b8160a0018181525050848160c0019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff1681525050838160e0018181525050600073ffffffffffffffffffffffffffffffffffffffff168160c0015173ffffffffffffffffffffffffffffffffffffffff1603610b3957348160e0015114610b34576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b2b90611f62565b60405180910390fd5b610bc3565b8060c0015173ffffffffffffffffffffffffffffffffffffffff166323b872dd33308460e001516040518463ffffffff1660e01b8152600401610b7e93929190611f82565b6020604051808303816000875af1158015610b9d573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610bc19190611d17565b505b8281610100018181525050600081604051602001610be19190611a6e565b60405160208183030381529060405280519060200120905060006003811115610c0d57610c0c6118bb565b5b60008083815260200190815260200160002060009054906101000a900460ff166003811115610c3f57610c3e6118bb565b5b14610c7f576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610c7690612005565b60405180910390fd5b7f91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56818b8b85608001518660a001518760c001518860e00151604051610cca9796959493929190612025565b60405180910390a1600160008083815260200190815260200160002060006101000a81548160ff02191690836003811115610d0857610d076118bb565b5b02179055508092505050979650505050505050565b60008060016000601b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179860001b7ffffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd036414180610d7957610d78612094565b5b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798890960001b60405160008152602001604052604051610dbc9493929190612153565b6020604051602081039080840390855afa158015610dde573d6000803e3d6000fd5b5050506020604051035190508073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161491505092915050565b60006020528060005260406000206000915054906101000a900460ff1681565b600081604051602001610e559190611a6e565b60405160208183030381529060405280519060200120905060016003811115610e8157610e806118bb565b5b60008083815260200190815260200160002060009054906101000a900460ff166003811115610eb357610eb26118bb565b5b14610ef3576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610eea906121e4565b60405180910390fd5b3373ffffffffffffffffffffffffffffffffffffffff16826000015173ffffffffffffffffffffffffffffffffffffffff1614610f65576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610f5c90612276565b60405180910390fd5b600260008083815260200190815260200160002060006101000a81548160ff02191690836003811115610f9b57610f9a6118bb565b5b02179055507f5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f81604051610fcf9190611833565b60405180910390a15050565b610feb8260001c8260001c610d1d565b61102a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161102190612308565b60405180910390fd5b5050565b6000826040516020016110419190611a6e565b604051602081830303815290604052805190602001209050600080600083815260200190815260200160002060009054906101000a900460ff16905060006003811115611091576110906118bb565b5b8160038111156110a4576110a36118bb565b5b036110e4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016110db90612374565b60405180910390fd5b6003808111156110f7576110f66118bb565b5b81600381111561110a576111096118bb565b5b0361114a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161114190611ae7565b60405180910390fd5b836020015173ffffffffffffffffffffffffffffffffffffffff1661116d611307565b73ffffffffffffffffffffffffffffffffffffffff16146111c3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016111ba906123e0565b60405180910390fd5b8360800151421015806111fa5750600260038111156111e5576111e46118bb565b5b8160038111156111f8576111f76118bb565b5b145b611239576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016112309061244c565b60405180910390fd5b8360a00151421061127f576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611276906124b8565b60405180910390fd5b61128d838560400151610fdb565b7f38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee82846040516112be929190611c2b565b60405180910390a1600360008084815260200190815260200160002060006101000a81548160ff021916908360038111156112fc576112fb6118bb565b5b021790555050505050565b6000611312336105e3565b1561132657601436033560601c9050611335565b61132e611339565b9050611336565b5b90565b600033905090565b604051806101200160405280600073ffffffffffffffffffffffffffffffffffffffff168152602001600073ffffffffffffffffffffffffffffffffffffffff16815260200160008019168152602001600080191681526020016000815260200160008152602001600073ffffffffffffffffffffffffffffffffffffffff16815260200160008152602001600081525090565b6000604051905090565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b611432826113e9565b810181811067ffffffffffffffff82111715611451576114506113fa565b5b80604052505050565b60006114646113d5565b90506114708282611429565b919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006114a082611475565b9050919050565b6114b081611495565b81146114bb57600080fd5b50565b6000813590506114cd816114a7565b92915050565b6000819050919050565b6114e6816114d3565b81146114f157600080fd5b50565b600081359050611503816114dd565b92915050565b6000819050919050565b61151c81611509565b811461152757600080fd5b50565b60008135905061153981611513565b92915050565b600061154a82611475565b9050919050565b61155a8161153f565b811461156557600080fd5b50565b60008135905061157781611551565b92915050565b60006101208284031215611594576115936113e4565b5b61159f61012061145a565b905060006115af848285016114be565b60008301525060206115c3848285016114be565b60208301525060406115d7848285016114f4565b60408301525060606115eb848285016114f4565b60608301525060806115ff8482850161152a565b60808301525060a06116138482850161152a565b60a08301525060c061162784828501611568565b60c08301525060e061163b8482850161152a565b60e0830152506101006116508482850161152a565b6101008301525092915050565b6000806101408385031215611675576116746113df565b5b60006116838582860161157d565b925050610120611695858286016114f4565b9150509250929050565b6116a88161153f565b82525050565b60006020820190506116c3600083018461169f565b92915050565b6000602082840312156116df576116de6113df565b5b60006116ed84828501611568565b91505092915050565b60008115159050919050565b61170b816116f6565b82525050565b60006020820190506117266000830184611702565b92915050565b60008060006101608486031215611746576117456113df565b5b60006117548682870161157d565b935050610120611766868287016114f4565b9250506101406117788682870161152a565b9150509250925092565b600080600080600080600060e0888a0312156117a1576117a06113df565b5b60006117af8a828b016114f4565b97505060206117c08a828b016114f4565b96505060406117d18a828b016114be565b95505060606117e28a828b0161152a565b94505060806117f38a828b01611568565b93505060a06118048a828b0161152a565b92505060c06118158a828b0161152a565b91505092959891949750929550565b61182d816114d3565b82525050565b60006020820190506118486000830184611824565b92915050565b60008060408385031215611865576118646113df565b5b60006118738582860161152a565b92505060206118848582860161152a565b9150509250929050565b6000602082840312156118a4576118a36113df565b5b60006118b2848285016114f4565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b600481106118fb576118fa6118bb565b5b50565b600081905061190c826118ea565b919050565b600061191c826118fe565b9050919050565b61192c81611911565b82525050565b60006020820190506119476000830184611923565b92915050565b60006101208284031215611964576119636113df565b5b60006119728482850161157d565b91505092915050565b61198481611495565b82525050565b611993816114d3565b82525050565b6119a281611509565b82525050565b6119b18161153f565b82525050565b610120820160008201516119ce600085018261197b565b5060208201516119e1602085018261197b565b5060408201516119f4604085018261198a565b506060820151611a07606085018261198a565b506080820151611a1a6080850182611999565b5060a0820151611a2d60a0850182611999565b5060c0820151611a4060c08501826119a8565b5060e0820151611a5360e0850182611999565b50610100820151611a68610100850182611999565b50505050565b600061012082019050611a8460008301846119b7565b92915050565b600082825260208201905092915050565b7f7377617020697320616c726561647920636f6d706c6574656400000000000000600082015250565b6000611ad1601983611a8a565b9150611adc82611a9b565b602082019050919050565b60006020820190508181036000830152611b0081611ac4565b9050919050565b7f726566756e64206d7573742062652063616c6c6564206279207468652073776160008201527f70206f776e657200000000000000000000000000000000000000000000000000602082015250565b6000611b63602783611a8a565b9150611b6e82611b07565b604082019050919050565b60006020820190508181036000830152611b9281611b56565b9050919050565b7f697427732074686520636f756e74657270617274792773207475726e2c20756e60008201527f61626c6520746f20726566756e642c2074727920616761696e206c6174657200602082015250565b6000611bf5603f83611a8a565b9150611c0082611b99565b604082019050919050565b60006020820190508181036000830152611c2481611be8565b9050919050565b6000604082019050611c406000830185611824565b611c4d6020830184611824565b9392505050565b6000819050919050565b6000611c79611c74611c6f84611475565b611c54565b611475565b9050919050565b6000611c8b82611c5e565b9050919050565b6000611c9d82611c80565b9050919050565b611cad81611c92565b82525050565b611cbc81611509565b82525050565b6000604082019050611cd76000830185611ca4565b611ce46020830184611cb3565b9392505050565b611cf4816116f6565b8114611cff57600080fd5b50565b600081519050611d1181611ceb565b92915050565b600060208284031215611d2d57611d2c6113df565b5b6000611d3b84828501611d02565b91505092915050565b7f636c61696d52656c617965722063616e206f6e6c792062652063616c6c65642060008201527f62792061207472757374656420666f7277617264657200000000000000000000602082015250565b6000611da0603683611a8a565b9150611dab82611d44565b604082019050919050565b60006020820190508181036000830152611dcf81611d93565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000611e1082611509565b9150611e1b83611509565b9250828203905081811115611e3357611e32611dd6565b5b92915050565b6000604082019050611e4e600083018561169f565b611e5b6020830184611cb3565b9392505050565b6000611e6d82611509565b9150611e7883611509565b9250828201905080821115611e9057611e8f611dd6565b5b92915050565b6000611ea182611509565b9150611eac83611509565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0483118215151615611ee557611ee4611dd6565b5b828202905092915050565b7f76616c7565206e6f742073616d652061732045544820616d6f756e742073656e60008201527f7400000000000000000000000000000000000000000000000000000000000000602082015250565b6000611f4c602183611a8a565b9150611f5782611ef0565b604082019050919050565b60006020820190508181036000830152611f7b81611f3f565b9050919050565b6000606082019050611f97600083018661169f565b611fa4602083018561169f565b611fb16040830184611cb3565b949350505050565b7f7377617020616c72656164792065786973747300000000000000000000000000600082015250565b6000611fef601383611a8a565b9150611ffa82611fb9565b602082019050919050565b6000602082019050818103600083015261201e81611fe2565b9050919050565b600060e08201905061203a600083018a611824565b6120476020830189611824565b6120546040830188611824565b6120616060830187611cb3565b61206e6080830186611cb3565b61207b60a083018561169f565b61208860c0830184611cb3565b98975050505050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000819050919050565b60008160001b9050919050565b60006120f56120f06120eb846120c3565b6120cd565b6114d3565b9050919050565b612105816120da565b82525050565b6000819050919050565b600060ff82169050919050565b600061213d6121386121338461210b565b611c54565b612115565b9050919050565b61214d81612122565b82525050565b600060808201905061216860008301876120fc565b6121756020830186612144565b6121826040830185611824565b61218f6060830184611824565b95945050505050565b7f73776170206973206e6f7420696e2050454e44494e4720737461746500000000600082015250565b60006121ce601c83611a8a565b91506121d982612198565b602082019050919050565b600060208201905081810360008301526121fd816121c1565b9050919050565b7f6f6e6c79207468652073776170206f776e65722063616e2063616c6c2073657460008201527f5265616479000000000000000000000000000000000000000000000000000000602082015250565b6000612260602583611a8a565b915061226b82612204565b604082019050919050565b6000602082019050818103600083015261228f81612253565b9050919050565b7f70726f76696465642073656372657420646f6573206e6f74206d61746368207460008201527f6865206578706563746564207075626c6963206b657900000000000000000000602082015250565b60006122f2603683611a8a565b91506122fd82612296565b604082019050919050565b60006020820190508181036000830152612321816122e5565b9050919050565b7f696e76616c696420737761700000000000000000000000000000000000000000600082015250565b600061235e600c83611a8a565b915061236982612328565b602082019050919050565b6000602082019050818103600083015261238d81612351565b9050919050565b7f6f6e6c7920636c61696d65722063616e20636c61696d21000000000000000000600082015250565b60006123ca601783611a8a565b91506123d582612394565b602082019050919050565b600060208201905081810360008301526123f9816123bd565b9050919050565b7f746f6f206561726c7920746f20636c61696d2100000000000000000000000000600082015250565b6000612436601383611a8a565b915061244182612400565b602082019050919050565b6000602082019050818103600083015261246581612429565b9050919050565b7f746f6f206c61746520746f20636c61696d210000000000000000000000000000600082015250565b60006124a2601283611a8a565b91506124ad8261246c565b602082019050919050565b600060208201905081810360008301526124d181612495565b905091905056fea2646970667358221220017dc5e8e8bb60a0ec1cd23175097f5c7af0308f62998c48f6294ec7d5cc16d764736f6c63430008100033" //nolint:lll

	ethAddrByteLen = len(ethcommon.Address{}) // 20 bytes
)

// forwarderAddressIndices is a slice of the start indices where the trusted forwarder
// address is compiled into the deployed contract byte code. When verifying the bytecode
// of a deployed contract, we need special treatment for these identical 20-byte address
// blocks. See TestForwarderAddressIndexes to update the values.
var forwarderAddressIndices = []int{1485, 1523}

var (
	errInvalidSwapContract      = errors.New("given contract address does not contain correct SwapFactory code")
	errInvalidForwarderContract = errors.New("given contract address does not contain correct Forwarder code")
)

// CheckSwapFactoryContractCode checks that the bytecode at the given address matches the
// SwapFactory.sol contract. The trusted forwarder address that the contract was deployed
// with is parsed out from the byte code and returned.
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
		return ethcommon.Address{}, fmt.Errorf("length mismatch: %w", errInvalidSwapContract)
	}

	allZeroAddr := ethcommon.Address{}

	// we fill this in with the trusted forwarder that the contract was deployed with
	var forwarderAddress ethcommon.Address

	for i, addrIndex := range forwarderAddressIndices {
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

	if (forwarderAddress == ethcommon.Address{}) {
		return forwarderAddress, nil
	}

	err = checkForwarderContractCode(ctx, ec, forwarderAddress)
	if err != nil {
		return ethcommon.Address{}, err
	}

	// return the trusted forwarder address that was parsed from the deployed contract byte code
	return forwarderAddress, nil
}

// checkSwapFactoryForwarder checks that the trusted forwarder contract used by
// the given swap contract has the expected bytecode.
func checkForwarderContractCode(
	ctx context.Context,
	ec *ethclient.Client,
	contractAddr ethcommon.Address,
) error {
	code, err := ec.CodeAt(ctx, contractAddr, nil)
	if err != nil {
		return err
	}

	expectedCode := ethcommon.FromHex(gsnforwarder.ForwarderMetaData.Bin)

	// expectedCode is the compiled code, while code is the deployed bytecode.
	// the deployed bytecode is a subset of the compiled code.
	if !bytes.Equal(expectedCode[705:9585], code) {
		return errInvalidForwarderContract
	}

	return nil
}
