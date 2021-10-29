package bob

import (
	"context"
	"math/big"
	"testing"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/swap-contract"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/stretchr/testify/require"
)

func newTestBob(t *testing.T) *bob {
	bob, err := NewBob(context.Background(), common.DefaultBobMoneroEndpoint, common.DefaultDaemonEndpoint, common.DefaultEthEndpoint, common.DefaultPrivKeyBob)
	require.NoError(t, err)
	_, _, err = bob.GenerateKeys()
	require.NoError(t, err)

	conn, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	pkBob, err := crypto.HexToECDSA(common.DefaultPrivKeyBob)
	require.NoError(t, err)

	bob.auth, err = bind.NewKeyedTransactorWithChainID(pkBob, big.NewInt(common.GanacheChainID))
	require.NoError(t, err)

	var pubkey [32]byte
	copy(pubkey[:], bob.pubkeys.SpendKey().Bytes())
	bob.contractAddr, _, bob.contract, err = swap.DeploySwap(bob.auth, conn, pubkey, [32]byte{})
	require.NoError(t, err)
	return bob
}

func TestBob_GenerateKeys(t *testing.T) {
	bob := newTestBob(t)

	pubSpendKey, privViewKey, err := bob.GenerateKeys()
	require.NoError(t, err)
	require.NotNil(t, bob.privkeys)
	require.NotNil(t, bob.pubkeys)
	require.NotNil(t, pubSpendKey)
	require.NotNil(t, privViewKey)
}

func TestBob_ClaimFunds(t *testing.T) {
	bob := newTestBob(t)

	_, err := bob.contract.(*swap.Swap).SetReady(bob.auth)
	require.NoError(t, err)

	txHash, err := bob.ClaimFunds()
	require.NoError(t, err)
	require.NotEqual(t, "", txHash)
}
