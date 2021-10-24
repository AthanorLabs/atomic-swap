// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package swapDLEQ

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

// SwapDLEQMetaData contains all meta data concerning the SwapDLEQ contract.
var SwapDLEQMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_pubKeyClaimX\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_pubKeyClaimY\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_pubKeyRefundX\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_pubKeyRefundY\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"s\",\"type\":\"uint256\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"p\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"q\",\"type\":\"uint256\"}],\"name\":\"Constructed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"b\",\"type\":\"bool\"}],\"name\":\"IsReady\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"s\",\"type\":\"uint256\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_s\",\"type\":\"uint256\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pubKeyClaimX\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pubKeyClaimY\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pubKeyRefundX\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pubKeyRefundY\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_s\",\"type\":\"uint256\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"set_ready\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeout_0\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeout_1\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x6101606040526000600160006101000a81548160ff02191690831515021790555060405162002542380380620025428339818101604052810190620000459190620001a6565b3373ffffffffffffffffffffffffffffffffffffffff1660a08173ffffffffffffffffffffffffffffffffffffffff16815250508360c081815250508260e081815250508161010081815250508061012081815250506201518042620000ac919062000247565b6101408181525050604051620000c29062000158565b604051809103906000f080158015620000df573d6000803e3d6000fd5b5073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff16815250507fbc91786c56ba12548733c92d8dc8bc6e725d7ac2eb0c3039fddbcfbf19a852a9828260405162000146929190620002b5565b60405180910390a150505050620002e2565b61151d806200102583390190565b600080fd5b6000819050919050565b62000180816200016b565b81146200018c57600080fd5b50565b600081519050620001a08162000175565b92915050565b60008060008060808587031215620001c357620001c262000166565b5b6000620001d3878288016200018f565b9450506020620001e6878288016200018f565b9350506040620001f9878288016200018f565b92505060606200020c878288016200018f565b91505092959194509250565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600062000254826200016b565b915062000261836200016b565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0382111562000299576200029862000218565b5b828201905092915050565b620002af816200016b565b82525050565b6000604082019050620002cc6000830185620002a4565b620002db6020830184620002a4565b9392505050565b60805160a05160c05160e051610100516101205161014051610ca96200037c6000396000818161021a0152818161047e01528181610664015261071b0152600081816102db01526106880152600081816102ba01526107b101526000818161054001526107d5015260008181610190015261051f0152600081816103e001526106c401526000818161027d01526104e20152610ca96000f3fe608060405234801561001057600080fd5b50600436106100935760003560e01c80634ded8d52116100665780634ded8d521461010c5780636edc1dcc1461012a57806374d7c13814610148578063bab8458d14610152578063dd5620b51461017057610093565b806313d9822314610098578063278ecde1146100b6578063379607f5146100d257806345bb8e09146100ee575b600080fd5b6100a061018e565b6040516100ad9190610810565b60405180910390f35b6100d060048036038101906100cb919061085c565b6101b2565b005b6100ec60048036038101906100e7919061085c565b610417565b005b6100f661065c565b6040516101039190610810565b60405180910390f35b610114610662565b6040516101219190610810565b60405180910390f35b610132610686565b60405161013f9190610810565b60405180910390f35b6101506106aa565b005b61015a6107af565b6040516101679190610810565b60405180910390f35b6101786107d3565b6040516101859190610810565b60405180910390f35b7f000000000000000000000000000000000000000000000000000000000000000081565b60011515600160009054906101000a900460ff161515141561021857600054421015610213576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161020a9061090c565b60405180910390fd5b61027b565b7f0000000000000000000000000000000000000000000000000000000000000000421061027a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161027190610978565b60405180910390fd5b5b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166366ce10b7827f00000000000000000000000000000000000000000000000000000000000000007f00000000000000000000000000000000000000000000000000000000000000006040518463ffffffff1660e01b815260040161031893929190610998565b60206040518083038186803b15801561033057600080fd5b505afa158015610344573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103689190610a07565b6103a7576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161039e90610aa6565b60405180910390fd5b7f3d2a04f53164bedf9a8a46353305d6b2d2261410406df3b41f99ce6489dc003c816040516103d69190610810565b60405180910390a17f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16ff5b60011515600160009054906101000a900460ff161515141561047c576000544210610477576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161046e90610b12565b60405180910390fd5b6104e0565b7f00000000000000000000000000000000000000000000000000000000000000004210156104df576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104d690610ba4565b60405180910390fd5b5b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166366ce10b7827f00000000000000000000000000000000000000000000000000000000000000007f00000000000000000000000000000000000000000000000000000000000000006040518463ffffffff1660e01b815260040161057d93929190610998565b60206040518083038186803b15801561059557600080fd5b505afa1580156105a9573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105cd9190610a07565b61060c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161060390610aa6565b60405180910390fd5b7f7a355715549cfe7c1cba26304350343fbddc4b4f72d3ce3e7c27117dd20b5cb88160405161063b9190610810565b60405180910390a13373ffffffffffffffffffffffffffffffffffffffff16ff5b60005481565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b600160009054906101000a900460ff1615801561071257507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b801561073d57507f000000000000000000000000000000000000000000000000000000000000000042105b61074657600080fd5b60018060006101000a81548160ff021916908315150217905550620151804261076f9190610bf3565b6000819055507f2724cf6c3ad6a3399ad72482e4013d0171794f3ef4c462b7e24790c658cb3cd460016040516107a59190610c58565b60405180910390a1565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b6000819050919050565b61080a816107f7565b82525050565b60006020820190506108256000830184610801565b92915050565b600080fd5b610839816107f7565b811461084457600080fd5b50565b60008135905061085681610830565b92915050565b6000602082840312156108725761087161082b565b5b600061088084828501610847565b91505092915050565b600082825260208201905092915050565b7f4974277320426f622773207475726e206e6f772c20706c65617365207761697460008201527f2100000000000000000000000000000000000000000000000000000000000000602082015250565b60006108f6602183610889565b91506109018261089a565b604082019050919050565b60006020820190508181036000830152610925816108e9565b9050919050565b7f4d697373656420796f7572206368616e63652100000000000000000000000000600082015250565b6000610962601383610889565b915061096d8261092c565b602082019050919050565b6000602082019050818103600083015261099181610955565b9050919050565b60006060820190506109ad6000830186610801565b6109ba6020830185610801565b6109c76040830184610801565b949350505050565b60008115159050919050565b6109e4816109cf565b81146109ef57600080fd5b50565b600081519050610a01816109db565b92915050565b600060208284031215610a1d57610a1c61082b565b5b6000610a2b848285016109f2565b91505092915050565b7f70726f76696465642073656372657420646f6573206e6f74206d61746368207460008201527f6865206578706563746564207075624b65790000000000000000000000000000602082015250565b6000610a90603283610889565b9150610a9b82610a34565b604082019050919050565b60006020820190508181036000830152610abf81610a83565b9050919050565b7f546f6f206c61746520746f20636c61696d210000000000000000000000000000600082015250565b6000610afc601283610889565b9150610b0782610ac6565b602082019050919050565b60006020820190508181036000830152610b2b81610aef565b9050919050565b7f2769735265616479203d3d2066616c7365272063616e6e6f7420636c61696d2060008201527f7965742100000000000000000000000000000000000000000000000000000000602082015250565b6000610b8e602483610889565b9150610b9982610b32565b604082019050919050565b60006020820190508181036000830152610bbd81610b81565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610bfe826107f7565b9150610c09836107f7565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03821115610c3e57610c3d610bc4565b5b828201905092915050565b610c52816109cf565b82525050565b6000602082019050610c6d6000830184610c49565b9291505056fea2646970667358221220d1390697bb816b8c0da3dc198430e5c03d3f6ca9e709974dbbd94d6ac68ecb7c64736f6c63430008090033608060405234801561001057600080fd5b506114fd806100206000396000f3fe608060405234801561001057600080fd5b50600436106101215760003560e01c80635f972df8116100ad578063913f424c11610071578063913f424c14610372578063bb8c256a146103a4578063db318833146103d5578063e241c1d914610407578063f47289e11461043857610121565b80635f972df81461027f57806366ce10b7146102b05780638081a1e7146102e05780638940aebe146103115780638cecf66e1461034257610121565b80631ecfe64d116100f45780631ecfe64d146101c35780632d58c9a2146101f45780632e52d606146102125780634df7e3d0146102305780635b7648111461024e57610121565b80630138e31b14610126578063022079d91461015757806306c91ce3146101875780630dbe671f146101a5575b600080fd5b610140600480360381019061013b9190610fbb565b61046a565b60405161014e929190611031565b60405180910390f35b610171600480360381019061016c919061105a565b610544565b60405161017e91906110f0565b60405180910390f35b61018f61065f565b60405161019c919061110b565b60405180910390f35b6101ad610683565b6040516101ba919061110b565b60405180910390f35b6101dd60048036038101906101d89190610fbb565b610688565b6040516101eb929190611031565b60405180910390f35b6101fc61078d565b604051610209919061110b565b60405180910390f35b61021a6107b1565b604051610227919061110b565b60405180910390f35b6102386107d5565b604051610245919061110b565b60405180910390f35b61026860048036038101906102639190610fbb565b6107da565b604051610276929190611031565b60405180910390f35b61029960048036038101906102949190610fbb565b610852565b6040516102a7929190611031565b60405180910390f35b6102ca60048036038101906102c59190611126565b6108ca565b6040516102d791906110f0565b60405180910390f35b6102fa60048036038101906102f59190611126565b610922565b604051610308929190611031565b60405180910390f35b61032b60048036038101906103269190611179565b6109bc565b604051610339929190611031565b60405180910390f35b61035c60048036038101906103579190611179565b610a13565b604051610369919061110b565b60405180910390f35b61038c60048036038101906103879190610fbb565b610b2c565b60405161039b939291906111a6565b60405180910390f35b6103be60048036038101906103b99190610fbb565b610be8565b6040516103cc929190611031565b60405180910390f35b6103ef60048036038101906103ea91906111dd565b610c86565b6040516103fe939291906111a6565b60405180910390f35b610421600480360381019061041c9190611126565b610ebe565b60405161042f929190611031565b60405180910390f35b610452600480360381019061044d9190611126565b610f58565b604051610461939291906111a6565b60405180910390f35b6000807ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f8061049c5761049b61126a565b5b7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f806104cb576104ca61126a565b5b8686097ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f806104fd576104fc61126a565b5b888609087ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f806105305761052f61126a565b5b848709809250819350505094509492505050565b6000807ffffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd036414190506000600160008060028a61057e9190611299565b141561058b57601b61058e565b601c5b8a60001b85806105a1576105a061126a565b5b8c8b0960001b604051600081526020016040526040516105c49493929190611347565b6020604051602081039080840390855afa1580156105e6573d6000803e3d6000fd5b505050602060405103519050600085856040516020016106079291906113ad565b6040516020818303038152906040528051906020012060001c90508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614935050505095945050505050565b7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f8179881565b600081565b6000807ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f806106ba576106b961126a565b5b7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f806106e9576106e861126a565b5b86867ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f6107169190611408565b097ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f806107465761074561126a565b5b888609087ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f806107795761077861126a565b5b848709809250819350505094509492505050565b7f483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b881565b7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f81565b600781565b6000807ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f8061080c5761080b61126a565b5b8487097ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f8061083e5761083d61126a565b5b848709809250819350505094509492505050565b6000807ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f806108845761088361126a565b5b8387097ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f806108b6576108b561126a565b5b858709809250819350505094509492505050565b60006109197f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f817987f483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8868686610544565b90509392505050565b60008060006109348487876001610b2c565b80935081945082955050505061094981610a13565b90507ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f8061097a5761097961126a565b5b81840992507ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f806109ae576109ad61126a565b5b818309915050935093915050565b600080610a0a7f79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f817987f483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b885610922565b91509150915091565b6000806000905060006001905060007ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f9050600085905060005b60008214610b1f578183610a61919061143c565b9050837ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f80610a9357610a9261126a565b5b7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f80610ac257610ac161126a565b5b8684097ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f610af09190611408565b87088095508196505050818282610b07919061146d565b84610b129190611408565b8093508194505050610a4d565b8495505050505050919050565b60008060008087905060008790506000879050600087905060008060006001905060008e1415610b6d57600080600199509950995050505050505050610bde565b5b60008714610bc75760006001881614610b9c57610b8f838383898989610c86565b8093508194508295505050505b600287610ba9919061143c565b9650610bb6868686610f58565b809650819750829850505050610b6e565b828282809a50819b50829c50505050505050505050505b9450945094915050565b6000806000610bfd8787600188886001610c86565b809350819450829550505050610c1281610a13565b90507ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f80610c4357610c4261126a565b5b81840992507ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f80610c7757610c7661126a565b5b81830991505094509492505050565b6000806000806000806000808d148015610ca0575060008c145b15610cb75789898996509650965050505050610eb2565b60008a148015610cc75750600089145b15610cde578c8c8c96509650965050505050610eb2565b898d148015610cec5750888c145b15610d4c57610cfd8d8c8f8e6107da565b8094508195505050610d138484600360016107da565b8094508195505050610d2984846000600161046a565b8094508195505050610d3f8c8c600260016107da565b8092508193505050610d75565b610d5889898e8e610688565b8094508195505050610d6c8a898f8e610688565b80925081935050505b610d8184848484610852565b8094508195505050610d95848486866107da565b8093508198505050610da987838f8e610688565b8093508198505050610dbd87838c8b610688565b8093508198505050610dd18d8c8985610688565b8092508197505050610de5868286866107da565b8092508197505050610df986828e8e610688565b8092508197505050808214610ea9577ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f80610e3757610e3661126a565b5b81880996507ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f80610e6b57610e6a61126a565b5b82870995507ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f80610e9f57610e9e61126a565b5b8183099450610ead565b8194505b505050505b96509650969350505050565b6000806000610ed08686866001610b2c565b809350819450829550505050610ee581610a13565b90507ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f80610f1657610f1561126a565b5b81840992507ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f80610f4a57610f4961126a565b5b818309915050935093915050565b6000806000610f6b868686898989610c86565b80935081945082955050505093509350939050565b600080fd5b6000819050919050565b610f9881610f85565b8114610fa357600080fd5b50565b600081359050610fb581610f8f565b92915050565b60008060008060808587031215610fd557610fd4610f80565b5b6000610fe387828801610fa6565b9450506020610ff487828801610fa6565b935050604061100587828801610fa6565b925050606061101687828801610fa6565b91505092959194509250565b61102b81610f85565b82525050565b60006040820190506110466000830185611022565b6110536020830184611022565b9392505050565b600080600080600060a0868803121561107657611075610f80565b5b600061108488828901610fa6565b955050602061109588828901610fa6565b94505060406110a688828901610fa6565b93505060606110b788828901610fa6565b92505060806110c888828901610fa6565b9150509295509295909350565b60008115159050919050565b6110ea816110d5565b82525050565b600060208201905061110560008301846110e1565b92915050565b60006020820190506111206000830184611022565b92915050565b60008060006060848603121561113f5761113e610f80565b5b600061114d86828701610fa6565b935050602061115e86828701610fa6565b925050604061116f86828701610fa6565b9150509250925092565b60006020828403121561118f5761118e610f80565b5b600061119d84828501610fa6565b91505092915050565b60006060820190506111bb6000830186611022565b6111c86020830185611022565b6111d56040830184611022565b949350505050565b60008060008060008060c087890312156111fa576111f9610f80565b5b600061120889828a01610fa6565b965050602061121989828a01610fa6565b955050604061122a89828a01610fa6565b945050606061123b89828a01610fa6565b935050608061124c89828a01610fa6565b92505060a061125d89828a01610fa6565b9150509295509295509295565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b60006112a482610f85565b91506112af83610f85565b9250826112bf576112be61126a565b5b828206905092915050565b6000819050919050565b6000819050919050565b60008160001b9050919050565b60006113066113016112fc846112ca565b6112de565b6112d4565b9050919050565b611316816112eb565b82525050565b600060ff82169050919050565b6113328161131c565b82525050565b611341816112d4565b82525050565b600060808201905061135c600083018761130d565b6113696020830186611329565b6113766040830185611338565b6113836060830184611338565b95945050505050565b6000819050919050565b6113a76113a282610f85565b61138c565b82525050565b60006113b98285611396565b6020820191506113c98284611396565b6020820191508190509392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061141382610f85565b915061141e83610f85565b925082821015611431576114306113d9565b5b828203905092915050565b600061144782610f85565b915061145283610f85565b9250826114625761146161126a565b5b828204905092915050565b600061147882610f85565b915061148383610f85565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff04831182151516156114bc576114bb6113d9565b5b82820290509291505056fea2646970667358221220f81983bfaa15e20a9df1c51d692c116ecca46d245b1ea48b7b50dbbfe30a0b4464736f6c63430008090033",
}

