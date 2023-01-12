package pricefeed

import (
	"context"
	"errors"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
)

const (
	// mainnetEndpoint is a mainnet ethereum endpoint, from
	// https://chainlist.org/chain/1, which stagenet users get pointed at for price
	// feeds, as Goeri doesn't have an XMR feed. Mainnet users will use the same
	// ethereum endpoint that they use for other swap transactions.
	mainnetEndpoint = "https://eth-rpc.gateway.pokt.network"

	// https://data.chain.link/ethereum/mainnet/crypto-usd/eth-usd
	chainlinkETHToUSDProxy = "0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419"

	// https://data.chain.link/ethereum/mainnet/crypto-usd/xmr-usd
	chainlinkXMRToUSDProxy = "0xfa66458cce7dd15d8650015c4fce4d278271618f"
)

var (
	errUnsupportedNetwork = errors.New("unsupported network")
	log                   = logging.Logger("pricefeed")
)

// PriceFeed contains the interesting data from a chainlink price feed query.
type PriceFeed struct {
	Description string // "COIN / USD"
	Price       *apd.Decimal
	UpdatedAt   time.Time
}

// GetETHUSDPrice returns the current ETH/USD price from the Chainlink oracle.
// It errors if the chain ID is not the Ethereum mainnet.
func GetETHUSDPrice(ctx context.Context, ec *ethclient.Client) (*PriceFeed, error) {
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	switch chainID.Uint64() {
	case common.MainnetChainID:
		// No extra work to do
	case common.GoerliChainID:
		// Push stagenet/goerli users to a mainnet endpoint
		ec, err = ethclient.Dial(mainnetEndpoint)
		if err != nil {
			return nil, err
		}
		defer ec.Close()
	case common.GanacheChainID, common.HardhatChainID:
		return &PriceFeed{
			Description: "ETH / USD (fake)",
			Price:       apd.New(123412345678, -8),
			UpdatedAt:   time.Now(),
		}, nil
	default:
		return nil, errUnsupportedNetwork
	}

	return getChainlinkPriceFeed(ctx, chainlinkETHToUSDProxy, ec)
}

// GetXMRUSDPrice returns the current XMR/USD price from the Chainlink oracle.
// It errors if the chain ID is not the Ethereum mainnet.
func GetXMRUSDPrice(ctx context.Context, ec *ethclient.Client) (*PriceFeed, error) {
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	switch chainID.Uint64() {
	case common.MainnetChainID:
		// No extra work to do
	case common.GoerliChainID:
		// Push stagenet/goerli users to a mainnet endpoint
		ec, err = ethclient.Dial(mainnetEndpoint)
		if err != nil {
			return nil, err
		}
		defer ec.Close()
	case common.GanacheChainID, common.HardhatChainID:
		return &PriceFeed{
			Description: "XMR / USD (fake)",
			Price:       apd.New(12312345678, -8), // 123.12345678
			UpdatedAt:   time.Now(),
		}, nil
	default:
		return nil, errUnsupportedNetwork
	}

	return getChainlinkPriceFeed(ctx, chainlinkXMRToUSDProxy, ec)
}

// getChainlinkPriceFeed retries the latest price feed data from the given contract address.
func getChainlinkPriceFeed(ctx context.Context, feedAddress string, ec *ethclient.Client) (*PriceFeed, error) {

	chainlinkPriceFeedProxy, err := contracts.NewAggregatorV3Interface(ethcommon.HexToAddress(feedAddress), ec)
	if err != nil {
		return nil, err
	}

	opts := &bind.CallOpts{
		Context: ctx,
	}

	roundData, err := chainlinkPriceFeedProxy.LatestRoundData(opts)
	if err != nil {
		return nil, err
	}

	decimals, err := chainlinkPriceFeedProxy.Decimals(opts)
	if err != nil {
		return nil, err
	}

	price := apd.NewWithBigInt(new(apd.BigInt).SetMathBigInt(roundData.Answer), -int32(decimals))
	updatedAt := time.Unix(roundData.StartedAt.Int64(), 0)

	description, err := chainlinkPriceFeedProxy.Description(opts)
	if err != nil {
		return nil, err
	}

	log.Debugf("%s: $%s (%s)", description, price, updatedAt)
	return &PriceFeed{
		Description: description,
		Price:       price,
		UpdatedAt:   updatedAt,
	}, nil
}
