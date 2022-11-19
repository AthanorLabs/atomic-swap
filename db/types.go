package db

import (
	"math/big"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// EthereumSwapInfo represents information required on the Ethereum side in case of recovery
type EthereumSwapInfo struct {
	StartNumber     *big.Int                  `json:"start_number"`
	SwapID          types.Hash                `json:"swap_id"`
	Swap            contracts.SwapFactorySwap `json:"swap"`
	ContractAddress ethcommon.Address         `json:"contract_address"`
}