// SwapDLEQABI is the input ABI used to generate the binding from.
// Deprecated: Use SwapDLEQMetaData.ABI instead.
var SwapDLEQABI = SwapDLEQMetaData.ABI

// SwapDLEQBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SwapDLEQMetaData.Bin instead.
var SwapDLEQBin = SwapDLEQMetaData.Bin

// DeploySwapDLEQ deploys a new Ethereum contract, binding an instance of SwapDLEQ to it.
func DeploySwapDLEQ(auth *bind.TransactOpts, backend bind.ContractBackend, _pubKeyClaimX *big.Int, _pubKeyClaimY *big.Int, _pubKeyRefundX *big.Int, _pubKeyRefundY *big.Int) (common.Address, *types.Transaction, *SwapDLEQ, error) {
	parsed, err := SwapDLEQMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SwapDLEQBin), backend, _pubKeyClaimX, _pubKeyClaimY, _pubKeyRefundX, _pubKeyRefundY)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SwapDLEQ{SwapDLEQCaller: SwapDLEQCaller{contract: contract}, SwapDLEQTransactor: SwapDLEQTransactor{contract: contract}, SwapDLEQFilterer: SwapDLEQFilterer{contract: contract}}, nil
}

// SwapDLEQ is an auto generated Go binding around an Ethereum contract.
type SwapDLEQ struct {
	SwapDLEQCaller     // Read-only binding to the contract
	SwapDLEQTransactor // Write-only binding to the contract
	SwapDLEQFilterer   // Log filterer for contract events
}

