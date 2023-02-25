package db

import (
	"math/big"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// EthereumSwapInfo represents information required on the Ethereum side in case of recovery
type EthereumSwapInfo struct {
	// StartNumber the block number of the `newSwap` transaction.
	// The same for both maker/taker.
	StartNumber *big.Int `json:"startNumber"`
	// SwapID is the swap ID used by the swap contract; not the same as the swap/offer ID
	// used by swapd.
	// It's the hash of the ABI encoded `contracts.SwapFactorySwap` struct.
	SwapID types.Hash `json:"swapID"`
	// Swap is the `Swap` structure inside SwapFactory.sol.
	Swap            *contracts.SwapFactorySwap `json:"swap"`
	ContractAddress ethcommon.Address          `json:"contractAddress"`
}
