package main

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/require"
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
		fmt.Sprintf("--%s=%s", flagToken, s.mockDAI.Address.Hex()),
	}
	err := cliApp().RunContext(context.Background(), args)
	require.NoError(s.T(), err)
}
