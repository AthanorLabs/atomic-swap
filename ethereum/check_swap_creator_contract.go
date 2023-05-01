// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// expectedSwapCreatorBytecodeHex is generated by deploying an instance of
// SwapCreator.sol and reading back the bytecode. See the unit test
// TestExpectedSwapCreatorBytecodeHex if you need to update this value.
const (
	expectedSwapCreatorBytecodeHex = "6080604052600436106100705760003560e01c8063b32d1b4f1161004e578063b32d1b4f146100d7578063c41e46cf1461010c578063eb84e7f21461012d578063fcaf229c1461016a57600080fd5b80631e6c5acc1461007557806356561693146100975780635cb96916146100b7575b600080fd5b34801561008157600080fd5b50610095610090366004610ed6565b61018a565b005b3480156100a357600080fd5b506100956100b2366004610f14565b6103d4565b3480156100c357600080fd5b506100956100d2366004610ed6565b6106b1565b3480156100e357600080fd5b506100f76100f2366004610fe2565b6107d0565b60405190151581526020015b60405180910390f35b61011f61011a366004611004565b6108a0565b604051908152602001610103565b34801561013957600080fd5b5061015d610148366004611075565b60006020819052908152604090205460ff1681565b60405161010391906110a4565b34801561017657600080fd5b506100956101853660046110cc565b610b82565b60008260405160200161019d9190611158565b60408051601f1981840301815291815281516020928301206000818152928390529082205490925060ff16908160038111156101db576101db61108e565b036101f957604051631115766760e01b815260040160405180910390fd5b600381600381111561020d5761020d61108e565b0361022b5760405163066916a960e01b815260040160405180910390fd5b83516001600160a01b031633146102555760405163148ca24360e11b815260040160405180910390fd5b8360a001514210801561028657508360800151421180610286575060028160038111156102845761028461108e565b145b156102a4576040516332a1860f60e11b815260040160405180910390fd5b6102b2838560600151610c60565b604051839083907e7c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f90600090a36000828152602081905260409020805460ff1916600317905560c08401516001600160a01b031661034c57835160e08501516040516001600160a01b039092169181156108fc0291906000818181858888f19350505050158015610346573d6000803e3d6000fd5b506103ce565b60c0840151845160e086015160405163a9059cbb60e01b81526001600160a01b039283166004820152602481019190915291169063a9059cbb906044016020604051808303816000875af11580156103a8573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103cc9190611167565b505b50505050565b60006001866040516020016103e99190611189565b60408051601f198184030181528282528051602091820120600084529083018083525260ff871690820152606081018590526080810184905260a0016020604051602081039080840390855afa158015610447573d6000803e3d6000fd5b5050506020604051035190508560000151602001516001600160a01b0316816001600160a01b03161461048d57604051638baa579f60e01b815260040160405180910390fd5b85606001516001600160a01b0316306001600160a01b0316146104c35760405163a710429d60e01b815260040160405180910390fd5b85516104cf9086610c87565b855160c001516001600160a01b031661057f578560000151602001516001600160a01b03166108fc8760200151886000015160e0015161050f91906111e7565b6040518115909202916000818181858888f19350505050158015610537573d6000803e3d6000fd5b5085604001516001600160a01b03166108fc87602001519081150290604051600060405180830381858888f19350505050158015610579573d6000803e3d6000fd5b506106a9565b855160c08101516020808301519089015160e0909301516001600160a01b039092169263a9059cbb926105b291906111e7565b6040516001600160e01b031960e085901b1681526001600160a01b03909216600483015260248201526044016020604051808303816000875af11580156105fd573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106219190611167565b50855160c001516040808801516020890151915163a9059cbb60e01b81526001600160a01b03918216600482015260248101929092529091169063a9059cbb906044016020604051808303816000875af1158015610683573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106a79190611167565b505b505050505050565b81602001516001600160a01b0316336001600160a01b0316146106e757604051633471640960e11b815260040160405180910390fd5b6106f18282610c87565b60c08201516001600160a01b03166107495781602001516001600160a01b03166108fc8360e001519081150290604051600060405180830381858888f19350505050158015610744573d6000803e3d6000fd5b505050565b60c0820151602083015160e084015160405163a9059cbb60e01b81526001600160a01b039283166004820152602481019190915291169063a9059cbb906044016020604051808303816000875af11580156107a8573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107449190611167565b5050565b600080600181601b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179870014551231950b75fc4402da1732fc9bebe197f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179889096040805160008152602081018083529590955260ff909316928401929092526060830152608082015260a0016020604051602081039080840390855afa15801561087d573d6000803e3d6000fd5b5050604051601f1901516001600160a01b03858116911614925050505b92915050565b6000826000036108c357604051637c946ed760e01b815260040160405180910390fd5b6001600160a01b0384166108f6573483146108f157604051632a9ffab760e21b815260040160405180910390fd5b61096f565b6040516323b872dd60e01b8152336004820152306024820152604481018490526001600160a01b038516906323b872dd906064016020604051808303816000875af1158015610949573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061096d9190611167565b505b88158061097a575087155b1561099857604051631bc61bed60e11b815260040160405180910390fd5b6001600160a01b0387166109be576040516208978560e71b815260040160405180910390fd5b8515806109c9575084155b156109e757604051631ffb86f160e21b815260040160405180910390fd5b6000604051806101200160405280336001600160a01b03168152602001896001600160a01b031681526020018b81526020018a81526020018842610a2b91906111fa565b815260200187610a3b8a426111fa565b610a4591906111fa565b8152602001866001600160a01b03168152602001858152602001848152509050600081604051602001610a789190611158565b60408051601f19818403018152919052805160209091012090506000808281526020819052604090205460ff166003811115610ab657610ab661108e565b14610ad4576040516339a2986760e11b815260040160405180910390fd5b7f91446ce035ac29998b5473504609a5ef5e961005daba4630a1684b63be848f56818c8c85608001518660a001518760c001518860e00151604051610b53979695949392919096875260208701959095526040860193909352606085019190915260808401526001600160a01b031660a083015260c082015260e00190565b60405180910390a16000818152602081905260409020805460ff191660011790559a9950505050505050505050565b600081604051602001610b959190611158565b60408051601f1981840301815291905280516020909101209050600160008281526020819052604090205460ff166003811115610bd457610bd461108e565b14610bf257604051630fe0fb5160e11b815260040160405180910390fd5b81516001600160a01b03163314610c1c5760405163148ca24360e11b815260040160405180910390fd5b600081815260208190526040808220805460ff191660021790555182917f5fc23b25552757626e08b316cc2387ad1bc70ee1594af7204db4ce0c39f5d15f91a25050565b610c6a82826107d0565b6107cc5760405163abab6bd760e01b815260040160405180910390fd5b600082604051602001610c9a9190611158565b60408051601f1981840301815291815281516020928301206000818152928390529082205490925060ff1690816003811115610cd857610cd861108e565b03610cf657604051631115766760e01b815260040160405180910390fd5b6003816003811115610d0a57610d0a61108e565b03610d285760405163066916a960e01b815260040160405180910390fd5b836080015142108015610d4d57506002816003811115610d4a57610d4a61108e565b14155b15610d6b5760405163d71d60b560e01b815260040160405180910390fd5b8360a001514210610d8f5760405163497df9d160e01b815260040160405180910390fd5b610d9d838560400151610c60565b604051839083907f38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee90600090a3506000908152602081905260409020805460ff191660031790555050565b604051610120810167ffffffffffffffff81118282101715610e1a57634e487b7160e01b600052604160045260246000fd5b60405290565b6001600160a01b0381168114610e3557600080fd5b50565b8035610e4381610e20565b919050565b60006101208284031215610e5b57600080fd5b610e63610de8565b9050610e6e82610e38565b8152610e7c60208301610e38565b602082015260408201356040820152606082013560608201526080820135608082015260a082013560a0820152610eb560c08301610e38565b60c082015260e082013560e082015261010080830135818301525092915050565b6000806101408385031215610eea57600080fd5b610ef48484610e48565b94610120939093013593505050565b803560ff81168114610e4357600080fd5b6000806000806000858703610200811215610f2e57600080fd5b61018080821215610f3e57600080fd5b60405191506080820182811067ffffffffffffffff82111715610f7157634e487b7160e01b600052604160045260246000fd5b604052610f7e8989610e48565b82526101208801356020830152610140880135610f9a81610e20565b6040830152610160880135610fae81610e20565b60608301529095508601359350610fc86101a08701610f03565b949793965093946101c081013594506101e0013592915050565b60008060408385031215610ff557600080fd5b50508035926020909101359150565b600080600080600080600080610100898b03121561102157600080fd5b8835975060208901359650604089013561103a81610e20565b9550606089013594506080890135935060a089013561105881610e20565b979a969950949793969295929450505060c08201359160e0013590565b60006020828403121561108757600080fd5b5035919050565b634e487b7160e01b600052602160045260246000fd5b60208101600483106110c657634e487b7160e01b600052602160045260246000fd5b91905290565b600061012082840312156110df57600080fd5b6110e98383610e48565b9392505050565b60018060a01b0380825116835280602083015116602084015260408201516040840152606082015160608401526080820151608084015260a082015160a08401528060c08301511660c08401525060e081015160e08301526101008082015181840152505050565b610120810161089a82846110f0565b60006020828403121561117957600080fd5b815180151581146110e957600080fd5b60006101808201905061119d8284516110f0565b602083015161012083015260408301516001600160a01b039081166101408401526060909301519092166101609091015290565b634e487b7160e01b600052601160045260246000fd5b8181038181111561089a5761089a6111d1565b8082018082111561089a5761089a6111d156fea26469706673582212206ac5abd4bdeadcd69a91cc110e06a7e6d6b455789d59db647a068af369a4623b64736f6c63430008130033" //nolint:lll
)

var (
	errInvalidSwapCreatorContract = errors.New("given contract address does not contain correct SwapCreator code")
)

// CheckSwapCreatorContractCode checks that the bytecode at the given address
// matches the SwapCreator.sol contract.
func CheckSwapCreatorContractCode(
	ctx context.Context,
	ec *ethclient.Client,
	contractAddr ethcommon.Address,
) error {
	code, err := ec.CodeAt(ctx, contractAddr, nil)
	if err != nil {
		return fmt.Errorf("failed to get code at %s: %w", contractAddr, err)
	}

	expectedCode := ethcommon.FromHex(expectedSwapCreatorBytecodeHex)

	if len(code) != len(expectedCode) {
		return fmt.Errorf("length mismatch: %w", errInvalidSwapCreatorContract)
	}

	if !bytes.Equal(expectedCode, code) {
		return errInvalidSwapCreatorContract
	}

	return nil
}