// SwapDLEQCaller is an auto generated read-only Go binding around an Ethereum contract.
type SwapDLEQCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapDLEQTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SwapDLEQTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapDLEQFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SwapDLEQFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapDLEQSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SwapDLEQSession struct {
	Contract     *SwapDLEQ         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SwapDLEQCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SwapDLEQCallerSession struct {
	Contract *SwapDLEQCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// SwapDLEQTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SwapDLEQTransactorSession struct {
	Contract     *SwapDLEQTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// SwapDLEQRaw is an auto generated low-level Go binding around an Ethereum contract.
type SwapDLEQRaw struct {
	Contract *SwapDLEQ // Generic contract binding to access the raw methods on
}

// SwapDLEQCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SwapDLEQCallerRaw struct {
	Contract *SwapDLEQCaller // Generic read-only contract binding to access the raw methods on
}

// SwapDLEQTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SwapDLEQTransactorRaw struct {
	Contract *SwapDLEQTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSwapDLEQ creates a new instance of SwapDLEQ, bound to a specific deployed contract.
func NewSwapDLEQ(address common.Address, backend bind.ContractBackend) (*SwapDLEQ, error) {
	contract, err := bindSwapDLEQ(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SwapDLEQ{SwapDLEQCaller: SwapDLEQCaller{contract: contract}, SwapDLEQTransactor: SwapDLEQTransactor{contract: contract}, SwapDLEQFilterer: SwapDLEQFilterer{contract: contract}}, nil
}

// NewSwapDLEQCaller creates a new read-only instance of SwapDLEQ, bound to a specific deployed contract.
func NewSwapDLEQCaller(address common.Address, caller bind.ContractCaller) (*SwapDLEQCaller, error) {
	contract, err := bindSwapDLEQ(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SwapDLEQCaller{contract: contract}, nil
}

// NewSwapDLEQTransactor creates a new write-only instance of SwapDLEQ, bound to a specific deployed contract.
func NewSwapDLEQTransactor(address common.Address, transactor bind.ContractTransactor) (*SwapDLEQTransactor, error) {
	contract, err := bindSwapDLEQ(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SwapDLEQTransactor{contract: contract}, nil
}

// NewSwapDLEQFilterer creates a new log filterer instance of SwapDLEQ, bound to a specific deployed contract.
func NewSwapDLEQFilterer(address common.Address, filterer bind.ContractFilterer) (*SwapDLEQFilterer, error) {
	contract, err := bindSwapDLEQ(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SwapDLEQFilterer{contract: contract}, nil
}

// bindSwapDLEQ binds a generic wrapper to an already deployed contract.
func bindSwapDLEQ(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SwapDLEQABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SwapDLEQ *SwapDLEQRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SwapDLEQ.Contract.SwapDLEQCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SwapDLEQ *SwapDLEQRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.SwapDLEQTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SwapDLEQ *SwapDLEQRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.SwapDLEQTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SwapDLEQ *SwapDLEQCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SwapDLEQ.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SwapDLEQ *SwapDLEQTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SwapDLEQ *SwapDLEQTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.contract.Transact(opts, method, params...)
}

// PubKeyClaimX is a free data retrieval call binding the contract method 0x13d98223.
//
// Solidity: function pubKeyClaimX() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCaller) PubKeyClaimX(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwapDLEQ.contract.Call(opts, &out, "pubKeyClaimX")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PubKeyClaimX is a free data retrieval call binding the contract method 0x13d98223.
//
// Solidity: function pubKeyClaimX() view returns(uint256)
func (_SwapDLEQ *SwapDLEQSession) PubKeyClaimX() (*big.Int, error) {
	return _SwapDLEQ.Contract.PubKeyClaimX(&_SwapDLEQ.CallOpts)
}

// PubKeyClaimX is a free data retrieval call binding the contract method 0x13d98223.
//
// Solidity: function pubKeyClaimX() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCallerSession) PubKeyClaimX() (*big.Int, error) {
	return _SwapDLEQ.Contract.PubKeyClaimX(&_SwapDLEQ.CallOpts)
}

// PubKeyClaimY is a free data retrieval call binding the contract method 0xdd5620b5.
//
// Solidity: function pubKeyClaimY() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCaller) PubKeyClaimY(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwapDLEQ.contract.Call(opts, &out, "pubKeyClaimY")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PubKeyClaimY is a free data retrieval call binding the contract method 0xdd5620b5.
//
// Solidity: function pubKeyClaimY() view returns(uint256)
func (_SwapDLEQ *SwapDLEQSession) PubKeyClaimY() (*big.Int, error) {
	return _SwapDLEQ.Contract.PubKeyClaimY(&_SwapDLEQ.CallOpts)
}

// PubKeyClaimY is a free data retrieval call binding the contract method 0xdd5620b5.
//
// Solidity: function pubKeyClaimY() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCallerSession) PubKeyClaimY() (*big.Int, error) {
	return _SwapDLEQ.Contract.PubKeyClaimY(&_SwapDLEQ.CallOpts)
}

// PubKeyRefundX is a free data retrieval call binding the contract method 0xbab8458d.
//
// Solidity: function pubKeyRefundX() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCaller) PubKeyRefundX(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwapDLEQ.contract.Call(opts, &out, "pubKeyRefundX")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PubKeyRefundX is a free data retrieval call binding the contract method 0xbab8458d.
//
// Solidity: function pubKeyRefundX() view returns(uint256)
func (_SwapDLEQ *SwapDLEQSession) PubKeyRefundX() (*big.Int, error) {
	return _SwapDLEQ.Contract.PubKeyRefundX(&_SwapDLEQ.CallOpts)
}

// PubKeyRefundX is a free data retrieval call binding the contract method 0xbab8458d.
//
// Solidity: function pubKeyRefundX() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCallerSession) PubKeyRefundX() (*big.Int, error) {
	return _SwapDLEQ.Contract.PubKeyRefundX(&_SwapDLEQ.CallOpts)
}

// PubKeyRefundY is a free data retrieval call binding the contract method 0x6edc1dcc.
//
// Solidity: function pubKeyRefundY() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCaller) PubKeyRefundY(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwapDLEQ.contract.Call(opts, &out, "pubKeyRefundY")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PubKeyRefundY is a free data retrieval call binding the contract method 0x6edc1dcc.
//
// Solidity: function pubKeyRefundY() view returns(uint256)
func (_SwapDLEQ *SwapDLEQSession) PubKeyRefundY() (*big.Int, error) {
	return _SwapDLEQ.Contract.PubKeyRefundY(&_SwapDLEQ.CallOpts)
}

// PubKeyRefundY is a free data retrieval call binding the contract method 0x6edc1dcc.
//
// Solidity: function pubKeyRefundY() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCallerSession) PubKeyRefundY() (*big.Int, error) {
	return _SwapDLEQ.Contract.PubKeyRefundY(&_SwapDLEQ.CallOpts)
}

// Timeout0 is a free data retrieval call binding the contract method 0x4ded8d52.
//
// Solidity: function timeout_0() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCaller) Timeout0(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwapDLEQ.contract.Call(opts, &out, "timeout_0")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Timeout0 is a free data retrieval call binding the contract method 0x4ded8d52.
//
// Solidity: function timeout_0() view returns(uint256)
func (_SwapDLEQ *SwapDLEQSession) Timeout0() (*big.Int, error) {
	return _SwapDLEQ.Contract.Timeout0(&_SwapDLEQ.CallOpts)
}

// Timeout0 is a free data retrieval call binding the contract method 0x4ded8d52.
//
// Solidity: function timeout_0() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCallerSession) Timeout0() (*big.Int, error) {
	return _SwapDLEQ.Contract.Timeout0(&_SwapDLEQ.CallOpts)
}

