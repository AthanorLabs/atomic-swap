package db

import (
	"math/big"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// EthereumSwapInfo represents information required on the Ethereum side in case of recovery
type EthereumSwapInfo struct {
	StartNumber     *big.Int                  `json:"startNumber"`
	SwapID          types.Hash                `json:"swapID"`
	Swap            contracts.SwapFactorySwap `json:"swap"`
	ContractAddress ethcommon.Address         `json:"contractAddress"`
}
