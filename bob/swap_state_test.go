package bob

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/monero/crypto"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/swap-contract"
	"github.com/noot/atomic-swap/types"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"
)

var (
	_            = logging.SetLogLevel("bob", "debug")
	testWallet   = "test-wallet"
	desiredAmout = common.NewEtherAmount(33)
)

type mockNet struct {
	msg net.Message
}

func (n *mockNet) SendSwapMessage(msg net.Message) error {
	n.msg = msg
	return nil
}

var defaultTimeoutDuration = big.NewInt(60 * 60 * 24) // 1 day = 60s * 60min * 24hr

func newTestBob(t *testing.T) (*bob, *swapState) {
	cfg := &Config{
		Ctx:                  context.Background(),
		Basepath:             "/tmp/bob",
		MoneroWalletEndpoint: common.DefaultBobMoneroEndpoint,
		MoneroDaemonEndpoint: common.DefaultMoneroDaemonEndpoint,
		WalletFile:           testWallet,
		WalletPassword:       "",
		EthereumEndpoint:     common.DefaultEthEndpoint,
		EthereumPrivateKey:   common.DefaultPrivKeyBob,
		Environment:          common.Development,
		ChainID:              common.MainnetConfig.EthereumChainID,
	}

	bob, err := NewBob(cfg)
	require.NoError(t, err)

	bobAddr, err := bob.client.GetAddress(0)
	require.NoError(t, err)

	_ = bob.daemonClient.GenerateBlocks(bobAddr.Address, 121)

	swapState, err := newSwapState(bob, types.Hash{}, common.MoneroAmount(33), desiredAmout)
	require.NoError(t, err)
	return bob, swapState
}

func newTestAliceSendKeySMessage(t *testing.T) (*net.SendKeysMessage, *mcrypto.PrivateKeyPair) {
	alicePrivKeys, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	sig, err := alicePrivKeys.SpendKey().Sign(alicePrivKeys.SpendKey().Public().Bytes())
	require.NoError(t, err)

	msg := &net.SendKeysMessage{
		PublicSpendKey:  alicePrivKeys.SpendKey().Public().Hex(),
		PublicViewKey:   alicePrivKeys.ViewKey().Public().Hex(),
		PrivateKeyProof: sig.Hex(),
	}

	return msg, alicePrivKeys
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

	var claimKey [32]byte
	copy(claimKey[:], common.Reverse(swapState.privkeys.SpendKey().Public().Bytes()))
	swapState.contractAddr, _, swapState.contract, err = swap.DeploySwap(swapState.txOpts, conn,
		claimKey, [32]byte{}, bob.ethAddress, defaultTimeoutDuration)
	require.NoError(t, err)

	_, err = swapState.contract.SetReady(swapState.txOpts)
	require.NoError(t, err)

	txHash, err := swapState.claimFunds()
	require.NoError(t, err)
	require.NotEqual(t, "", txHash)
}

func TestSwapState_handleSendKeysMessage(t *testing.T) {
	_, s := newTestBob(t)

	msg := &net.SendKeysMessage{}
	err := s.handleSendKeysMessage(msg)
	require.Equal(t, errMissingKeys, err)

	msg, alicePrivKeys := newTestAliceSendKeySMessage(t)
	alicePubKeys := alicePrivKeys.PublicKeyPair()

	err = s.handleSendKeysMessage(msg)
	require.NoError(t, err)
	require.Equal(t, &net.NotifyContractDeployed{}, s.nextExpectedMessage)
	require.Equal(t, alicePubKeys.SpendKey().Hex(), s.alicePublicKeys.SpendKey().Hex())
	require.Equal(t, alicePubKeys.ViewKey().Hex(), s.alicePublicKeys.ViewKey().Hex())
}

func deploySwap(t *testing.T, bob *bob, swapState *swapState, refundKey [32]byte, amount *big.Int,
	timeout time.Duration) (ethcommon.Address, *swap.Swap) {
	conn, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	tm := big.NewInt(int64(timeout.Seconds()))

	var claimKey [32]byte
	copy(claimKey[:], common.Reverse(swapState.privkeys.SpendKey().Public().Bytes()))

	swapState.txOpts.Value = amount
	defer func() {
		swapState.txOpts.Value = nil
	}()

	addr, _, contract, err := swap.DeploySwap(swapState.txOpts, conn, claimKey, refundKey, bob.ethAddress, tm)
	require.NoError(t, err)
	return addr, contract
}

func TestSwapState_HandleProtocolMessage_NotifyContractDeployed_ok(t *testing.T) {
	bob, s := newTestBob(t)
	defer s.cancel()
	s.nextExpectedMessage = &net.NotifyContractDeployed{}
	_, _, err := s.generateKeys()
	require.NoError(t, err)

	aliceKeys, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	s.setAlicePublicKeys(aliceKeys.PublicKeyPair())

	msg := &net.NotifyContractDeployed{}
	resp, done, err := s.HandleProtocolMessage(msg)
	require.Equal(t, errMissingAddress, err)
	require.Nil(t, resp)
	require.True(t, done)

	duration, err := time.ParseDuration("2s")
	require.NoError(t, err)
	addr, _ := deploySwap(t, bob, s, [32]byte{}, desiredAmout.BigInt(), duration)

	s.providesAmount = 1
	msg = &net.NotifyContractDeployed{
		Address: addr.String(),
	}

	resp, done, err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, net.NotifyXMRLockType, resp.Type())
	require.False(t, done)
	require.NotNil(t, s.contract)
	require.Equal(t, addr, s.contractAddr)
	require.Equal(t, duration, s.t1.Sub(s.t0))
	require.Equal(t, &net.NotifyReady{}, s.nextExpectedMessage)
}