// Timeout1 is a free data retrieval call binding the contract method 0x45bb8e09.
//
// Solidity: function timeout_1() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCaller) Timeout1(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwapDLEQ.contract.Call(opts, &out, "timeout_1")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Timeout1 is a free data retrieval call binding the contract method 0x45bb8e09.
//
// Solidity: function timeout_1() view returns(uint256)
func (_SwapDLEQ *SwapDLEQSession) Timeout1() (*big.Int, error) {
	return _SwapDLEQ.Contract.Timeout1(&_SwapDLEQ.CallOpts)
}

// Timeout1 is a free data retrieval call binding the contract method 0x45bb8e09.
//
// Solidity: function timeout_1() view returns(uint256)
func (_SwapDLEQ *SwapDLEQCallerSession) Timeout1() (*big.Int, error) {
	return _SwapDLEQ.Contract.Timeout1(&_SwapDLEQ.CallOpts)
}

// Claim is a paid mutator transaction binding the contract method 0x379607f5.
//
// Solidity: function claim(uint256 _s) returns()
func (_SwapDLEQ *SwapDLEQTransactor) Claim(opts *bind.TransactOpts, _s *big.Int) (*types.Transaction, error) {
	return _SwapDLEQ.contract.Transact(opts, "claim", _s)
}

