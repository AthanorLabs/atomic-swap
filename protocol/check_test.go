package protocol

import (
	"context"
	"testing"

	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestCheckContractCode(t *testing.T) {
	ec, chainID := tests.NewEthClient(t)
	ctx := context.Background()
	pk := tests.GetMakerTestKey(t)

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)

	_, tx, _, err := contracts.DeploySwapFactory(txOpts, ec, ethcommon.Address{})
	require.NoError(t, err)

	addr, err := bind.WaitDeployed(ctx, ec, tx)
	require.NoError(t, err)

	err = CheckContractCode(ctx, ec, addr)
	require.NoError(t, err)

	// deploy with some arbitrary trustedForwarder address
	_, tx, _, err = contracts.DeploySwapFactory(
		txOpts,
		ec,
		ethcommon.HexToAddress("0x64e902cD8A29bBAefb9D4e2e3A24d8250C606ee7"),
	)
	require.NoError(t, err)

	addr, err = bind.WaitDeployed(ctx, ec, tx)
	require.NoError(t, err)

	err = CheckContractCode(ctx, ec, addr)
	require.NoError(t, err)
}
