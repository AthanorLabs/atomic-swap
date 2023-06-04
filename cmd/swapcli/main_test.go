package main

import (
	"context"
	"fmt"
	"time"

	"github.com/MarinX/monerorpc/wallet"
	"github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/monero"
)

func (s *swapCLITestSuite) Test_runGetVersions() {
	// get the version of swapcli in isolation
	args := []string{"swapcli", "--version"}
	err := cliApp().RunContext(context.Background(), args)
	require.NoError(s.T(), err)

	// get both the swapcli version and the daemon version information
	args = []string{"swapcli", "version"}
	err = cliApp().RunContext(context.Background(), args)
	require.NoError(s.T(), err)
}

func (s *swapCLITestSuite) Test_runBalances() {
	args := []string{
		"swapcli",
		"balances",
		fmt.Sprintf("--%s=%s", flagToken, s.mockDaiAddr().Hex()),
	}
	err := cliApp().RunContext(context.Background(), args)
	require.NoError(s.T(), err)
}

func (s *swapCLITestSuite) Test_runRunETHTransfer() {
	bobAddr := s.bobConf.EthereumClient.Address().String()
	aliceAddr := s.aliceConf.EthereumClient.Address().String()
	amount := coins.EtherToWei(coins.StrToDecimal("0.123"))
	bobEC := s.bobConf.EthereumClient

	s.T().Logf("Alice is transferring %s ETH to Bob", amount.AsEtherString())
	args := []string{
		"swapcli",
		"transfer-eth",
		s.aliceSwapdPortFlag(),
		fmt.Sprintf("--%s=%s", flagTo, bobAddr),
		fmt.Sprintf("--%s=%s", flagAmount, amount.AsEtherString()),
	}
	err := cliApp().RunContext(context.Background(), args)
	require.NoError(s.T(), err)

	bobBal, err := bobEC.Balance(context.Background())
	require.NoError(s.T(), err)
	require.GreaterOrEqual(s.T(), bobBal.Cmp(amount), 0)

	s.T().Log("Bob is sweeping his entire ETH balance back to Alice")
	args = []string{
		"swapcli",
		"sweep-eth",
		fmt.Sprintf("--%s=%d", flagSwapdPort, s.bobRPCPort()),
		fmt.Sprintf("--%s=%s", flagTo, aliceAddr),
	}
	err = cliApp().RunContext(context.Background(), args)
	require.NoError(s.T(), err)

	bobBal, err = bobEC.Balance(context.Background())
	require.NoError(s.T(), err)
	require.True(s.T(), bobBal.Decimal().IsZero())
}

func (s *swapCLITestSuite) Test_runRunETHTransfer_toContract() {
	ctx := context.Background()
	ec := s.aliceConf.EthereumClient
	zeroETH := new(apd.Decimal)
	token := contracts.GetMockTether(s.T(), ec.Raw(), ec.PrivateKey())

	startTokenBal, err := ec.ERC20Balance(ctx, token.Address)
	require.NoError(s.T(), err)

	// transfer zero ETH to the token address. We'll fail the 1st attempt by not
	// setting the gas limit.
	args := []string{
		"swapcli",
		"transfer-eth",
		s.aliceSwapdPortFlag(),
		fmt.Sprintf("--%s=%s", flagTo, token.Address),
		fmt.Sprintf("--%s=%s", flagAmount, zeroETH),
	}
	err = cliApp().RunContext(context.Background(), args)
	require.ErrorContains(s.T(), err, "gas limit is required when transferring to a contract")

	// 2nd attempt with the gas limit set will succeed
	args = append(args, fmt.Sprintf("--%s=%d", flagGasLimit, 2*params.TxGas))
	err = cliApp().RunContext(context.Background(), args)
	require.NoError(s.T(), err)

	// our test token contract mints you 100 standard token units when sending
	// it a zero value transaction.
	endTokenBal, err := ec.ERC20Balance(ctx, token.Address)
	require.NoError(s.T(), err)
	require.Greater(s.T(), endTokenBal.AsStd().Cmp(startTokenBal.AsStd()), 0)
}

func (s *swapCLITestSuite) Test_runRunXMRTransfer() {
	bobAddr := s.bobConf.MoneroClient.PrimaryAddress()
	aliceAddr := s.aliceConf.MoneroClient.PrimaryAddress()
	amount := coins.MoneroToPiconero(coins.StrToDecimal("0.123"))
	aliceMC := s.aliceConf.MoneroClient
	bobMC := s.bobConf.MoneroClient

	monero.MineMinXMRBalance(s.T(), bobMC, amount)

	s.T().Logf("Bob is transferring %s XMR to Alice", amount.AsMoneroString())
	args := []string{
		"swapcli",
		"transfer-xmr",
		s.bobSwapdPortFlag(),
		fmt.Sprintf("--%s=%s", flagTo, aliceAddr),
		fmt.Sprintf("--%s=0.123", flagAmount),
	}
	err := cliApp().RunContext(context.Background(), args)
	require.NoError(s.T(), err)

	// Transfer only waits for 1 confirmation. Wait the remaining 9 blocks
	// (should be around 9 seconds) for the balance to fully unlock.
	var aliceBalResp *wallet.GetBalanceResponse
	for i := 0; i < 12; i++ {
		aliceBalResp, err = aliceMC.GetBalance(0)
		require.NoError(s.T(), err)
		if aliceBalResp.BlocksToUnlock == 0 {
			break
		}
		s.T().Logf("Waiting for Alice's balance to unlock, %d blocks remaining", aliceBalResp.BlocksToUnlock)
		time.Sleep(1 * time.Second)
	}

	require.Zero(s.T(), aliceBalResp.BlocksToUnlock)
	require.Equal(s.T(), aliceBalResp.Balance, aliceBalResp.UnlockedBalance)
	require.GreaterOrEqual(s.T(), coins.NewPiconeroAmount(aliceBalResp.Balance).Cmp(amount), 0)

	s.T().Log("Alice is sweeping her entire XMR balance back to Bob")
	args = []string{
		"swapcli",
		"sweep-xmr",
		s.aliceSwapdPortFlag(),
		fmt.Sprintf("--%s=%s", flagTo, bobAddr),
	}
	err = cliApp().RunContext(context.Background(), args)
	require.NoError(s.T(), err)

	aliaceBal, err := aliceMC.GetBalance(0)
	require.NoError(s.T(), err)
	require.Zero(s.T(), aliaceBal.Balance)
}
