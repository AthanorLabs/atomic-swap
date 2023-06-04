package main

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/daemon"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/tests"
)

// swapCLITestSuite provides a suite that unit tests can associate themselves
// if they need access to a swapd daemon and some preconfigured ERC20 tokens.
type swapCLITestSuite struct {
	suite.Suite
	aliceConf  *daemon.SwapdConfig
	bobConf    *daemon.SwapdConfig
	mockTether *coins.ERC20TokenInfo
	mockDAI    *coins.ERC20TokenInfo
}

func TestRunSwapcliWithDaemonTests(t *testing.T) {
	s := new(swapCLITestSuite)

	s.aliceConf = daemon.CreateTestConf(t, tests.GetMakerTestKey(t))

	bobEthKey, err := crypto.GenerateKey() // Bob has no ETH
	require.NoError(t, err)
	s.bobConf = daemon.CreateTestConf(t, bobEthKey)

	// by default you'll get Alice's RPC endpoint, specifying flag has precedence
	t.Setenv("SWAPD_PORT", strconv.Itoa(int(s.aliceConf.RPCPort)))
	daemon.LaunchDaemons(t, 10*time.Minute, s.aliceConf, s.bobConf)

	ec := s.aliceConf.EthereumClient.Raw()
	pk := s.aliceConf.EthereumClient.PrivateKey()
	s.mockTether = contracts.GetMockTether(t, ec, pk)
	s.mockDAI = contracts.GetMockDAI(t, ec, pk)

	cliutil.SetLogLevels("debug") // turn on logging after daemons are started
	suite.Run(t, s)
}

func (s *swapCLITestSuite) aliceRPCPort() uint16 {
	return s.aliceConf.RPCPort
}

func (s *swapCLITestSuite) bobRPCPort() uint16 {
	return s.bobConf.RPCPort
}

func (s *swapCLITestSuite) aliceSwapdPortFlag() string {
	return fmt.Sprintf("--%s=%d", flagSwapdPort, s.aliceRPCPort())
}

func (s *swapCLITestSuite) bobSwapdPortFlag() string {
	return fmt.Sprintf("--%s=%d", flagSwapdPort, s.bobRPCPort())
}

func (s *swapCLITestSuite) aliceClient() *rpcclient.Client {
	return rpcclient.NewClient(context.Background(), s.aliceRPCPort())
}

func (s *swapCLITestSuite) bobClient() *rpcclient.Client {
	return rpcclient.NewClient(context.Background(), s.bobRPCPort())
}

func (s *swapCLITestSuite) mockDaiAddr() ethcommon.Address {
	return s.mockDAI.Address
}

func (s *swapCLITestSuite) mockTetherAddr() ethcommon.Address {
	return s.mockTether.Address
}
