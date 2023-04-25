// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/athanorlabs/atomic-swap/common"

	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// expectedSwapCreatorBytecodeHex is generated by deploying an instance of SwapCreator.sol
// with the trustedForwarder address set to all zeros and reading back the bytecode. See
// the unit test TestExpectedSwapCreatorBytecodeHex if you need to update this value.
const (
	expectedSwapCreatorBytecodeHex = "6080604052600436106100865760003560e01c806373e4771c1161005957806373e4771c1461014e578063b32d1b4f1461016e578063c41e46cf1461018e578063eb84e7f2146101af578063fcaf229c146101ec57600080fd5b80631e6c5acc1461008b57806356c022bb146100ad578063572b6c05146100fe5780635cb969161461012e575b600080fd5b34801561009757600080fd5b506100ab6100a6366004610e9e565b61020c565b005b3480156100b957600080fd5b506100e17f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b0390911681526020015b60405180910390f35b34801561010a57600080fd5b5061011e610119366004610ecb565b610456565b60405190151581526020016100f5565b34801561013a57600080fd5b506100ab610149366004610e9e565b610488565b34801561015a57600080fd5b506100ab610169366004610eef565b610571565b34801561017a57600080fd5b5061011e610189366004610f26565b610739565b6101a161019c366004610f48565b610809565b6040519081526020016100f5565b3480156101bb57600080fd5b506101df6101ca366004610fb9565b60006020819052908152604090205460ff1681565b6040516100f59190610fe8565b3480156101f857600080fd5b506100ab610207366004611010565b610aeb565b60008260405160200161021f919061102d565b60408051601f1981840301815291815281516020928301206000818152928390529082205490925060ff169081600381111561025d5761025d610fd2565b0361027b57604051631115766760e01b815260040160405180910390fd5b600381600381111561028f5761028f610fd2565b036102ad5760405163066916a960e01b815260040160405180910390fd5b83516001600160a01b031633146102d75760405163148ca24360e11b815260040160405180910390fd5b8360a0015142108015610308575083608001514211806103085750600281600381111561030657610306610fd2565b145b15610326576040516332a1860f60e11b815260040160405180910390fd5b610334838560600151610bc9565b604051839083907e7c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f90600090a36000828152602081905260409020805460ff1916600317905560c08401516001600160a01b03166103ce57835160e08501516040516001600160a01b039092169181156108fc0291906000818181858888f193505050501580156103c8573d6000803e3d6000fd5b50610450565b60c0840151845160e086015160405163a9059cbb60e01b81526001600160a01b039283166004820152602481019190915291169063a9059cbb906044016020604051808303816000875af115801561042a573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061044e919061109c565b505b50505050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0390811691161490565b6104928282610bf0565b60c08201516001600160a01b03166104ea5781602001516001600160a01b03166108fc8360e001519081150290604051600060405180830381858888f193505050501580156104e5573d6000803e3d6000fd5b505050565b60c0820151602083015160e084015160405163a9059cbb60e01b81526001600160a01b039283166004820152602481019190915291169063a9059cbb906044016020604051808303816000875af1158015610549573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906104e5919061109c565b5050565b61057a33610456565b61059757604051637e2ea6d560e11b815260040160405180910390fd5b6105a18383610bf0565b60c08301516001600160a01b031661062a5782602001516001600160a01b03166108fc828560e001516105d491906110d4565b6040518115909202916000818181858888f193505050501580156105fc573d6000803e3d6000fd5b50604051329082156108fc029083906000818181858888f19350505050158015610450573d6000803e3d6000fd5b8260c001516001600160a01b031663a9059cbb8460200151838660e0015161065291906110d4565b6040516001600160e01b031960e085901b1681526001600160a01b03909216600483015260248201526044016020604051808303816000875af115801561069d573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106c1919061109c565b5060c083015160405163a9059cbb60e01b8152326004820152602481018390526001600160a01b039091169063a9059cbb906044016020604051808303816000875af1158015610715573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610450919061109c565b600080600181601b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179870014551231950b75fc4402da1732fc9bebe197f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179889096040805160008152602081018083529590955260ff909316928401929092526060830152608082015260a0016020604051602081039080840390855afa1580156107e6573d6000803e3d6000fd5b5050604051601f1901516001600160a01b03858116911614925050505b92915050565b60008260000361082c57604051637c946ed760e01b815260040160405180910390fd5b6001600160a01b03841661085f5734831461085a57604051632a9ffab760e21b815260040160405180910390fd5b6108d8565b6040516323b872dd60e01b8152336004820152306024820152604481018490526001600160a01b038516906323b872dd906064016020604051808303816000875af11580156108b2573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906108d6919061109c565b505b8815806108e3575087155b1561090157604051631bc61bed60e11b815260040160405180910390fd5b6001600160a01b038716610927576040516208978560e71b815260040160405180910390fd5b851580610932575084155b1561095057604051631ffb86f160e21b815260040160405180910390fd5b6000604051806101200160405280336001600160a01b03168152602001896001600160a01b031681526020018b81526020018a8152602001884261099491906110e7565b8152602001876109a48a426110e7565b6109ae91906110e7565b8152602001866001600160a01b031681526020018581526020018481525090506000816040516020016109e1919061102d565b60408051601f19818403018152919052805160209091012090506000808281526020819052604090205460ff166003811115610a1f57610a1f610fd2565b14610a3d576040516339a2986760e11b815260040160405180910390fd5b7f91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56818c8c85608001518660a001518760c001518860e00151604051610abc979695949392919096875260208701959095526040860193909352606085019190915260808401526001600160a01b031660a083015260c082015260e00190565b60405180910390a16000818152602081905260409020805460ff191660011790559a9950505050505050505050565b600081604051602001610afe919061102d565b60408051601f1981840301815291905280516020909101209050600160008281526020819052604090205460ff166003811115610b3d57610b3d610fd2565b14610b5b57604051630fe0fb5160e11b815260040160405180910390fd5b81516001600160a01b03163314610b855760405163148ca24360e11b815260040160405180910390fd5b600081815260208190526040808220805460ff191660021790555182917f5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f91a25050565b610bd38282610739565b61056d5760405163abab6bd760e01b815260040160405180910390fd5b600082604051602001610c03919061102d565b60408051601f1981840301815291815281516020928301206000818152928390529082205490925060ff1690816003811115610c4157610c41610fd2565b03610c5f57604051631115766760e01b815260040160405180910390fd5b6003816003811115610c7357610c73610fd2565b03610c915760405163066916a960e01b815260040160405180910390fd5b83602001516001600160a01b0316610ca7610d8e565b6001600160a01b031614610cce57604051633471640960e11b815260040160405180910390fd5b836080015142108015610cf357506002816003811115610cf057610cf0610fd2565b14155b15610d115760405163d71d60b560e01b815260040160405180910390fd5b8360a001514210610d355760405163497df9d160e01b815260040160405180910390fd5b610d43838560400151610bc9565b604051839083907f38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee90600090a3506000908152602081905260409020805460ff191660031790555050565b6000610d9933610456565b15610dab575060131936013560601c90565b503390565b604051610120810167ffffffffffffffff81118282101715610de257634e487b7160e01b600052604160045260246000fd5b60405290565b6001600160a01b0381168114610dfd57600080fd5b50565b8035610e0b81610de8565b919050565b60006101208284031215610e2357600080fd5b610e2b610db0565b9050610e3682610e00565b8152610e4460208301610e00565b602082015260408201356040820152606082013560608201526080820135608082015260a082013560a0820152610e7d60c08301610e00565b60c082015260e082013560e082015261010080830135818301525092915050565b6000806101408385031215610eb257600080fd5b610ebc8484610e10565b94610120939093013593505050565b600060208284031215610edd57600080fd5b8135610ee881610de8565b9392505050565b60008060006101608486031215610f0557600080fd5b610f0f8585610e10565b956101208501359550610140909401359392505050565b60008060408385031215610f3957600080fd5b50508035926020909101359150565b600080600080600080600080610100898b031215610f6557600080fd5b88359750602089013596506040890135610f7e81610de8565b9550606089013594506080890135935060a0890135610f9c81610de8565b979a969950949793969295929450505060c08201359160e0013590565b600060208284031215610fcb57600080fd5b5035919050565b634e487b7160e01b600052602160045260246000fd5b602081016004831061100a57634e487b7160e01b600052602160045260246000fd5b91905290565b6000610120828403121561102357600080fd5b610ee88383610e10565b81516001600160a01b03908116825260208084015182169083015260408084015190830152606080840151908301526080808401519083015260a0808401519083015260c0808401519091169082015260e0808301519082015261010091820151918101919091526101200190565b6000602082840312156110ae57600080fd5b81518015158114610ee857600080fd5b634e487b7160e01b600052601160045260246000fd5b81810381811115610803576108036110be565b80820180821115610803576108036110be56fea2646970667358221220a75326d41574d36189871c40b37894bb93ca35029fb6761e949335295c76985064736f6c63430008130033" //nolint:lll

	ethAddrByteLen = len(ethcommon.Address{}) // 20 bytes
)

