package main

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/athanorlabs/atomic-swap/daemon"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/tests"
)

// swapCLITestSuite provides a suite that unit tests can associate themselves
// if they need access to a swapd daemon and some preconfigured ERC20 tokens.
type swapCLITestSuite struct {
	suite.Suite
	conf       *daemon.SwapdConfig
	mockTokens map[string]ethcommon.Address
}

func TestRunIntegrationTests(t *testing.T) {
	s := new(swapCLITestSuite)

	s.conf = daemon.CreateTestConf(t, tests.GetMakerTestKey(t))
	t.Setenv("SWAPD_PORT", strconv.Itoa(int(s.conf.RPCPort)))
	daemon.LaunchDaemons(t, 10*time.Minute, s.conf)
	s.mockTokens = daemon.GetMockTokens(t, s.conf.EthereumClient)
	suite.Run(t, s)
}

func (s *swapCLITestSuite) rpcEndpoint() *rpcclient.Client {
	return rpcclient.NewClient(context.Background(), fmt.Sprintf("http://127.0.0.1:%d", s.conf.RPCPort))
}

func (s *swapCLITestSuite) mockDaiAddr() ethcommon.Address {
	return s.mockTokens[daemon.MockDAI]
}

func (s *swapCLITestSuite) mockTetherAddr() ethcommon.Address {
	return s.mockTokens[daemon.MockTether]
}
