package xmrmaker

import (
	"context"
	"math/big"
	"testing"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/swapfactory"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func TestCheckContractCode(t *testing.T) {
	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	pk, err := ethcrypto.HexToECDSA(common.DefaultPrivKeyXMRMaker)
	require.NoError(t, err)

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(common.GanacheChainID))
	require.NoError(t, err)

	addr, _, _, err := swapfactory.DeploySwapFactory(txOpts, ec)
	require.NoError(t, err)

	err = checkContractCode(context.Background(), ec, addr)
	require.NoError(t, err)
}