// forwarderAddrIndices is a slice of the start indices where the trusted forwarder
// address is compiled into the deployed contract byte code. When verifying the bytecode
// of a deployed contract, we need special treatment for these identical 20-byte address
// blocks. See TestForwarderAddrIndexes to update the values.
var forwarderAddrIndices = []int{203, 1124}

var (
	errInvalidSwapCreatorContract = errors.New("given contract address does not contain correct SwapCreator code")
	errInvalidForwarderContract   = errors.New("given contract address does not contain correct Forwarder code")
)

// CheckSwapCreatorContractCode checks that the bytecode at the given address matches the
// SwapCreator.sol contract. The trusted forwarder address that the contract was deployed
// with is parsed out from the byte code and returned.
func CheckSwapCreatorContractCode(
	ctx context.Context,
	ec *ethclient.Client,
	contractAddr ethcommon.Address,
) (ethcommon.Address, error) {
	code, err := ec.CodeAt(ctx, contractAddr, nil)
	if err != nil {
		return ethcommon.Address{}, err
	}

	expectedCode := ethcommon.FromHex(expectedSwapCreatorBytecodeHex)

	if len(code) != len(expectedCode) {
		return ethcommon.Address{}, fmt.Errorf("length mismatch: %w", errInvalidSwapCreatorContract)
	}

	allZeroAddr := ethcommon.Address{}

	// we fill this in with the trusted forwarder that the contract was deployed with
	var forwarderAddr ethcommon.Address

	for i, addrIndex := range forwarderAddrIndices {
		curAddr := code[addrIndex : addrIndex+ethAddrByteLen]
		if i == 0 {
			// initialise the trusted forwarder address on the first index
			copy(forwarderAddr[:], curAddr)
		} else {
			// check that any remaining forwarder addresses match the one we found at the first index
			if !bytes.Equal(curAddr, forwarderAddr[:]) {
				return ethcommon.Address{}, errInvalidSwapCreatorContract
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
		return ethcommon.Address{}, errInvalidSwapCreatorContract
	}

	if (forwarderAddr == ethcommon.Address{}) {
		return forwarderAddr, nil
	}

	err = CheckForwarderContractCode(ctx, ec, forwarderAddr)
	if err != nil {
		return ethcommon.Address{}, err
	}

	// return the trusted forwarder address that was parsed from the deployed contract byte code
	return forwarderAddr, nil
}

// CheckForwarderContractCode checks that the trusted forwarder contract used by
// the given swap contract has the expected bytecode.
func CheckForwarderContractCode(
	ctx context.Context,
	ec *ethclient.Client,
	contractAddr ethcommon.Address,
) error {
	// mainnet override - since the forwarder contract deployed on mainnet is compiled
	// with solidity 0.8.7, but we're using 0.8.19 for SwapCreator.sol, we can just
	// check that the address is what's expected.
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return err
	}

	if contractAddr == common.MainnetConfig().ForwarderAddr && chainID.Uint64() == common.MainnetChainID {
		return nil
	}

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
