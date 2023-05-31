package main

import (
	"context"
	"strconv"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

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
	conf       *daemon.SwapdConfig
	mockTether *coins.ERC20TokenInfo
	mockDAI    *coins.ERC20TokenInfo
}

func TestRunSwapcliWithDaemonTests(t *testing.T) {
	s := new(swapCLITestSuite)

	s.conf = daemon.CreateTestConf(t, tests.GetMakerTestKey(t))
	t.Setenv("SWAPD_PORT", strconv.Itoa(int(s.conf.RPCPort)))
	daemon.LaunchDaemons(t, 10*time.Minute, s.conf)
	ec := s.conf.EthereumClient.Raw()
	pk := s.conf.EthereumClient.PrivateKey()
	s.mockTether = contracts.GetMockTether(t, ec, pk)
	s.mockDAI = contracts.GetMockDAI(t, ec, pk)
	suite.Run(t, s)
}

func (s *swapCLITestSuite) rpcEndpoint() *rpcclient.Client {
	return rpcclient.NewClient(context.Background(), s.conf.RPCPort)
}

func (s *swapCLITestSuite) mockDaiAddr() ethcommon.Address {
	return s.mockDAI.Address
}

func (s *swapCLITestSuite) mockTetherAddr() ethcommon.Address {
	return s.mockTether.Address
}
