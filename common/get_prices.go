package common

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// MainnetEndpoint is a mainnet Ethereum endpoint.
// from https://chainlist.org/chain/1
const MainnetEndpoint = "https://eth-rpc.gateway.pokt.network"

var (
	functionSig = ethcrypto.Keccak256([]byte("latestRoundData()"))[:4]

	errUnsupportedNetwork = errors.New("unsupported network")
)

// see latestRoundData at https://docs.chain.link/data-feeds/price-feeds/api-reference/
func latestRoundDataReturnArgs() abi.Arguments {
	uint256Ty, err := abi.NewType("uint256", "", nil)
	if err != nil {
		panic(fmt.Errorf("failed to create uint256 type: %w", err))
	}

	int256Ty, err := abi.NewType("int256", "", nil)
	if err != nil {
		panic(fmt.Errorf("failed to create int256 type: %w", err))
	}

	uint80Ty, err := abi.NewType("uint80", "", nil)
	if err != nil {
		panic(fmt.Errorf("failed to create uint80 type: %w", err))
	}

	return abi.Arguments{
		{
			Type: uint80Ty, // roundId
		},
		{
			Type: int256Ty, // answer (the price)
		},
		{
			Type: uint256Ty, // startedAt
		},
		{
			Type: uint256Ty, // updatedAt
		},
		{
			Type: uint80Ty, // answeredInRound
		},
	}
}

// GetETHUSDPrice returns the current ETH/USD price from the Chainlink oracle.
// It errors if the chain ID is not the Ethereum mainnet.
func GetETHUSDPrice(ctx context.Context, ec *ethclient.Client) (*big.Int, error) {
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	var to ethcommon.Address

	switch chainID.Uint64() {
	case MainnetChainID:
		// see https://data.chain.link/ethereum/mainnet/crypto-usd/eth-usd
		to = ethcommon.HexToAddress("0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419")
	case GoerliChainID, GanacheChainID, HardhatChainID:
		// for goerli or test clients, just return 1
		return big.NewInt(1000), nil
	default:
		return nil, errUnsupportedNetwork
	}

	return callLatestRoundData(ctx, ec, to)
}

// GetXMRUSDPrice returns the current XMR/USD price from the Chainlink oracle.
// It errors if the chain ID is not the Ethereum mainnet.
func GetXMRUSDPrice(ctx context.Context, ec *ethclient.Client) (*big.Int, error) {
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	var to ethcommon.Address

	switch chainID.Uint64() {
	case MainnetChainID:
		// see https://data.chain.link/ethereum/mainnet/crypto-usd/xmr-usd
		to = ethcommon.HexToAddress("0xfa66458cce7dd15d8650015c4fce4d278271618f")
	case GoerliChainID, GanacheChainID, HardhatChainID:
		// for goerli or test clients, just return 1
		return big.NewInt(100), nil
	default:
		return nil, errUnsupportedNetwork
	}

	return callLatestRoundData(ctx, ec, to)
}

func callLatestRoundData(ctx context.Context, ec *ethclient.Client, to ethcommon.Address) (*big.Int, error) {
	msg := ethereum.CallMsg{
		To:   &to,
		Data: functionSig,
	}

	ret, err := ec.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}

	arguments := latestRoundDataReturnArgs()
	args, err := arguments.Unpack(
		ret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack arguments: %w", err)
	}

	return args[1].(*big.Int), nil
}