// Claim is a paid mutator transaction binding the contract method 0x379607f5.
//
// Solidity: function claim(uint256 _s) returns()
func (_SwapDLEQ *SwapDLEQSession) Claim(_s *big.Int) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.Claim(&_SwapDLEQ.TransactOpts, _s)
}

// Claim is a paid mutator transaction binding the contract method 0x379607f5.
//
// Solidity: function claim(uint256 _s) returns()
func (_SwapDLEQ *SwapDLEQTransactorSession) Claim(_s *big.Int) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.Claim(&_SwapDLEQ.TransactOpts, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 _s) returns()
func (_SwapDLEQ *SwapDLEQTransactor) Refund(opts *bind.TransactOpts, _s *big.Int) (*types.Transaction, error) {
	return _SwapDLEQ.contract.Transact(opts, "refund", _s)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 _s) returns()
func (_SwapDLEQ *SwapDLEQSession) Refund(_s *big.Int) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.Refund(&_SwapDLEQ.TransactOpts, _s)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 _s) returns()
func (_SwapDLEQ *SwapDLEQTransactorSession) Refund(_s *big.Int) (*types.Transaction, error) {
	return _SwapDLEQ.Contract.Refund(&_SwapDLEQ.TransactOpts, _s)
}

// SetReady is a paid mutator transaction binding the contract method 0x74d7c138.
//
// Solidity: function set_ready() returns()
func (_SwapDLEQ *SwapDLEQTransactor) SetReady(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwapDLEQ.contract.Transact(opts, "set_ready")
}

