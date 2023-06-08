package txsender

import (
	"context"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/coins"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/tests"
)

func init() {
	cliutil.SetLogLevels("debug")
}

func newTestSender(t *testing.T) (*privateKeySender, *coins.ERC20TokenInfo) {
	ctx := context.Background()
	pk := tests.GetTakerTestKey(t)
	ec := extethclient.CreateTestClient(t, pk)

	swapCreatorAddr, swapCreator := contracts.DevDeploySwapCreator(t, ec.Raw(), pk)

	token := contracts.GetMockTether(t, ec.Raw(), pk)
	tokenBinding, err := contracts.NewIERC20(token.Address, ec.Raw())
	require.NoError(t, err)

	sender := NewSenderWithPrivateKey(ctx, ec, swapCreatorAddr, swapCreator, tokenBinding)
	return sender.(*privateKeySender), token
}

// Verify that our test token behaves like Tether in that changing the approval amount
// must be set to zero before it can be changed to a non-zero value.
func Test_privateKeySender_approve(t *testing.T) {
	sender, token := newTestSender(t)

	zeroAmt := coins.NewTokenAmountFromDecimals(new(apd.Decimal), token)
	tooMuchAmt := coins.NewTokenAmountFromDecimals(coins.StrToDecimal("1000000"), token)

	// Ensure this fails, approving zero amounts should be done using approveNoChecks
	err := sender.approveTransferFrom(zeroAmt)
	require.ErrorContains(t, err, "can not be called with a zero amount")

	// Ensure our balance check failure works
	err = sender.approveTransferFrom(tooMuchAmt)
	require.ErrorContains(t, err, " is under ")

	// Make sure that we always start testing in a know state, where the swapCreator
	// contract is not approved to transfer any token amount
	err = sender.approveNoChecks(zeroAmt)
	require.NoError(t, err)

	// First approve succeeds, as the current allowance is zero
	amt := coins.NewTokenAmountFromDecimals(coins.StrToDecimal("3"), token)
	err = sender.approveTransferFrom(amt)
	require.NoError(t, err)

	// Second approve succeeds, as we are already approved for more than is being
	// asked for. This is an attempt to optimize gas, but it should happen very
	// often. The only way you can easily get into this state is if a previous
	// swap failed after approve but before a successful NewSwap transaction.
	amt = coins.NewTokenAmountFromDecimals(coins.StrToDecimal("2"), token)
	err = sender.approveTransferFrom(amt)
	require.NoError(t, err)

	// Third approve fails, because we are trying to increase the approval
	// amount without zeroing it first. Contracts like Tether do not allow this,
	// as it gives the contract being granted approval the potential to quickly
	// transfer the previously approved amount before the new approval gets
	// mined and then do a second transfer using the new approved amount.
	amt = coins.NewTokenAmountFromDecimals(coins.StrToDecimal("4"), token)
	err = sender.approveNoChecks(amt)
	require.ErrorContains(t, err, `token approve tx for 4 "USDT" creation failed`)

	// The next approval, of the same amount that just failed, is actually 2
	// approvals. When calling approveTransferFrom, the code will see that the
	// approval amount needs to be raised and zero it out before raising it.
	err = sender.approveTransferFrom(amt)
	require.NoError(t, err)
}
