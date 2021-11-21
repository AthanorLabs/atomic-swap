package bob

import (
	"context"
	"math/big"
	"testing"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/swap-contract"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

var defaultTimeoutDuration = big.NewInt(60 * 60 * 24) // 1 day = 60s * 60min * 24hr

func newTestBob(t *testing.T) (*bob, *swapState) {
	bob, err := NewBob(context.Background(), common.DefaultBobMoneroEndpoint, common.DefaultMoneroDaemonEndpoint, common.DefaultEthEndpoint, common.DefaultPrivKeyBob)
	require.NoError(t, err)
	swapState := newSwapState(bob, 1, 1)
	return bob, swapState
}

func TestSwapState_GenerateKeys(t *testing.T) {
	_, swapState := newTestBob(t)

	pubSpendKey, privViewKey, err := swapState.generateKeys()
	require.NoError(t, err)
	require.NotNil(t, swapState.privkeys)
	require.NotNil(t, swapState.pubkeys)
	require.NotNil(t, pubSpendKey)
	require.NotNil(t, privViewKey)
}

func TestSwapState_ClaimFunds(t *testing.T) {
	bob, swapState := newTestBob(t)
	_, _, err := swapState.generateKeys()
	require.NoError(t, err)

	conn, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	pkBob, err := crypto.HexToECDSA(common.DefaultPrivKeyBob)
	require.NoError(t, err)

	bob.auth, err = bind.NewKeyedTransactorWithChainID(pkBob, big.NewInt(common.GanacheChainID))
	require.NoError(t, err)

	var pubkey [32]byte
	copy(pubkey[:], swapState.pubkeys.SpendKey().Bytes())
	swapState.contractAddr, _, swapState.contract, err = swap.DeploySwap(bob.auth, conn, pubkey, [32]byte{}, defaultTimeoutDuration)
	require.NoError(t, err)

	_, err = swapState.contract.SetReady(bob.auth)
	require.NoError(t, err)

	txHash, err := swapState.claimFunds()
	require.NoError(t, err)
	require.NotEqual(t, "", txHash)
}

func TestSwapState_HandleProtocolMessage_SendKeysMessage(t *testing.T) {
	_, s := newTestBob(t)

	msg := &net.SendKeysMessage{}
	err := s.handleSendKeysMessage(msg)
	require.Equal(t, errMissingKeys, err)

	alicePrivKeys, err := monero.GenerateKeys()
	require.NoError(t, err)
	alicePubKeys := alicePrivKeys.PublicKeyPair()

	msg = &net.SendKeysMessage{
		PublicSpendKey: alicePrivKeys.SpendKey().Public().Hex(),
		PublicViewKey:  alicePrivKeys.ViewKey().Public().Hex(),
	}

	err = s.handleSendKeysMessage(msg)
	require.NoError(t, err)
	require.Equal(t, &net.NotifyContractDeployed{}, s.nextExpectedMessage)
	require.Equal(t, alicePubKeys.SpendKey().Hex(), s.alicePublicKeys.SpendKey().Hex())
	require.Equal(t, alicePubKeys.ViewKey().Hex(), s.alicePublicKeys.ViewKey().Hex())
}

func TestSwapState_HandleProtocolMessage_NotifyContractDeployed(t *testing.T) {

}

func TestSwapState_HandleProtocolMessage_NotifyReady(t *testing.T) {

}

func TestSwapState_HandleProtocolMessage_NotifyRefund(t *testing.T) {

}
