// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"context"
	"testing"
	"time"

	"github.com/ChainSafe/chaindb"
	"github.com/MarinX/monerorpc/wallet"
	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2ptest "github.com/libp2p/go-libp2p/core/test"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/db"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"
)

//
// This file only contains mock definitions used by other test files
//

type mockNet struct {
	peerID peer.ID
}

func (*mockNet) Addresses() []ma.Multiaddr {
	panic("not implemented")
}

func (m *mockNet) PeerID() peer.ID {
	if m.peerID == "" {
		var err error
		m.peerID, err = libp2ptest.RandPeerID()
		if err != nil {
			panic(err)
		}
	}
	return m.peerID
}

func (*mockNet) ConnectedPeers() []string {
	panic("not implemented")
}

func (*mockNet) Discover(_ string, _ time.Duration) ([]peer.ID, error) {
	return nil, nil
}

func (*mockNet) Query(_ peer.ID) (*message.QueryResponse, error) {
	return &message.QueryResponse{Offers: []*types.Offer{{ID: testSwapID}}}, nil
}

func (*mockNet) Initiate(_ peer.AddrInfo, _ common.Message, _ common.SwapStateNet) error {
	return nil
}

func (*mockNet) CloseProtocolStream(_ types.Hash) {
	panic("not implemented")
}

func mockSwapManager(t *testing.T) swap.Manager {
	db, err := db.NewDatabase(&chaindb.Config{
		DataDir:  t.TempDir(),
		InMemory: true,
	})
	require.NoError(t, err)

	sm, err := swap.NewManager(db)
	require.NoError(t, err)

	one := apd.New(1, 0)

	sm.AddSwap(swap.NewInfo(
		testPeerID,
		testSwapID,
		coins.ProvidesETH,
		one,
		one,
		coins.ToExchangeRate(one),
		types.EthAssetETH,
		types.CompletedSuccess,
		1,
	))

	sm.PushNewStatus(testSwapID, types.CompletedSuccess)

	return sm
}

type mockXMRTaker struct{}

func (*mockXMRTaker) Provides() coins.ProvidesCoin {
	panic("not implemented")
}

func (*mockXMRTaker) GetOngoingSwapState(_ types.Hash) common.SwapState {
	return new(mockSwapState)
}

func (*mockXMRTaker) InitiateProtocol(_ peer.ID, _ *apd.Decimal, _ *types.Offer) (common.SwapState, error) {
	return new(mockSwapState), nil
}

func (*mockXMRTaker) Refund(_ types.Hash) (ethcommon.Hash, error) {
	panic("not implemented")
}

func (*mockXMRTaker) SetSwapTimeout(_ time.Duration) {
	panic("not implemented")
}

func (*mockXMRTaker) ExternalSender(_ types.Hash) (*txsender.ExternalSender, error) {
	panic("not implemented")
}

type mockXMRMaker struct{}

func (m *mockXMRMaker) Provides() coins.ProvidesCoin {
	panic("not implemented")
}

func (m *mockXMRMaker) GetOngoingSwapState(_ types.Hash) common.SwapState {
	panic("not implemented")
}

func (*mockXMRMaker) MakeOffer(_ *types.Offer, _ bool) (*types.OfferExtra, error) {
	offerExtra := types.NewOfferExtra(false)
	return offerExtra, nil
}

func (*mockXMRMaker) GetOffers() []*types.Offer {
	panic("not implemented")
}

func (*mockXMRMaker) ClearOffers(_ []types.Hash) error {
	panic("not implemented")
}

func (*mockXMRMaker) GetMoneroBalance() (*mcrypto.Address, *wallet.GetBalanceResponse, error) {
	panic("not implemented")
}

type mockSwapState struct{}

func (*mockSwapState) NotifyStreamClosed() {}

func (*mockSwapState) HandleProtocolMessage(_ common.Message) error {
	return nil
}

func (*mockSwapState) Exit() error {
	return nil
}

func (*mockSwapState) SendKeysMessage() common.Message {
	return &message.SendKeysMessage{}
}

func (*mockSwapState) OfferID() types.Hash {
	return testSwapID
}

type mockProtocolBackend struct {
	sm swap.Manager
}

func newMockProtocolBackend(t *testing.T) *mockProtocolBackend {
	return &mockProtocolBackend{
		sm: mockSwapManager(t),
	}
}

func (*mockProtocolBackend) Ctx() context.Context {
	return context.Background()
}

func (*mockProtocolBackend) Env() common.Environment {
	return common.Development
}

func (*mockProtocolBackend) SetSwapTimeout(_ time.Duration) {
	panic("not implemented")
}

func (*mockProtocolBackend) SwapTimeout() time.Duration {
	panic("not implemented")
}

func (b *mockProtocolBackend) SwapManager() swap.Manager {
	return b.sm
}

func (*mockProtocolBackend) SetXMRDepositAddress(*mcrypto.Address, types.Hash) {
	panic("not implemented")
}

func (*mockProtocolBackend) ClearXMRDepositAddress(types.Hash) {
	panic("not implemented")
}

func (*mockProtocolBackend) ETHClient() extethclient.EthClient {
	panic("not implemented")
}

func (*mockProtocolBackend) SwapCreatorAddr() ethcommon.Address {
	return ethcommon.Address{}
}

func (*mockProtocolBackend) TransferXMR(_ *mcrypto.Address, _ *coins.PiconeroAmount) (string, error) {
	panic("not implemented")
}

func (*mockProtocolBackend) SweepXMR(_ *mcrypto.Address) ([]string, error) {
	panic("not implemented")
}

func (*mockProtocolBackend) TransferETH(_ ethcommon.Address, _ *coins.WeiAmount, _ *uint64) (*ethtypes.Receipt, error) {
	panic("not implemented")
}

func (*mockProtocolBackend) SweepETH(_ ethcommon.Address) (*ethtypes.Receipt, error) {
	panic("not implemented")
}
