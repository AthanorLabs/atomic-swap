package tests

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/rpcclient/wsclient"
)

func setupXMRTakerAuth(t *testing.T) (*bind.TransactOpts, *ethclient.Client, *ecdsa.PrivateKey) {
	conn, chainID := NewEthClient(t)
	pk, err := ethcrypto.HexToECDSA(common.DefaultPrivKeyXMRTaker)
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)
	return auth, conn, pk
}

// deploys ERC20Mock.sol and assigns the whole token balance to the XMRTaker default address.
func deployERC20Mock(t *testing.T) ethcommon.Address {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := ethcrypto.PubkeyToAddress(*pub)

	decimals := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)
	balance := big.NewInt(0).Mul(big.NewInt(999), decimals)
	erc20Addr, erc20Tx, _, err := contracts.DeployERC20Mock(auth, conn, "ERC20Mock", "MOCK", addr, balance)
	require.NoError(t, err)
	_, err = block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	return erc20Addr
}

func TestSuccess_ERC20_OneSwap(t *testing.T) {
	if os.Getenv(generateBlocksEnv) != falseStr {
		generateBlocks(64)
	}

	erc20Addr := deployERC20Mock(t)

	const testTimeout = time.Second * 75

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bwsc, err := wsclient.NewWsClient(ctx, defaultXMRMakerDaemonWSEndpoint)
	require.NoError(t, err)

	offerID, statusCh, err := bwsc.MakeOfferAndSubscribe(0.1, xmrmakerProvideAmount,
		types.ExchangeRate(exchangeRate), types.EthAsset(erc20Addr))
	require.NoError(t, err)

	bc := rpcclient.NewClient(defaultXMRMakerDaemonEndpoint)
	offersBefore, err := bc.GetOffers()
	require.NoError(t, err)
	defer func() {
		err = bc.ClearOffers(nil)
		require.NoError(t, err)
	}()

	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			select {
			case status := <-statusCh:
				t.Log("> XMRMaker got status:", status)
				if status.IsOngoing() {
					continue
				}

				if status != types.CompletedSuccess {
					errCh <- fmt.Errorf("swap did not complete successfully: got %s", status)
				}
				return
			case <-time.After(testTimeout):
				errCh <- errors.New("make offer subscription timed out")
				return
			}
		}
	}()

	ac := rpcclient.NewClient(defaultXMRTakerDaemonEndpoint)
	awsc, err := wsclient.NewWsClient(ctx, defaultXMRTakerDaemonWSEndpoint)
	require.NoError(t, err)

	// TODO: implement discovery over websockets (#97)
	providers, err := ac.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)

	takerStatusCh, err := awsc.TakeOfferAndSubscribe(providers[0][0], offerID, 0.05)
	require.NoError(t, err)

	go func() {
		defer wg.Done()
		for status := range takerStatusCh {
			t.Log("> XMRTaker got status:", status)
			if status.IsOngoing() {
				continue
			}
			if status != types.CompletedSuccess {
				errCh <- fmt.Errorf("swap did not complete successfully: got %s", status)
			}
			return
		}
	}()

	wg.Wait()

	select {
	case err = <-errCh:
		require.NoError(t, err)
	default:
	}

	offersAfter, err := bc.GetOffers()
	require.NoError(t, err)
	require.Equal(t, 1, len(offersBefore)-len(offersAfter))
}