// SetReady is a paid mutator transaction binding the contract method 0x74d7c138.
//
// Solidity: function set_ready() returns()
func (_SwapDLEQ *SwapDLEQSession) SetReady() (*types.Transaction, error) {
	return _SwapDLEQ.Contract.SetReady(&_SwapDLEQ.TransactOpts)
}

// SetReady is a paid mutator transaction binding the contract method 0x74d7c138.
//
// Solidity: function set_ready() returns()
func (_SwapDLEQ *SwapDLEQTransactorSession) SetReady() (*types.Transaction, error) {
	return _SwapDLEQ.Contract.SetReady(&_SwapDLEQ.TransactOpts)
}

// SwapDLEQClaimedIterator is returned from FilterClaimed and is used to iterate over the raw logs and unpacked data for Claimed events raised by the SwapDLEQ contract.
type SwapDLEQClaimedIterator struct {
	Event *SwapDLEQClaimed // Event containing the contract specifics and raw log

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
func (it *SwapDLEQClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapDLEQClaimed)
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
		it.Event = new(SwapDLEQClaimed)
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
func (it *SwapDLEQClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapDLEQClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapDLEQClaimed represents a Claimed event raised by the SwapDLEQ contract.
type SwapDLEQClaimed struct {
	S   *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterClaimed is a free log retrieval operation binding the contract event 0x7a355715549cfe7c1cba26304350343fbddc4b4f72d3ce3e7c27117dd20b5cb8.
//
// Solidity: event Claimed(uint256 s)
func (_SwapDLEQ *SwapDLEQFilterer) FilterClaimed(opts *bind.FilterOpts) (*SwapDLEQClaimedIterator, error) {

	logs, sub, err := _SwapDLEQ.contract.FilterLogs(opts, "Claimed")
	if err != nil {
		return nil, err
	}
	return &SwapDLEQClaimedIterator{contract: _SwapDLEQ.contract, event: "Claimed", logs: logs, sub: sub}, nil
}

// WatchClaimed is a free log subscription operation binding the contract event 0x7a355715549cfe7c1cba26304350343fbddc4b4f72d3ce3e7c27117dd20b5cb8.
//
// Solidity: event Claimed(uint256 s)
func (_SwapDLEQ *SwapDLEQFilterer) WatchClaimed(opts *bind.WatchOpts, sink chan<- *SwapDLEQClaimed) (event.Subscription, error) {

	logs, sub, err := _SwapDLEQ.contract.WatchLogs(opts, "Claimed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapDLEQClaimed)
				if err := _SwapDLEQ.contract.UnpackLog(event, "Claimed", log); err != nil {
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
func (_SwapDLEQ *SwapDLEQFilterer) ParseClaimed(log types.Log) (*SwapDLEQClaimed, error) {
	event := new(SwapDLEQClaimed)
	if err := _SwapDLEQ.contract.UnpackLog(event, "Claimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapDLEQConstructedIterator is returned from FilterConstructed and is used to iterate over the raw logs and unpacked data for Constructed events raised by the SwapDLEQ contract.
type SwapDLEQConstructedIterator struct {
	Event *SwapDLEQConstructed // Event containing the contract specifics and raw log

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
func (it *SwapDLEQConstructedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapDLEQConstructed)
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
		it.Event = new(SwapDLEQConstructed)
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
func (it *SwapDLEQConstructedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapDLEQConstructedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapDLEQConstructed represents a Constructed event raised by the SwapDLEQ contract.
type SwapDLEQConstructed struct {
	P   *big.Int
	Q   *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterConstructed is a free log retrieval operation binding the contract event 0xbc91786c56ba12548733c92d8dc8bc6e725d7ac2eb0c3039fddbcfbf19a852a9.
//
// Solidity: event Constructed(uint256 p, uint256 q)
func (_SwapDLEQ *SwapDLEQFilterer) FilterConstructed(opts *bind.FilterOpts) (*SwapDLEQConstructedIterator, error) {

	logs, sub, err := _SwapDLEQ.contract.FilterLogs(opts, "Constructed")
	if err != nil {
		return nil, err
	}
	return &SwapDLEQConstructedIterator{contract: _SwapDLEQ.contract, event: "Constructed", logs: logs, sub: sub}, nil
}

// WatchConstructed is a free log subscription operation binding the contract event 0xbc91786c56ba12548733c92d8dc8bc6e725d7ac2eb0c3039fddbcfbf19a852a9.
//
// Solidity: event Constructed(uint256 p, uint256 q)
func (_SwapDLEQ *SwapDLEQFilterer) WatchConstructed(opts *bind.WatchOpts, sink chan<- *SwapDLEQConstructed) (event.Subscription, error) {

	logs, sub, err := _SwapDLEQ.contract.WatchLogs(opts, "Constructed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapDLEQConstructed)
				if err := _SwapDLEQ.contract.UnpackLog(event, "Constructed", log); err != nil {
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

// ParseConstructed is a log parse operation binding the contract event 0xbc91786c56ba12548733c92d8dc8bc6e725d7ac2eb0c3039fddbcfbf19a852a9.
//
// Solidity: event Constructed(uint256 p, uint256 q)
func (_SwapDLEQ *SwapDLEQFilterer) ParseConstructed(log types.Log) (*SwapDLEQConstructed, error) {
	event := new(SwapDLEQConstructed)
	if err := _SwapDLEQ.contract.UnpackLog(event, "Constructed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapDLEQIsReadyIterator is returned from FilterIsReady and is used to iterate over the raw logs and unpacked data for IsReady events raised by the SwapDLEQ contract.
type SwapDLEQIsReadyIterator struct {
	Event *SwapDLEQIsReady // Event containing the contract specifics and raw log

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
func (it *SwapDLEQIsReadyIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapDLEQIsReady)
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
		it.Event = new(SwapDLEQIsReady)
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
func (it *SwapDLEQIsReadyIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapDLEQIsReadyIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapDLEQIsReady represents a IsReady event raised by the SwapDLEQ contract.
type SwapDLEQIsReady struct {
	B   bool
	Raw types.Log // Blockchain specific contextual infos
}

// FilterIsReady is a free log retrieval operation binding the contract event 0x2724cf6c3ad6a3399ad72482e4013d0171794f3ef4c462b7e24790c658cb3cd4.
//
// Solidity: event IsReady(bool b)
func (_SwapDLEQ *SwapDLEQFilterer) FilterIsReady(opts *bind.FilterOpts) (*SwapDLEQIsReadyIterator, error) {

	logs, sub, err := _SwapDLEQ.contract.FilterLogs(opts, "IsReady")
	if err != nil {
		return nil, err
	}
	return &SwapDLEQIsReadyIterator{contract: _SwapDLEQ.contract, event: "IsReady", logs: logs, sub: sub}, nil
}

// WatchIsReady is a free log subscription operation binding the contract event 0x2724cf6c3ad6a3399ad72482e4013d0171794f3ef4c462b7e24790c658cb3cd4.
//
// Solidity: event IsReady(bool b)
func (_SwapDLEQ *SwapDLEQFilterer) WatchIsReady(opts *bind.WatchOpts, sink chan<- *SwapDLEQIsReady) (event.Subscription, error) {

	logs, sub, err := _SwapDLEQ.contract.WatchLogs(opts, "IsReady")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapDLEQIsReady)
				if err := _SwapDLEQ.contract.UnpackLog(event, "IsReady", log); err != nil {
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
func (_SwapDLEQ *SwapDLEQFilterer) ParseIsReady(log types.Log) (*SwapDLEQIsReady, error) {
	event := new(SwapDLEQIsReady)
	if err := _SwapDLEQ.contract.UnpackLog(event, "IsReady", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwapDLEQRefundedIterator is returned from FilterRefunded and is used to iterate over the raw logs and unpacked data for Refunded events raised by the SwapDLEQ contract.
type SwapDLEQRefundedIterator struct {
	Event *SwapDLEQRefunded // Event containing the contract specifics and raw log

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
func (it *SwapDLEQRefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapDLEQRefunded)
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
		it.Event = new(SwapDLEQRefunded)
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
func (it *SwapDLEQRefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapDLEQRefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapDLEQRefunded represents a Refunded event raised by the SwapDLEQ contract.
type SwapDLEQRefunded struct {
	S   *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterRefunded is a free log retrieval operation binding the contract event 0x3d2a04f53164bedf9a8a46353305d6b2d2261410406df3b41f99ce6489dc003c.
//
// Solidity: event Refunded(uint256 s)
func (_SwapDLEQ *SwapDLEQFilterer) FilterRefunded(opts *bind.FilterOpts) (*SwapDLEQRefundedIterator, error) {

	logs, sub, err := _SwapDLEQ.contract.FilterLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return &SwapDLEQRefundedIterator{contract: _SwapDLEQ.contract, event: "Refunded", logs: logs, sub: sub}, nil
}

// WatchRefunded is a free log subscription operation binding the contract event 0x3d2a04f53164bedf9a8a46353305d6b2d2261410406df3b41f99ce6489dc003c.
//
// Solidity: event Refunded(uint256 s)
func (_SwapDLEQ *SwapDLEQFilterer) WatchRefunded(opts *bind.WatchOpts, sink chan<- *SwapDLEQRefunded) (event.Subscription, error) {

	logs, sub, err := _SwapDLEQ.contract.WatchLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapDLEQRefunded)
				if err := _SwapDLEQ.contract.UnpackLog(event, "Refunded", log); err != nil {
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
func (_SwapDLEQ *SwapDLEQFilterer) ParseRefunded(log types.Log) (*SwapDLEQRefunded, error) {
	event := new(SwapDLEQRefunded)
	if err := _SwapDLEQ.contract.UnpackLog(event, "Refunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