func TestSwapState_HandleProtocolMessage_NotifyContractDeployed_timeout(t *testing.T) {
	bob, s := newTestBob(t)
	defer s.cancel()
	s.net = new(mockNet)
	s.nextExpectedMessage = &net.NotifyContractDeployed{}
	_, _, err := s.generateKeys()
	require.NoError(t, err)

	aliceKeys, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	s.setAlicePublicKeys(aliceKeys.PublicKeyPair())

	msg := &net.NotifyContractDeployed{}
	resp, done, err := s.HandleProtocolMessage(msg)
	require.Equal(t, errMissingAddress, err)
	require.Nil(t, resp)
	require.True(t, done)

	duration, err := time.ParseDuration("15s")
	require.NoError(t, err)
	addr, _ := deploySwap(t, bob, s, [32]byte{}, desiredAmout.BigInt(), duration)

	msg = &net.NotifyContractDeployed{
		Address: addr.String(),
	}

	resp, done, err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, net.NotifyXMRLockType, resp.Type())
	require.False(t, done)
	require.NotNil(t, s.contract)
	require.Equal(t, addr, s.contractAddr)
	require.Equal(t, duration, s.t1.Sub(s.t0))
	require.Equal(t, &net.NotifyReady{}, s.nextExpectedMessage)

	time.Sleep(duration * 3)
	require.NotNil(t, s.net.(*mockNet).msg)
}

func TestSwapState_HandleProtocolMessage_NotifyReady(t *testing.T) {
	bob, s := newTestBob(t)

	s.nextExpectedMessage = &net.NotifyReady{}
	_, _, err := s.generateKeys()
	require.NoError(t, err)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)
	_, s.contract = deploySwap(t, bob, s, [32]byte{}, desiredAmout.BigInt(), duration)

	_, err = s.contract.SetReady(s.txOpts)
	require.NoError(t, err)

	msg := &net.NotifyReady{}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.True(t, done)
	require.NotNil(t, resp)
	require.Equal(t, net.NotifyClaimedType, resp.Type())
}

func TestSwapState_handleRefund(t *testing.T) {
	bob, s := newTestBob(t)

	_, _, err := s.generateKeys()
	require.NoError(t, err)

	aliceKeys, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	s.setAlicePublicKeys(aliceKeys.PublicKeyPair())

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	var refundKey [32]byte
	copy(refundKey[:], common.Reverse(aliceKeys.SpendKey().Public().Bytes()))
	_, s.contract = deploySwap(t, bob, s, refundKey, desiredAmout.BigInt(), duration)

	// lock XMR
	addrAB, err := s.lockFunds(s.providesAmount)
	require.NoError(t, err)

	// call refund w/ Alice's spend key
	secret := aliceKeys.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	tx, err := s.contract.Refund(s.txOpts, sc)
	require.NoError(t, err)

	addr, err := s.handleRefund(tx.Hash().String())
	require.NoError(t, err)
	require.Equal(t, addrAB, addr)
}

func TestSwapState_HandleProtocolMessage_NotifyRefund(t *testing.T) {
	bob, s := newTestBob(t)

	_, _, err := s.generateKeys()
	require.NoError(t, err)

	aliceKeys, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	s.setAlicePublicKeys(aliceKeys.PublicKeyPair())

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	var refundKey [32]byte
	copy(refundKey[:], common.Reverse(aliceKeys.SpendKey().Public().Bytes()))
	_, s.contract = deploySwap(t, bob, s, refundKey, desiredAmout.BigInt(), duration)

	// lock XMR
	_, err = s.lockFunds(s.providesAmount)
	require.NoError(t, err)

	// call refund w/ Alice's spend key
	secret := aliceKeys.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	tx, err := s.contract.Refund(s.txOpts, sc)
	require.NoError(t, err)

	msg := &net.NotifyRefund{
		TxHash: tx.Hash().String(),
	}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.True(t, done)
	require.Nil(t, resp)
}

// test that if the protocol exits early, and Alice refunds, Bob can reclaim his monero
func TestSwapState_ProtocolExited_Reclaim(t *testing.T) {
	bob, s := newTestBob(t)

	_, _, err := s.generateKeys()
	require.NoError(t, err)

	aliceKeys, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	s.setAlicePublicKeys(aliceKeys.PublicKeyPair())

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	var refundKey [32]byte
	copy(refundKey[:], common.Reverse(aliceKeys.SpendKey().Public().Bytes()))
	s.contractAddr, s.contract = deploySwap(t, bob, s, refundKey, desiredAmout.BigInt(), duration)

	// lock XMR
	_, err = s.lockFunds(s.providesAmount)
	require.NoError(t, err)

	// call refund w/ Alice's spend key
	secret := aliceKeys.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	tx, err := s.contract.Refund(s.txOpts, sc)
	require.NoError(t, err)

	receipt, err := bob.ethClient.TransactionReceipt(s.ctx, tx.Hash())
	require.NoError(t, err)
	require.Equal(t, 1, len(receipt.Logs))
	require.Equal(t, 1, len(receipt.Logs[0].Topics))
	require.Equal(t, refundedTopic, receipt.Logs[0].Topics[0])

	s.nextExpectedMessage = &net.NotifyReady{}
	err = s.ProtocolExited()
	require.NoError(t, err)

	balance, err := bob.client.GetBalance(0)
	require.NoError(t, err)
	require.Equal(t, s.providesAmount.Uint64(), uint64(balance.Balance))
}
